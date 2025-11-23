.PHONY: down up up-default setup format lint unit

setup:
	go mod tidy
	go mod download

down:
	docker-compose down --volumes --remove-orphans

up:
	docker-compose up -d --build

up-default:
	cp .env.example .env || true
	docker-compose up -d --build

format:
	gofmt -w .
	goimports -w .

lint:
	golangci-lint run ./...

mocks:
	mockery

unit:
	./scripts/coverage.sh
