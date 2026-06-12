# Backend — arquitetura

Adaptação do padrão Lingo para ecom-op-creatives.

---

## Runtimes

```text
Frontend (Next.js)
       │
       ▼
Go API (server/)              ← REST /v1, auth, pipeline CRUD, enqueue
       │
       ├─ PostgreSQL            ← estado, pipeline_steps, artifacts
       ├─ RabbitMQ              ← job dispatch
       └─ S3/MinIO
              │
              ▼
       worker-llm (Go)
       worker-media (Go)
       worker-render (Node)
```

---

## Camadas

```text
handler  →  service  →  repository
   │            │
   │            └─ RabbitMQ publish, orquestração pipeline
   └─ HTTP parse, auth, JSON
```

| Camada | Responsabilidade | Proibido |
|--------|------------------|----------|
| **Handler** | Parse, auth, chamar service, WriteError | GORM, regras de negócio |
| **Service** | Validações, transações, state machine pipeline | http.Request |
| **Repository** | Queries GORM/Postgres | apperrors, RabbitMQ |

---

## Layout server/

```text
server/
├── cmd/server/main.go
├── cmd/migrate/main.go
└── internal/
    ├── httpserver/
    ├── middleware/
    ├── apperrors/
    ├── models/
    ├── migrate/
    ├── platform/postgres/
    ├── platform/rabbitmq/
    ├── platform/s3/
    ├── agent/              # agentes LLM (shared com worker-llm)
    ├── product/
    ├── campaign/
    ├── creativerun/
    ├── pipeline/
    └── auth/
```

---

## Domínios

| Módulo | Responsabilidade |
|--------|------------------|
| `product` | CRUD produtos, URL scrape metadata |
| `campaign` | Campanhas, agendamento, batch runs |
| `creativerun` | Execução pipeline, status |
| `pipeline` | Steps, enqueue, reprocess, invalidate |
| `agent` | Prompts, LLM clients, schemas |

---

## HTTP

| Item | Convenção |
|------|-----------|
| Prefixo | `/v1` |
| IDs | UUID |
| JSON | camelCase |
| Erros | `{ "code": "...", "message": "..." }` |
| Health | `GET /health`, `GET /ready` |

---

## Pipeline state machine

Service layer em `internal/pipeline/`:

```go
// EnqueueStep publica job RabbitMQ após persistir step
func (s *Service) EnqueueStep(ctx context.Context, runID, stepType string) error

// InvalidateFromStep marca downstream como invalidated
func (s *Service) InvalidateFromStep(ctx context.Context, stepID uuid.UUID) error

// CompleteStep salva output, enfileira próximo
func (s *Service) CompleteStep(ctx context.Context, stepID uuid.UUID, output json.RawMessage) error
```

---

## Workers

Workers consomem RabbitMQ e chamam packages em `internal/agent/`, `internal/video/`, etc.

Código compartilhado entre server e workers via `internal/` no mesmo module ou packages locais.

Ver [../docs/WORKERS.md](../docs/WORKERS.md).

---

## Referências

- [SERVER_LAYERS_AND_ERRORS.md](./SERVER_LAYERS_AND_ERRORS.md)
- [adr/0001-go-gorm-layers.md](./adr/0001-go-gorm-layers.md)
- [adr/0002-rabbitmq-pipeline.md](./adr/0002-rabbitmq-pipeline.md)
- [adr/0003-video-providers-pluggable.md](./adr/0003-video-providers-pluggable.md)
