up:
	docker compose up

build:
	docker compose up --build

delete:
	docker compose down -v --remove-orphans

down:
	docker compose down


swag:
	swag init -g main.go -d cmd,internal/transport/api/v1/handler,internal/transport/api/v1/dto --parseInternal
