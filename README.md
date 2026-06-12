# ecom-op-creatives — backend

Go API + workers para pipeline de criativos e-commerce.

## Repositório

Parte do monorepo [ecom-op-creatives](https://github.com/woragis/ecom-op-creatives) como submodule.

## Stack

| Componente | Tecnologia |
|------------|------------|
| API | Go 1.22+, stdlib HTTP, GORM |
| DB | PostgreSQL |
| Filas | RabbitMQ |
| Workers | Go (llm, media) + Node (render — submodule ou sibling) |

## Estrutura

```text
backend/
├── docs/
│   ├── ARCHITECTURE.md
│   ├── SERVER_LAYERS_AND_ERRORS.md
│   └── adr/
├── migrations/
├── server/              # Go HTTP API
│   ├── cmd/server/
│   ├── cmd/migrate/
│   └── internal/
├── worker-llm/
├── worker-media/
└── worker-render/       # Node — ver worker-render/README.md
```

## Camadas (padrão Lingo)

```text
handler → service → repository
```

## Desenvolvimento

```bash
cp .env.example .env
docker compose up -d
make migrate
make dev          # API on :8080
make worker-stub  # pipeline stub worker (separate terminal)
```

Integration tests:

```bash
DATABASE_URL=postgres://creatives:creatives@localhost:5432/creatives?sslmode=disable \
  go test -tags=integration ./internal/product/repository/...
```

## Phase 0 entregue

- Go API: `GET /health`, `GET /ready`
- CRUD products + creative runs
- Pipeline com 12 steps + RabbitMQ enqueue
- Worker stub (`cmd/worker-stub`) marca steps como done
- Migrations SQL + AutoMigrate

## Documentação

- [docs/ARCHITECTURE.md](./docs/ARCHITECTURE.md)
- [docs/SERVER_LAYERS_AND_ERRORS.md](./docs/SERVER_LAYERS_AND_ERRORS.md)
- [../ARCHITECTURE.md](../ARCHITECTURE.md) (monorepo root)

## Módulo Go

```text
github.com/woragis/ecom-op-creatives-backend/server
```
