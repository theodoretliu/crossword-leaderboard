# Development
`bin/dev` (or `make dev`) starts everything at http://localhost:20000:

- Postgres in docker on :20001, seeded from `crossword_production.dump` on first run if present
- the Go API on :20002
- the Vite dev server, which proxies `/api/*` to the API with the prefix stripped (same as production)

# Deployment
Everything is served from one origin by Caddy on a single VM (Google Cloud, account theodoretliu@gmail.com, project Kipp, Compute Engine):

- `/api/*` goes to the Go API container (prefix stripped)
- everything else is the static Vite build, with an SPA fallback to `index.html`

`./deploy.sh` (or `make deploy`) builds the backend image and frontend locally, uploads them, and runs
`docker compose --profile production up -d --wait`. See the header of `deploy.sh` for rollbacks.
`make logs` tails the server logs.