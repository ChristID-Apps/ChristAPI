.PHONY: help docker-build docker-up docker-down docker-logs docker-restart docker-migrate-up docker-migrate-down

help:
	@echo "Available commands:"
	@echo "  make docker-build    - Build Docker image"
	@echo "  make docker-up       - Start services with docker-compose"
	@echo "  make docker-down     - Stop services"
	@echo "  make docker-logs     - View container logs"
	@echo "  make docker-restart  - Restart services"
	@echo "  make docker-shell    - Open shell in API container"
	@echo "  make docker-db-shell - Open PostgreSQL shell"
	@echo "  make docker-migrate-up   - Apply pending migrations"
	@echo "  make docker-migrate-down - Roll back one migration"

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
	docker compose up -d postgres
	docker compose run --rm migrate -path=/migrations -database "postgres://christ_user:christ_password@postgre-chrisapi:5432/christ_db?sslmode=disable" up

docker-migrate-down:
	docker compose up -d postgres
	docker compose run --rm migrate -path=/migrations -database "postgres://christ_user:christ_password@postgre-chrisapi:5432/christ_db?sslmode=disable" down 1

docker-clean:
	docker-compose down -v
	docker system prune -f
