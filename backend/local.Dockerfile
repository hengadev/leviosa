# Dev-mode image: live reload via air. The full source tree is bind-mounted over this
# at runtime (see compose.yaml), so air watches host file changes and rebuilds/restarts
# the binary automatically — no manual `docker compose up --build` per change.
FROM golang:1.24.2-alpine

RUN apk add --no-cache git gcc musl-dev

WORKDIR /app

# air v1.63+ requires Go >= 1.25; v1.62.0 is the newest that builds on this image.
RUN go install github.com/air-verse/air@v1.62.0

COPY go.mod go.sum ./
RUN go mod download

COPY . .

EXPOSE 3500

CMD ["air", "-c", ".air.toml"]
