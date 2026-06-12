# ADR 0003 — Video providers plugáveis

## Status

Accepted

## Context

Múltiplos providers de vídeo (Kling, Runway, Luma, Veo) com APIs diferentes, rate limits e qualidade variável. Usuário deve escolher provider na UI; default via env.

## Decision

Interface Go:

```go
type VideoProvider interface {
    ID() string
    Generate(ctx context.Context, req VideoRequest) (VideoJob, error)
    Poll(ctx context.Context, jobID string) (VideoResult, error)
}
```

Prioridade de resolução:

1. `creative_run.video_provider` (UI override)
2. Campaign default
3. `VIDEO_PROVIDER_DEFAULT` env

Registro em worker-media startup.

## Consequences

- Cada provider em package isolado (`internal/video/kling/`, etc.)
- Polling assíncrono unificado no worker-media
- Custo tracking por provider em `api_costs`

## Providers (fase 1)

- kling
- runway
- luma
- veo

## Audio

ElevenLabs fixo — não plugável (decisão de produto).
