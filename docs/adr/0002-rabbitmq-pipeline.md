# ADR 0002 — RabbitMQ para pipeline jobs

## Status

Accepted

## Context

Pipeline de 12 steps, jobs de 30s a 15min (video gen), necessidade de retry, DLQ, persistência e recuperação após falha.

## Decision

- **PostgreSQL** = fonte da verdade (pipeline_steps, output_json)
- **RabbitMQ** = dispatch de jobs por fila tipada
- Filas separadas: research, llm, audio, image, video, render, supervisor
- Cada fila com DLQ
- Retry 3x com backoff antes de DLQ

Kafka reservado para fase 3 (event log analytics), não como job queue.

## Consequences

- Ops RabbitMQ (management UI, monitoring)
- Workers idempotentes obrigatórios
- Cron para stale jobs (running > 30min)

## Alternatives considered

| Opção | Rejeitado porque |
|-------|------------------|
| Redis/BullMQ only | Persistência e DLQ menos robustos para jobs longos |
| Kafka as job queue | Overkill, ops complexo para task queue |
| Sync API only | Video gen impossível síncrono |
