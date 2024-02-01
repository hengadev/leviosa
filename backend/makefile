run: 
	@echo "Running the application..."
	@go run ./cmd/event-reservation-app/main.go

build:
	@echo "Building the application..."
	@go build -o ./bin/app ./cmd/event-reservation-app/main.go
	@echo "Application built..."

test:
	@echo "Running all tests...\n"
	@# NOTE: To run all the tests when we are done
	@# @go test -v ./tests/...
	@go test -v ./tests/... -run TestPOSTSignUp

clean:
	@echo "Removing all binaries..."
	@rm -rf ./bin/*
	@echo "All binaries removed !"
