.PHONY: default install test lint update build copy clear zip

default: dev

include .env

run:
	@go run ./cmd/cli/main.go
install:
	@go mod download
test:
	@go test -count=1 -v ./...
lint:
	@golangci-lint run && nilaway ./...
update:
	@go mod tidy && go get -u ./...
build:
	@GOARCH=amd64 GOOS=linux CGO_ENABLED=0 go build -tags lambda.norpc -ldflags="-w -s" -o ./tmp/bootstrap ./cmd/lambda/main.go
copy:
	@cp .env ./tmp/.env && cp users.json ./tmp/users.json
clear:
	@rm ./tmp/.env && rm ./tmp/users.json && rm ./tmp/bootstrap
zip:
	@zip -j ./tmp/lambda.zip ./tmp/bootstrap ./tmp/.env ./tmp/users.json
deploy:
	@make build && make copy && make zip && make clear
