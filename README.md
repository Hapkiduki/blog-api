# blog-api

Golang api with graphql, docker, logging, tracing, feature first, postgres, redis

## Folder structure

```
blog-api/
├── cmd/server/main.go          ← Entry point. Wires everything together.
├── internal/                   ← Private application code. Cannot be imported by other Go modules.
│   ├── post/                   ← Post domain: model, repository, service, resolver
│   ├── user/                   ← User domain: same layered structure
│   ├── comment/                ← Comment domain
│   ├── tag/                    ← Tag domain
│   └── middleware/             ← HTTP middlewares (logging, tracing, auth)
├── graph/
│   ├── schema/                 ← GraphQL schema files (.graphqls) — the source of truth
│   └── generated/              ← Auto-generated code by gqlgen (never edit manually)
├── pkg/                        ← Shared utility packages. CAN be imported by other projects.
│   ├── logger/                 ← Zap logger constructor
│   ├── tracer/                 ← OpenTelemetry tracer setup
│   ├── database/               ← pgx connection pool
│   └── ...                     ← Other shared utilities
├── migrations/                 ← SQL migration files (up/down pairs)
├── docker/                     ← Prometheus config, Grafana dashboards, etc.
├── docker-compose.yml          ← Local dev services (Postgres, Jaeger, etc.)
├── Dockerfile                  ← Multi-stage container build
├── Makefile                    ← Developer commands
└── .env.example                ← Template for environment variables
```
