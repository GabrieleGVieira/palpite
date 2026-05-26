.PHONY: run test fmt vet migrate

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
