imageName:= "henga/backend-event_reservation_app"

.PHONY: docker
docker:
	@docker image rm $(imageName):latest -f
	@echo "Removed old image $(imageName)"
	@docker build -t $(imageName) .
	@echo "Image built !\n"
	@echo "Running image..."
	@docker run -p 3500:3500 $(imageName):latest 

.PHONY: run
run: 
	@echo "Running the application..."
	@# @echo "Removing, just for testing, the previous db.sqlite instance..."
	@# @rm db.sqlite
	@go run ./cmd/event-reservation-app/*.go -port 3500

.PHONY: build
build:
	@echo "Building the application..."
	@go build -o ./bin/app ./cmd/event-reservation-app/*.go
	@echo "Application built..."

.PHONY: test
test:
	@echo "Running all tests...\n"
	@go test -v -race ./tests/... -run TestGETEvents

.PHONY: ta
ta: # to test all
	@echo "Running all tests...\n"
	@go test -v ./tests/...

.PHONY: clean
clean:
	@echo "Removing all binaries..."
	@rm -rf ./bin/*
	@echo "All binaries removed !"

# NOTE: goose related commands
ifneq (,$(wildcard ./.env))
    include .env
    export
endif

migrationPath:="./internal/sqlite/migrations"
goosePrefix:=GOOSE_DRIVER=$(GOOSE_DRIVER) GOOSE_DBSTRING=$(GOOSE_DBSTRING) goose -dir=$(migrationPath) 
name:=""

.PHONY: db-status
db-status:
	@$(goosePrefix) status

.PHONY: create
create:
	goose -dir=$(migrationPath) create $(name) sql

.PHONY: up
up:
	@$(goosePrefix) up

.PHONY: down
down:
	@$(goosePrefix) down

.PHONY: reset
reset:
	@$(goosePrefix) reset
