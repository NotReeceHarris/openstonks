.PHONY: build dev

build:
	docker compose build

dev:
	docker compose -f docker-compose.yml -f docker-compose.dev.yml up --build
