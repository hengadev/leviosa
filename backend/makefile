run: 
	@echo "Running the application..."
	@go run ./cmd/event-reservation-app/main.go

build:
	@echo "Building the application..."
	@go build -o ./bin/app ./cmd/event-reservation-app/main.go
	@echo "Application built..."

test:
	@echo "Running all tests...\n"
	@go test -v ./tests/...

clean:
	@echo "Removing all binaries..."
	@rm -rf ./bin/*
	@echo "All binaries removed !"
