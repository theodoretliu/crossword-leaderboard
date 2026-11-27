.PHONY: dev deploy logs restart-scraper migration

REMOTE ?= theodoretliu@136.118.50.156
COMPOSE = cd /opt/crossword && sudo docker compose --profile production

# Postgres + Go API + Vite with the /api proxy, at http://localhost:20000
dev:
	bin/dev

# Build locally, ship to the server, restart, wait for health checks
deploy:
	./deploy.sh

logs:
	ssh $(REMOTE) '$(COMPOSE) logs -f --tail 100'

restart-scraper:
	ssh $(REMOTE) '$(COMPOSE) restart scraper'

migration:
	migrate create -ext sql -dir migrations -seq migration
