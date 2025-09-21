default: run
include .env

.PHONY: run
current-year:
	@.bin/current-year.sh

current-month:
	@.bin/current-month.sh

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

.PHONY: clear
clear:
	@rm ./tmp/bootstrap

.PHONY: zip
zip:
	@zip -j ./tmp/lambda.zip ./tmp/bootstrap

.PHONY: deploy
deploy:
	@make build && make zip && make clear
