# StudyBuddy API Gateway

Single public HTTP entrypoint for all StudyBuddy microservices. Validates JWT on protected routes, applies CORS and rate limits, and reverse-proxies to internal services.

## Stack

- Go (`studybuddy/backend` module)
- [chi](https://github.com/go-chi/chi) v5
- [go-chi/cors](https://github.com/go-chi/cors) — no CORS exists elsewhere in the repo today
- `golang.org/x/time/rate` — in-memory token bucket (not distributed; state resets on restart and is per gateway instance)

## Quick start

From repo root:

```bash
cp api-gateway/.env.example .env   # or merge vars into root .env
docker compose up -d               # gateway on :8000; backends internal only
```

Local run (backends must be reachable at `*_SERVICE_URL`):

```bash
go run ./cmd/api-gateway
```

Health: `GET http://localhost:8000/health`

## Environment variables

See [`.env.example`](.env.example) for the full list. Required:

| Variable | Default | Description |
|----------|---------|-------------|
| `GATEWAY_PORT` | `8000` | Listen port |
| `JWT_SECRET` | — | Must match all services (HS256) |
| `AUTH_SERVICE_URL` … `POINTS_SERVICE_URL` | `http://<service>:<port>` | Upstream bases (Docker DNS names in compose) |
| `CORS_ALLOWED_ORIGINS` | localhost dev origins | Comma-separated |
| `RATE_LIMIT_RPM` | `100` | Global per-IP requests/minute |
| `RATE_LIMIT_AUTH_RPM` | `20` | Per-IP limit on `/api/v1/auth/login` and `/api/v1/auth/register` |
| `HEALTH_CHECK_TIMEOUT` | `3s` | Downstream probe timeout |
| `SHUTDOWN_TIMEOUT` | `10s` | Graceful shutdown drain |

## Routing table

Paths are forwarded **unchanged** to the upstream (no prefix stripping).

| Path prefix | Upstream env var | Default target |
|-------------|------------------|------------------|
| `GET /health` | *(gateway)* | Aggregates all 8 services |
| `/api/v1/auth/*` | `AUTH_SERVICE_URL` | `http://auth:8080` |
| `/api/v1/users/{userID}/points` | `POINTS_SERVICE_URL` | `http://points:8087` |
| `/api/v1/users/*` | `USERS_SERVICE_URL` | `http://users:8081` |
| `/api/v1/interests`, `/api/v1/interests/*` | `USERS_SERVICE_URL` | `http://users:8081` |
| `/api/v1/admin/*` | `USERS_SERVICE_URL` | `http://users:8081` |
| `/api/v1/courses`, `/api/v1/courses/*` | `COURSES_SERVICE_URL` | `http://courses:8082` |
| `/api/v1/availability/*` | `AVAILABILITY_SERVICE_URL` | `http://availability:8083` |
| `/api/v1/sessions`, `/api/v1/sessions/*` | `AVAILABILITY_SERVICE_URL` | `http://availability:8083` |
| `/api/v1/matching/*` | `MATCHING_SERVICE_URL` | `http://matching:8084` |
| `/api/v1/groups`, `/api/v1/groups/*` | `GROUPS_SERVICE_URL` | `http://groups:8085` |
| `/api/v1/reviews`, `/api/v1/reviews/*` | `REVIEWS_SERVICE_URL` | `http://reviews:8086` |
| `/api/v1/points/*` | `POINTS_SERVICE_URL` | `http://points:8087` |

## Auth at the gateway

**Public (no JWT):**

- `/health`
- `/api/v1/auth/register`, `/login`, `/refresh`
- `/api/v1/reviews/users/{userID}/rating`
- `/api/v1/users/{userID}/points`
- `/api/v1/availability/gcal/callback`

**Protected:** all other `/api/v1/*` — gateway validates Bearer JWT via `pkg/auth`, returns `401 {"error":"unauthorized"}` on failure, otherwise proxies with the original `Authorization` header.

## Google Calendar redirect

Register OAuth redirect URL through the gateway:

`http://localhost:8000/api/v1/availability/gcal/callback`

Set on the **availability** service: `GCAL_REDIRECT_URL=http://localhost:8000/api/v1/availability/gcal/callback`

## Errors

Gateway-generated errors use `pkg/httputil`: `{"error":"..."}`. Downstream error bodies are not modified.

## Layout

```
cmd/api-gateway/main.go
api-gateway/delivery/
  router.go       — Chi routes and middleware chain
  proxy.go        — ReverseProxy
  middleware.go   — recovery, request ID, logging, CORS, rate limit, auth
  health.go       — GET /health aggregation
```
