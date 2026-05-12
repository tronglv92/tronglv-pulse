.PHONY: run ingest worker cron sim test coverage grpc validate mocks init clean migrate-up migrate-down migrate-status migrate-force

MIGRATIONS_PATH := db/migrations

run:
	go run ./cmd/api/main.go -f etc/api.yaml

ingest:
	go run ./cmd/ingest/main.go -f etc/ingest.yaml

worker:
	go run ./cmd/worker/main.go -f etc/worker.yaml

cron:
	go run ./cmd/cron/main.go -f etc/cron.yaml

sim:
	go run ./cmd/sim/main.go

test:
	go test ./... \
		-coverprofile=coverage.out \
		-covermode=atomic \
		-run . \
		$(shell go list ./... | grep -v 'api/.*\.pb' | grep -v 'mock_')

coverage: test
	go tool cover -html=coverage.out -o coverage.html

grpc:
	protoc \
		-I api \
		-I protos \
		--go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		api/**/*.proto

validate:
	protoc \
		-I api \
		-I protos \
		--validate_out="lang=go,paths=source_relative:." \
		api/**/*.proto

gateway:
	protoc \
		-I api \
		-I protos \
		--grpc-gateway_out=. --grpc-gateway_opt=paths=source_relative \
		api/**/*.proto

mocks:
	mockery --all --dir internal/contract --output internal/contract/mock --outpkg mock

init:
	go install github.com/air-verse/air@latest
	go install github.com/vektra/mockery/v2@latest
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

## ── Database migrations ──────────────────────────────────────────────────────

migrate-up:
	@set -a; [ -f .env ] && . ./.env; set +a; \
	if [ -z "$$DB_HOST" ] || [ -z "$$DB_USER" ] || [ -z "$$DB_NAME" ]; then \
		echo "ERROR: DB variables not set. Run: cp .env.example .env"; exit 1; \
	fi; \
	DSN="postgres://$$DB_USER:$$DB_PASSWORD@$$DB_HOST:$$DB_PORT/$$DB_NAME?sslmode=disable"; \
	echo "Applying all pending migrations..."; \
	migrate -path $(MIGRATIONS_PATH) -database "$$DSN" up; \
	echo "Done."

migrate-down:
	@set -a; [ -f .env ] && . ./.env; set +a; \
	if [ -z "$$DB_HOST" ] || [ -z "$$DB_USER" ] || [ -z "$$DB_NAME" ]; then \
		echo "ERROR: DB variables not set. Run: cp .env.example .env"; exit 1; \
	fi; \
	echo "WARNING: This will roll back the last migration and may cause data loss."; \
	printf "Type 'yes' to confirm: "; read CONFIRM; \
	if [ "$$CONFIRM" = "yes" ]; then \
		DSN="postgres://$$DB_USER:$$DB_PASSWORD@$$DB_HOST:$$DB_PORT/$$DB_NAME?sslmode=disable"; \
		migrate -path $(MIGRATIONS_PATH) -database "$$DSN" down 1; \
		echo "Rolled back 1 migration."; \
	else \
		echo "Aborted."; \
	fi

migrate-status:
	@set -a; [ -f .env ] && . ./.env; set +a; \
	if [ -z "$$DB_HOST" ] || [ -z "$$DB_USER" ] || [ -z "$$DB_NAME" ]; then \
		echo "ERROR: DB variables not set. Run: cp .env.example .env"; exit 1; \
	fi; \
	DSN="postgres://$$DB_USER:$$DB_PASSWORD@$$DB_HOST:$$DB_PORT/$$DB_NAME?sslmode=disable"; \
	migrate -path $(MIGRATIONS_PATH) -database "$$DSN" version

migrate-force:
	@if [ -z "$(VERSION)" ]; then echo "Usage: make migrate-force VERSION=<version>"; exit 1; fi
	@set -a; [ -f .env ] && . ./.env; set +a; \
	if [ -z "$$DB_HOST" ] || [ -z "$$DB_USER" ] || [ -z "$$DB_NAME" ]; then \
		echo "ERROR: DB variables not set. Run: cp .env.example .env"; exit 1; \
	fi; \
	echo "WARNING: Force-setting version to $(VERSION). Use only to fix a dirty state."; \
	printf "Type 'yes' to confirm: "; read CONFIRM; \
	if [ "$$CONFIRM" = "yes" ]; then \
		DSN="postgres://$$DB_USER:$$DB_PASSWORD@$$DB_HOST:$$DB_PORT/$$DB_NAME?sslmode=disable"; \
		migrate -path $(MIGRATIONS_PATH) -database "$$DSN" force $(VERSION); \
		echo "Version forced to $(VERSION)."; \
	else \
		echo "Aborted."; \
	fi

## ── Cleanup ──────────────────────────────────────────────────────────────────

clean:
	rm -f pulse-api pulse-ingest pulse-worker pulse-cron coverage.out coverage.html build-errors.log
	rm -rf tmp/
