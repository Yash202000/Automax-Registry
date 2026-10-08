.PHONY: run test tidy

run:
	go run ./cmd/server/main.go

test:
	go test ./...

tidy:
	go mod tidy
