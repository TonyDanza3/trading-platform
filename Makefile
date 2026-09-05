.PHONY: up down logs tidy test

up:
	docker compose -f deployments/docker-compose.yml up --build

down:
	docker compose -f deployments/docker-compose.yml down

logs:
	docker compose -f deployments/docker-compose.yml logs -f api-gateway auth-service user-service market-data-service

tidy:
	go work sync

test:
	go test ./...
