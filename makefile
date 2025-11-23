.PHONY: down up up-default setup format lint unit e2e

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

e2e:
	@echo "Running e2e tests"
	docker-compose -f docker-compose.e2e.yml up --build --abort-on-container-exit --exit-code-from e2e-tests e2e-tests
	docker-compose -f docker-compose.e2e.yml down --volumes --remove-orphans;
