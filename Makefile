default: run
include .env

.PHONY: run
run:
	@go run ./cmd/cli/main.go

.PHONY: install
install:
	@go mod download

.PHONY: generate
generate:
	@go generate ./...

.PHONY: test
test:
	@go test -count=1 -v ./...

.PHONY: lint
lint:
	@golangci-lint run && nilaway ./...

.PHONY: update
update:
	@go mod tidy && go get -u ./...

.PHONY: build
build:
	@GOARCH=amd64 GOOS=linux CGO_ENABLED=0 go build -tags lambda.norpc -ldflags="-w -s" -o ./tmp/bootstrap ./cmd/lambda/main.go

.PHONY: copy
copy:
	@cp .env ./tmp/.env && cp users.json ./tmp/users.json

.PHONY: clear
clear:
	@rm ./tmp/.env && rm ./tmp/users.json && rm ./tmp/bootstrap

.PHONY: zip
zip:
	@zip -j ./tmp/lambda.zip ./tmp/bootstrap ./tmp/.env ./tmp/users.json

.PHONY: deploy
deploy:
	@make build && make copy && make zip && make clear
