build:
	@go build -o bin/fs driver/main.go

run: build
	@./bin/fs

test:
	@go test ./... -v -cover -race
