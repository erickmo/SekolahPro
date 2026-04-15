SERVICE=boilerplate
MAIN=./cmd/api
BIN=./bin/$(SERVICE)

.PHONY: build dev air test test-race test-integration lint sqlc mock migrate-up migrate-down migrate-create infra-up infra-down tidy tools-install reset swagger

build:
	go build -o $(BIN) $(MAIN)

dev:
	go clean -cache
	rm -f $(BIN)
	go build -o $(BIN) $(MAIN)
	@-lsof -ti :8080 | xargs kill -9 2>/dev/null || true
	$(BIN)

air:
	@-lsof -ti :8080 | xargs kill -9 2>/dev/null || true
	air

test:
	go test ./internal/... ./pkg/... -v -timeout 30s

test-race:
	go test -race ./... -timeout 60s

test-integration:
	go test ./tests/integration/... -v -timeout 120s -tags integration

lint:
	golangci-lint run ./...

sqlc:
	sqlc generate

mock:
	mockery --all --dir=internal/domain --output=internal/mocks --outpkg=mocks

migrate-up:
	@set -a; [ -f .env ] && . ./.env; set +a; bash scripts/migrate.sh up "$$DATABASE_URL"

migrate-down:
	@set -a; [ -f .env ] && . ./.env; set +a; bash scripts/migrate.sh down "$$DATABASE_URL"

migrate-create:
	@next=$$(ls migrations/[0-9]*.sql 2>/dev/null | grep -oE '/[0-9]+_' | grep -oE '[0-9]+' | sort -n | tail -1); \
	next=$$(( $${next:-0} + 1 )); \
	printf -v padded '%03d' $$next; \
	printf -- "-- +migrate Up\n\n-- +migrate Down\n" > "migrations/$${padded}_$(name).sql"; \
	echo "Created migrations/$${padded}_$(name).sql"

infra-up:
	docker compose up -d

infra-down:
	docker compose down

tidy:
	go mod tidy

docker-build:
	docker build -t $(SERVICE):latest .

swagger:
	@mkdir -p docs
	swag init -g cmd/api/main.go -o docs --parseDependency --parseInternal
	@echo "✓ Swagger docs digenerate di docs/"

tools-install:
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
	go install github.com/vektra/mockery/v2@latest
	go install github.com/air-verse/air@latest
	go install github.com/swaggo/swag/cmd/swag@latest

run: infra-up
	sleep 3 && go run $(MAIN)

reset:
	@echo ""
	@echo "⚠️  RESET — Ini akan:"
	@echo "  1. Kill proses di port 8080"
	@echo "  2. Hapus file .env"
	@echo "  3. Clean build cache + build ulang binary"
	@echo ""
	@echo "Ketik YES! untuk melanjutkan:"
	@read confirm && [ "$$confirm" = "YES!" ] || (echo "Dibatalkan." && exit 1)
	@echo ""
	@-lsof -ti :8080 | xargs kill -9 2>/dev/null || true
	@echo ">>> [1/3] Menghapus .env..."
	rm -f .env
	@echo ">>> [2/3] Clean build..."
	go clean -cache
	rm -f $(BIN)
	@echo ">>> [3/3] Build + run..."
	go build -o $(BIN) $(MAIN)
	@echo ""
	@echo "✓ Reset selesai. Menjalankan API..."
	$(BIN)
