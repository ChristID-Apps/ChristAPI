.PHONY: help docker-build docker-up docker-down docker-logs docker-restart

help:
	@echo "Available commands:"
	@echo "  make docker-build    - Build Docker image"
	@echo "  make docker-up       - Start services with docker-compose"
	@echo "  make docker-down     - Stop services"
	@echo "  make docker-logs     - View container logs"
	@echo "  make docker-restart  - Restart services"
	@echo "  make docker-shell    - Open shell in API container"
	@echo "  make docker-db-shell - Open PostgreSQL shell"

docker-build:
	docker-compose build

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

docker-logs:
	docker-compose logs -f

docker-restart:
	docker-compose restart

docker-shell:
	docker-compose exec api sh

docker-db-shell:
	docker-compose exec postgres psql -U ${DB_USER:-christ_user} -d ${DB_NAME:-christ_db}

docker-migrate-up:
	docker compose run --rm migrate -path=/migrations -database "postgres://$(DB_USER):$(DB_PASSWORD)@postgre-chrisapi:5432/$(DB_NAME)?sslmode=$(DB_SSLMODE)" up

docker-migrate-down:
	docker compose run --rm migrate -path=/migrations -database "postgres://$(DB_USER):$(DB_PASSWORD)@postgre-chrisapi:5432/$(DB_NAME)?sslmode=$(DB_SSLMODE)" down

docker-clean:
	docker-compose down -v
	docker system prune -f
