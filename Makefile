.PHONY: build run test vet

build:
	go build -o design-to-code ./cmd/parser/

run:
	go run cmd/parser/main.go --dir testdata/

run-gin:
	go run cmd/parser/main.go --dir /tmp/gin --output /tmp/gin-graph.json

test:
	go test ./... -v

vet:
	go vet ./...