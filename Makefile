.PHONY: build run test clean tidy

TEMPL = $(shell go env GOPATH)/bin/templ

build: tidy
	@echo "Building..."
	@$(TEMPL) generate
	@go build -o piechart-demo .

run: build
	@echo "Starting server..."
	@./piechart-demo

dev:
	@echo "Starting dev server..."
	@$(TEMPL) generate --watch & \
		go run . & \
		wait

tidy:
	@echo "Tidying dependencies..."
	@go mod tidy

clean:
	@echo "Cleaning..."
	@rm -f piechart-demo
	@rm -f components/*_templ.go templates/*_templ.go

test:
	@echo "Running tests..."
	@go test ./...