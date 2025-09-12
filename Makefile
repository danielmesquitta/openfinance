default: run
include .env

MONTH ?= 0
YEAR ?= 0

.PHONY: run
run:
	@go run ./cmd/cli/main.go $(if $(MONTH),--month $(MONTH),) $(if $(YEAR),--year $(YEAR),)

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
