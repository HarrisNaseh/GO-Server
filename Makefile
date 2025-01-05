build:
	@go build -o bin/goserver

run: build
	@./bin/goserver

