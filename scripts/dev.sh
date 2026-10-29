#!/usr/bin/env bash
# make dev: the dependencies (Postgres, Redis, RabbitMQ, LocalStack S3+KMS) in Docker,
# the backend and the frontend natively, with seeded test accounts and availabilities.
#
# The dependency stack runs under its own compose project (leviosa-dev) with moved host
# ports, so it can run next to `make local-dev` / `make local-start`. It stays up between
# runs; it is recreated from scratch whenever LocalStack is not running, because
# LocalStack keeps no state and Postgres data encrypted under a lost KMS key is
# unreadable. `make dev-down` stops it and wipes its data.
#
# Ctrl-C stops the backend and the frontend; the dependencies keep running.
set -euo pipefail

ROOT=$(cd "$(dirname "$0")/.." && pwd)
BACKEND="$ROOT/backend"
FRONTEND="$ROOT/frontend"

export DEV_POSTGRES_PORT=${DEV_POSTGRES_PORT:-45432}
export DEV_REDIS_PORT=${DEV_REDIS_PORT:-46379}
export DEV_RABBITMQ_PORT=${DEV_RABBITMQ_PORT:-45672}
export DEV_RABBITMQ_UI_PORT=${DEV_RABBITMQ_UI_PORT:-45673}
export DEV_LOCALSTACK_PORT=${DEV_LOCALSTACK_PORT:-44566}
BACKEND_PORT=${DEV_BACKEND_PORT:-3500}
# 5173 is often taken by another checkout's `pnpm dev`.
FRONTEND_PORT=${DEV_FRONTEND_PORT:-5180}
SEED_DATA_PATH=${SEED_DATA_PATH:-cmd/seed/seed_data.example.json}

compose() { docker compose -p leviosa-dev -f "$BACKEND/compose.yaml" "$@"; }

if [ "${1:-}" = down ]; then
	compose down -v
	exit 0
fi

for port in "$BACKEND_PORT" "$FRONTEND_PORT"; do
	if (exec 3<>"/dev/tcp/127.0.0.1/$port") 2>/dev/null || (exec 3<>"/dev/tcp/::1/$port") 2>/dev/null; then
		echo "Port $port is already in use (another make dev?). Set DEV_BACKEND_PORT / DEV_FRONTEND_PORT to move it."
		exit 1
	fi
done

# --- Env files (same first-run copies as make local-dev) ------------------------------
if [ ! -f "$BACKEND/development.env" ]; then
	echo "No backend/development.env: copying backend/example.env (integrations off)."
	cp "$BACKEND/example.env" "$BACKEND/development.env"
fi
if [ ! -f "$FRONTEND/.env" ]; then
	echo "No frontend/.env: copying frontend/.env.example."
	cp "$FRONTEND/.env.example" "$FRONTEND/.env"
fi

# example.env gives every key a working local default (encx KEK alias, pepper, admin);
# development.env (yours, or rendered by make env) overrides it. Both hold values with
# spaces (ADMIN_ADDRESS1), so they can't be sourced.
load_env() {
	while IFS= read -r line || [ -n "$line" ]; do
		case "$line" in "" | \#*) continue ;; esac
		line=${line%$'\r'}
		export "$line"
	done <"$1"
}
load_env "$BACKEND/example.env"
load_env "$BACKEND/development.env"

export ENVIRONMENT=development
export SERVER_PORT=$BACKEND_PORT
export DATABASE_URL="postgresql://user:password@127.0.0.1:$DEV_POSTGRES_PORT/leviosa_dev"
export REDIS_ADDR="127.0.0.1:$DEV_REDIS_PORT" REDIS_PASSWORD=secret REDIS_DB=0
export RABBITMQ_URL="amqp://guest:guest@127.0.0.1:$DEV_RABBITMQ_PORT/%2F"
export AWS_ENDPOINT_URL="http://127.0.0.1:$DEV_LOCALSTACK_PORT" S3_ENDPOINT="http://127.0.0.1:$DEV_LOCALSTACK_PORT"
export S3_REGION=us-east-1 S3_BUCKET_NAME=leviosa S3_ACCESS_KEY_ID=test S3_SECRET_ACCESS_KEY=test
export FRONTEND_ORIGIN="http://localhost:$FRONTEND_PORT"
export SESSION_COOKIE_NAME=leviosa_session CLIENT_IP_HEADER=x-client-ip
export BOOKING_TOKEN_SECRET=${BOOKING_TOKEN_SECRET:-local-dev-booking-token-secret}

# --- Dependencies ----------------------------------------------------------------------
if [ -z "$(compose ps -q --status running localstack 2>/dev/null)" ]; then
	echo "Starting a fresh dependency stack (leviosa-dev)..."
	compose down -v >/dev/null 2>&1 || true
fi
compose up -d --wait postgres redis rabbitmq localstack
compose run --rm localstack-init >/dev/null

# --- Backend: build, migrate, seed -----------------------------------------------------
cd "$BACKEND"
mkdir -p tmp
echo "Building the backend..."
go build -o tmp/leviosa-dev ./cmd/app
tmp/leviosa-dev migrate
echo "Seeding accounts and the practice..."
if ! seed_log=$(SEED_DATA_PATH="$SEED_DATA_PATH" go run ./cmd/seed 2>&1); then
	echo "$seed_log"
	exit 1
fi
echo "$seed_log" | grep -E 'Seeded|completed' || true

[ -d "$FRONTEND/node_modules" ] || (cd "$FRONTEND" && pnpm install)

BACKEND_PID=
cleanup() { [ -n "$BACKEND_PID" ] && kill "$BACKEND_PID" 2>/dev/null || true; }
trap cleanup EXIT INT TERM

tmp/leviosa-dev serve > >(sed -u 's/^/[api] /') 2>&1 &
BACKEND_PID=$!

API="http://127.0.0.1:$BACKEND_PORT"
for _ in $(seq 1 60); do
	kill -0 "$BACKEND_PID" 2>/dev/null || break
	curl -sf "$API/health" >/dev/null && break
	sleep 1
done
kill -0 "$BACKEND_PID" 2>/dev/null || { echo "The backend exited."; exit 1; }
curl -sf "$API/health" >/dev/null || { echo "The backend did not start."; exit 1; }

# --- Availabilities (only on an empty agenda), created as the partner through the API --
psql_dev() { compose exec -T postgres psql -U user -d leviosa_dev -Atc "$1"; }
if [ "$(psql_dev 'SELECT COUNT(*) FROM booking.availabilities')" = 0 ]; then
	partner_email=$(sed -n '/"partner"/,/}/s/.*"email": *"\([^"]*\)".*/\1/p' "$SEED_DATA_PATH" | head -1)
	partner_password=$(sed -n '/"partner"/,/}/s/.*"password": *"\([^"]*\)".*/\1/p' "$SEED_DATA_PATH" | head -1)
	jar=$(mktemp)
	curl -sf -c "$jar" -o /dev/null -H 'Content-Type: application/json' \
		-d "{\"email\":\"$partner_email\",\"password\":\"$partner_password\"}" "$API/auth/login"
	partner_id=$(psql_dev 'SELECT user_id FROM auth.partners LIMIT 1')
	room_id=$(psql_dev 'SELECT id FROM booking.rooms LIMIT 1')
	paris() { TZ=Europe/Paris date -d "$1" --iso-8601=seconds; }
	tue=$(date -d 'next tuesday' +%F)
	fri=$(date -d 'next friday' +%F)
	# Weekly Tue/Thu/Sat 09:00-13:00, and a one-off Friday 14:00-18:00 (Paris time).
	curl -sf -b "$jar" -o /dev/null -H 'Content-Type: application/json' "$API/availabilities/recurring" \
		-d "{\"partner_id\":\"$partner_id\",\"room_id\":\"$room_id\",\"start_time\":\"$(paris "$tue 09:00")\",\"end_time\":\"$(paris "$tue 13:00")\",\"max_capacity\":1,\"pattern\":{\"type\":\"weekly\",\"interval\":1,\"days_of_week\":[2,4,6]}}" &&
		curl -sf -b "$jar" -o /dev/null -H 'Content-Type: application/json' "$API/availabilities" \
			-d "{\"partner_id\":\"$partner_id\",\"room_id\":\"$room_id\",\"start_time\":\"$(paris "$fri 14:00")\",\"end_time\":\"$(paris "$fri 18:00")\",\"max_capacity\":1}" &&
		echo "Seeded availabilities: weekly Tue/Thu/Sat 09:00-13:00 from $tue, one-off $fri 14:00-18:00." ||
		echo "Could not seed the availabilities (create them from the partner agenda)."
	rm -f "$jar"
fi

cat <<EOF

Leviosa is starting at http://localhost:$FRONTEND_PORT (backend on $API)
  Practice page: /  and  /en
  Log in at /auth with the accounts of $SEED_DATA_PATH (passwords there):
    admins (admins[]), the practitioner (practice.partner), test clients (clients[])
  Payment needs Stripe test keys (make env) and make stripe-listen; emails need SMTP.
  Ctrl-C stops the apps; make dev-down stops the dependencies and wipes their data.

EOF

cd "$FRONTEND"
API_URL="$API" pnpm dev --port "$FRONTEND_PORT" --strictPort
