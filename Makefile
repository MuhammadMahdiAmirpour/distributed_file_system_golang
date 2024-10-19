build:
	@go build -o bin/fs driver/main.go

run: build
	@./bin/fs

test:
	@go test ./... -v -cover -race

docker-build:
	@docker-compose build

docker-run: docker-build
	@docker-compose up --build

docker-stop:
	@docker-compose stop

docker-down:
	@docker-compose down
