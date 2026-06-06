.PHONY: build dev visualize

build:
	docker compose build

dev:
	docker compose -f docker-compose.yml -f docker-compose.dev.yml up --build

visualize:
	cd visualizer && pip install -q -r requirements.txt && python visualize.py $(SYMBOLS)
