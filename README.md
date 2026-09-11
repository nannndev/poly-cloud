# Poly Cloud

Storage aggregator monorepo based on `docs/`: Go API, Nuxt 4 frontend, Postgres metadata index, and rclone engine.

## Quick start

```sh
cp .env.example .env
make dev
```

- Frontend: http://localhost:3000
- API health: http://localhost:8080/healthz
- Postgres: localhost:5432

The current scaffold establishes the deployment boundary and database schema. OAuth, provider accounts, indexing, routing, and stream-through file operations are the next implementation layer.
