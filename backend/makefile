imageName:= "henga/backend-event_reservation_app"

docker:
	@docker image rm $(imageName):latest -f
	@echo "Removed old image $(imageName)"
	@docker build -t $(imageName) .
	@echo "Image built !\n"
	@echo "Running image..."
	@docker run -p 3500:3500 $(imageName):latest 

run: 
	@echo "Running the application..."
	@# @echo "Removing, just for testing, the previous db.sqlite instance..."
	@# @rm db.sqlite
	@go run ./cmd/event-reservation-app/*.go -port 3500

build:
	@echo "Building the application..."
	@go build -o ./bin/app ./cmd/event-reservation-app/*.go
	@echo "Application built..."

test:
	@echo "Running all tests...\n"
	@go test -v -race ./tests/... -run TestGETEvents

ta: # to test all
	@echo "Running all tests...\n"
	@go test -v ./tests/...

clean:
	@echo "Removing all binaries..."
	@rm -rf ./bin/*
	@echo "All binaries removed !"

.PHONY: run build test ta clean 

# NOTE: goose related commands
ifneq (,$(wildcard ./.env))
    include .env
    export
endif

migrationPath:="./internal/sqlite/migrations"
goosePrefix:=GOOSE_DRIVER=$(GOOSE_DRIVER) GOOSE_DBSTRING=$(GOOSE_DBSTRING) goose -dir=$(migrationPath) 
name:=""

db-status:
	@$(goosePrefix) status

create:
	goose -dir=$(migrationPath) create $(name) sql

up:
	@$(goosePrefix) up

down:
	@$(goosePrefix) down

reset:
	@$(goosePrefix) reset

.PHONY: db-status create up down reset
