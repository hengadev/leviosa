run: 
	@echo "Running the application..."
	@# @echo "Removing, just for testing, the previous db.sqlite instance..."
	@# @rm db.sqlite
	@go run ./cmd/event-reservation-app/main.go -port 3500

build:
	@echo "Building the application..."
	@go build -o ./bin/app ./cmd/event-reservation-app/main.go
	@echo "Application built..."

test:
	@echo "Running all tests...\n"
	@go test -v ./tests/... -run TestGETEvents

ta: # for test all
	@echo "Running all tests...\n"
	@go test -v ./tests/...

clean:
	@echo "Removing all binaries..."
	@rm -rf ./bin/*
	@echo "All binaries removed !"
