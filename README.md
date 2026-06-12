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
make dev              # API :8080
make worker-pipeline  # real pipeline (use LLM_MOCK=1 for local dev)
```

With mocks (no API keys):

```env
LLM_MOCK=1
SERPER_MOCK=1
ELEVENLABS_MOCK=1
RENDER_MOCK=1
```

Integration tests:

```bash
DATABASE_URL=postgres://creatives:creatives@localhost:5432/creatives?sslmode=disable \
  go test -tags=integration ./internal/product/repository/...
```

## Phase 1 entregue

- Research agent: LLM queries → Serper → LLM synthesis
- Agents: hooks, script, director, prompter, supervisor
- ElevenLabs voice step
- Subtitles from script timing
- Remotion render manifest + worker-render
- `worker-pipeline` replaces stub for real execution
- Media served at `GET /media/runs/{id}/...`

## Documentação

- [docs/ARCHITECTURE.md](./docs/ARCHITECTURE.md)
- [docs/SERVER_LAYERS_AND_ERRORS.md](./docs/SERVER_LAYERS_AND_ERRORS.md)
- [../ARCHITECTURE.md](../ARCHITECTURE.md) (monorepo root)

## Módulo Go

```text
github.com/woragis/ecom-op-creatives-backend/server
```
