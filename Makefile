run:
	docker compose up --build

clean:
	docker compose down --volumes

test:
	go test -race -v ./...

linter:
	docker run --rm -v $(PWD):/app -w /app golangci/golangci-lint:v1.59.1 golangci-lint run