.PHONY: down up up-dev

down:
	docker-compose down --volumes --remove-orphans

up:
	docker-compose up -d

up-dev:
	docker-compose up -d --build

format:
	gofmt -w .