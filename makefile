.PHONY: default dev test auth install

default: dev

dev:
	@air
install:
	@go mod download
test:
	@go test -count=1 -v ./...
docs:
	@swag init -g ./cmd/server/main.go -o ./internal/app/http/docs
create_migration:
	@npx prisma migrate dev --create-only --schema=./sql/schema.prisma
migrate:
	@npx prisma migrate deploy --schema=./sql/schema.prisma
db_generate:
	@sqlc generate
db_ui:
	@npx prisma studio --schema=./sql/schema.prisma
