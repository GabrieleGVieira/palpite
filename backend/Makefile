.PHONY: run test fmt vet migrate seed matchsync explanations

LIMIT ?= 15

run:
	set -a; [ -f .env ] && . ./.env; set +a; go run ./cmd/api

test:
	go test ./...

fmt:
	go fmt ./...

vet:
	go vet ./...

migrate:
	set -a; [ -f .env ] && . ./.env; set +a; go run ./cmd/migrate

seed:
	go run ./cmd/seed/main.go

matchsync:
	go run ./cmd/matchsync/main.go

explanations:
	@test -n "$(FROM_DATE)" || (echo "FROM_DATE is required. Usage: make explanations FROM_DATE=2026-06-01 TO_DATE=2026-07-31 LIMIT=50"; exit 1)
	@test -n "$(TO_DATE)" || (echo "TO_DATE is required. Usage: make explanations FROM_DATE=2026-06-01 TO_DATE=2026-07-31 LIMIT=50"; exit 1)
	set -a; [ -f .env ] && . ./.env; set +a; go run ./internal/workers/generate_prediction_explanations --from-date=$(FROM_DATE) --to-date=$(TO_DATE) --limit=$(LIMIT)
