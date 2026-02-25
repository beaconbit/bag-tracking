.PHONY: build run test clean tidy

build: tidy
	@echo "Building..."
	@templ generate
	@go build -o piechart-demo .

run: build
	@echo "Starting server..."
	@./piechart-demo

dev:
	@echo "Starting dev server..."
	@templ generate --watch & \
		go run main.go & \
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