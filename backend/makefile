# env vars
ifneq (,$(wildcard ./development.env))
    include development.env
    export
endif

# dockerfile vars
imageName:= "henga/event_reservation_app"
redisName:= "session-manager"
PORT=3500

# goose vars
goosePrefix:=GOOSE_DRIVER=$(GOOSE_DRIVER) GOOSE_DBSTRING=$(GOOSE_DBSTRING) goose -dir=$(MIGRATION_PATH) 
name:=""

.PHONY: help all run docker redis build test tesall clean db-status create up down reset

# TODO: find how to use that command with this link : https://dev.to/flpslv/using-makefiles-to-build-and-publish-docker-containers-7c8
help:
	@echo "==========================="
	@echo "Makefile backend arguments:"
	@echo "==========================="
	@echo ""
	@echo "imageName"
	@echo "redisName"
	@echo ""
	@echo "==========================="
	@echo "Makefile backend commands:"
	@echo "==========================="
	@echo ""
	@echo "docker: dockerize the application"
	@echo "run: run the application with the go run command"
	@echo "redis: configure the redis container and run it"

.DEFAULT_GOAL := all

all: docker redis

run: redis
	@echo "Running the application..."
	@if [ -f "$(DATABASE_FILENAME).db" ]; then \
		rm "$(DATABASE_FILENAME).db"; \
	fi
	@APP_ENV=staging go run  ./cmd/event-reservation-app/*.go


docker:
	@docker image rm $(imageName):backend -f
	@echo "Removed old image $(imageName)"
	@docker build -t $(imageName):backend .
	@echo "Image built !\n"
	@echo "Running image..."
	@docker run -e APP_ENV='development' -e PORT=$(PORT) -p $(PORT):$(PORT) $(imageName):backend 

redis:
	$(eval CONTAINER_ID := $(shell docker ps -a --filter name=$(redisName) --format '{{.ID}}'))
	@echo "Removing the container with ID : $(CONTAINER_ID)"
	@if [ -n "$(CONTAINER_ID)" ]; then \
	  docker stop $(redisName) > /dev/null 2>&1; \
	  docker rm $(redisName) > /dev/null 2>&1; \
	fi
	@echo "Container stopped and removed successfully"
	@echo "Running redis server with configuration..."
	@docker run -d \
	--name $(redisName) \
	-p 6379:6379 \
	redis \
	--requirepass $(REDIS_PASSWORD)
	@echo "Done.\n"

build:
	@echo "Building the application..."
	@go build -o ./bin/app ./cmd/event-reservation-app/*.go
	@echo "Application built..."

test:
	@echo "Running all tests...\n"
	@go test -v -race ./tests/... -run TestGETEvents

testall:
	@echo "Running all tests...\n"
	@go test -v ./tests/...

clean:
	@echo "Removing all binaries..."
	@rm -rf ./bin/*
	@echo "All binaries removed !"

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
