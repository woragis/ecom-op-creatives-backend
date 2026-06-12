# Server layers and errors

Convenções de erros e camadas HTTP — padrão Lingo.

---

## Erros (`internal/apperrors`)

- Um `Code*` por ramo de falha — grepável, estável para cliente
- Mensagem pública em **inglês**
- `Kind` → HTTP status centralizado

### Exemplo de nomenclatura

```text
PRODUCT_GET_V1_SERVICE_NOT_FOUND
CREATIVE_RUN_CREATE_V1_SERVICE_INVALID_PRODUCT
PIPELINE_STEP_V1_SERVICE_ALREADY_RUNNING
PIPELINE_REPROCESS_V1_SERVICE_STEP_NOT_EDITABLE
VIDEO_GENERATE_V1_PROVIDER_RATE_LIMITED
```

### Resposta HTTP

```json
{
  "code": "PRODUCT_GET_V1_SERVICE_NOT_FOUND",
  "message": "Product not found"
}
```

---

## Handler rules

1. Parse request → validate transport
2. Extract auth (JWT)
3. Call service
4. Map error → `WriteError(w, err)`
5. Write JSON success

Handlers **nunca** importam GORM ou RabbitMQ diretamente.

---

## Service rules

1. Business validation
2. Transactions quando necessário
3. Chamar repository + platform (rabbitmq, s3)
4. Retornar `apperrors` tipados

---

## Repository rules

1. Apenas queries
2. Retornar `gorm.ErrRecordNotFound` — service traduz para apperror
3. Sem lógica de pipeline

---

## Idempotência

`Idempotency-Key` header em POSTs com efeito:

- `POST /v1/creative-runs`
- `POST /v1/pipeline/steps/{id}/reprocess`

Redis opcional (fase 2); até lá dedup por constraint DB.
