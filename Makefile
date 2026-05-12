.PHONY: run ingest worker cron sim test coverage grpc validate mocks init clean

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

clean:
	rm -f pulse-api pulse-ingest pulse-worker pulse-cron coverage.out coverage.html build-errors.log
	rm -rf tmp/
