.PHONY:  run format

run:
	go run ./cmd/md-parser/main.go cmd/md-parser/test.md

format:
	go fmt ./...