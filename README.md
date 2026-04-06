# WoW Log Analyzer

WoW Log Analyzer is an MVP web application that analyzes a player's Warcraft Logs fight by comparing it against a filtered cohort of high-performing players and generating grounded AI insights.

## MVP goal

A user pastes a Warcraft Logs report URL, selects a fight and character, and receives:

- deterministic metric comparisons against a strong comparison cohort
- top 3–5 actionable differences
- concise AI-generated coaching based only on structured metric deltas

## Tech stack

### Frontend

- React
- TypeScript
- Vite
- Tailwind CSS
- Zustand

### Backend

- Go microservices
- API Gateway
- Log Service
- Analysis Service
- AI Insight Service

### Infrastructure

- PostgreSQL
- Redis
- Docker Compose

## Core principles

- Deterministic analysis first
- AI explanation second
- No AI-generated claims directly from raw logs
- Keep outputs auditable and explainable
- Prioritize small, reviewable implementation steps

## Planned MVP flow

1. User pastes a Warcraft Logs URL
2. App parses the report code
3. Log Service fetches report metadata, fights, and player options
4. User selects a fight and character
5. Log Service fetches and normalizes fight data
6. Analysis Service compares the user against a filtered comparison cohort
7. AI Service generates top insights from structured deltas only
8. Frontend renders the report

## Local development

### Prerequisites

- Node.js 20+
- Go 1.22+
- Docker
- Docker Compose

### Start infrastructure

```bash
docker compose up -d postgres redis
```

### Frontend

```bash
cd frontend
npm install
npm run dev
```

### Backend

```bash
sed -i 's/\r$//' .env

set -a
source .env
set +a

cd services/api-gateway
go run ./cmd/server
```

## Initial Services

- API Gateway: single entry point for frontend, routes requests to internal services, shapes responses for frontend, later handle auth/session if needed
- Log Service: Warcraft Logs API integration, report retrieval, fight retrieval, character/actor extraction, event ingestion, normalization into internal models
- Analysis Service: metric computation, cohort aggregation, delta calculation, ranking differences, generating structured insight candidates
- AI Service: accept structured comparison outputs, generate concise insight text, return typed outputs, provide fallback if model call fails

## Initial MVP Metrics

- DPS
- HPS
- Casts per minute
- Major cooldown count
- Major cooldown timing drift
- Buff uptime
- Downtime percentage

## Non-goals for MVP

- full class/spec support
- guild features
- billing
- live background processing
- broad historical dashboards
- encounter-specific expert systems
