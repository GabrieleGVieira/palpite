.PHONY: run test fmt vet

run:
	set -a; [ -f .env ] && . ./.env; set +a; go run ./cmd/api

test:
	go test ./...

fmt:
	go fmt ./...

vet:
	go vet ./...

seed:
	go run ./cmd/seed/main.go
