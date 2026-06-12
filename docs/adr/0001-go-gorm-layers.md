# ADR 0001 — Go, GORM, handler-service-repository

## Status

Accepted

## Context

Backend precisa de API rápida, domínio com pipeline complexo, erros estáveis para frontend, deploy simples.

## Decision

Usar padrão Lingo:

- Go 1.22+ stdlib HTTP (ServeMux 1.22+)
- GORM + PostgreSQL
- Camadas handler → service → repository
- Wiring manual em `cmd/server/main.go`
- Sem framework DI (Wire, Fx)
- Sem Gin/Echo

## Consequences

- Consistência com outros projetos woragis (Lingo, CampusWorld)
- Curva baixa para quem já conhece o padrão
- Testes unitários por camada
