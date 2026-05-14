.PHONY: run migrate-up migrate-down migrate-new

run:
	go run ./cmd/server

migrate-up:
	@set -a; source .env; set +a; \
	migrate -database "$$DATABASE_URL" -path ./migrations up

migrate-down:
	@set -a; source .env; set +a; \
	migrate -database "$$DATABASE_URL" -path ./migrations down 1

migrate-new:
	@read -p "Migration name: " name; \
	migrate create -ext sql -dir ./migrations -seq $$name