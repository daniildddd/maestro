# Configuration

All settings come from the environment (`kelseyhightower/envconfig` under the
hood). Start from [`.env.example`](../.env.example):

```bash
cp .env.example .env
```

> **Gaps in `.env.example`:** it does not list `VALIDATE_DB_*` and `HEALTH_*`
> (both have working defaults, see below). `JWT_SECRET`, `POSTGRES_USER`,
> `POSTGRES_PASSWORD`, `KAFKA_CONNECT_BASE_URL` and `MIDDLEWARE_ALLOWED_ORIGINS`
> have no defaults and must be set.

When running via [`docker-compose.yml`](../docker-compose.yml), the compose file
overrides `POSTGRES_HOST=postgres`, `KAFKA_CONNECT_BASE_URL=http://connect:8083`
and `SERVER_ADDR=:8080` — the values in your `.env` for these keys are ignored
by the `maestro` service (except `POSTGRES_USER/PASSWORD/DB`, which are passed
through).

## Variables

`required` means the process panics on startup when the variable is missing
or empty.

### `LOGGER_*` — logging

| Variable | Default | Required | Purpose |
|----------|---------|----------|---------|
| `LOGGER_LEVEL` | `DEBUG` | — | `DEBUG`, `INFO`, `WARN`, `ERROR` |
| `LOGGER_FOLDER` | — | **yes** | directory for log files (must be writable; the image runs as non-root `65532`) |

### `POSTGRES_*` — app database and CDC source checks

| Variable | Default | Required | Purpose |
|----------|---------|----------|---------|
| `POSTGRES_HOST` | `localhost` | — | compose overrides to `postgres` |
| `POSTGRES_PORT` | `5432` | — | |
| `POSTGRES_USER` | — | **yes** | |
| `POSTGRES_PASSWORD` | — | **yes** | |
| `POSTGRES_DB` | `postgres` | — | |
| `POSTGRES_TIMEOUT` | `5s` | — | connect/query timeout |

### `AUTH_*` / `JWT_*` / `HASHER_*` / `REFRESH_*` — auth

| Variable | Default | Required | Purpose |
|----------|---------|----------|---------|
| `AUTH_COOKIE_SECURE` | `false` | — | `Secure` flag on the refresh cookie (set `true` in production with HTTPS) |
| `AUTH_COOKIE_DOMAIN` | `""` | — | cookie domain |
| `AUTH_CLEANUP_INTERVAL` | `1h` | — | how often expired refresh tokens are swept |
| `JWT_SECRET` | — | **yes** | HMAC secret, 32+ random chars; rotation invalidates all sessions |
| `JWT_ACCESS_TTL` | `15m` | — | access token lifetime |
| `JWT_ISSUER` | `maestro` | — | `iss` claim |
| `HASHER_COST` | `10` | — | bcrypt cost, 4–14 |
| `REFRESH_TTL` | `720h` | — | refresh token lifetime (30 days) |

### `SERVER_*` / `MIDDLEWARE_*` — public HTTP server

| Variable | Default | Required | Purpose |
|----------|---------|----------|---------|
| `SERVER_ADDR` | `localhost:8080` | — | compose/Dockerfile override to `:8080` |
| `SERVER_READ_TIMEOUT` | `30s` | — | |
| `SERVER_READ_HEADER_TIMEOUT` | `10s` | — | |
| `SERVER_WRITE_TIMEOUT` | `30s` | — | |
| `SERVER_IDLE_TIMEOUT` | `120s` | — | |
| `SERVER_MAX_HEADER_BYTES` | `8192` | — | |
| `SERVER_SHUTDOWN_TIMEOUT` | `30s` | — | graceful shutdown deadline |
| `SERVER_DISABLE_GENERAL_OPTIONS_HANDLER` | `true` | — | |
| `MIDDLEWARE_ALLOWED_ORIGINS` | — | **yes** | CORS origins, e.g. `http://127.0.0.1:8080` (use `-` to disable in local single-origin runs) |

### `METRICS_*` / `METRICS_COLLECTOR_*` — internal metrics

| Variable | Default | Required | Purpose |
|----------|---------|----------|---------|
| `METRICS_ADDR` | `:9100` | — | internal Prometheus listener (never exposed to the host in compose) |
| `METRICS_BUCKETS` | `0.05,0.1,0.25,0.5,1,2.5,5,10` | — | histogram buckets for request latency |
| `METRICS_READ_HEADER_TIMEOUT` | `10s` | — | |
| `METRICS_SHUTDOWN_TIMEOUT` | `10s` | — | |
| `METRICS_COLLECTOR_INTERVAL` | `15s` | — | how often connector states are scraped from Connect |
| `METRICS_COLLECTOR_TIMEOUT` | `10s` | — | per-scrape timeout |

### `KAFKA_CONNECT_*` — Connect client

| Variable | Default | Required | Purpose |
|----------|---------|----------|---------|
| `KAFKA_CONNECT_BASE_URL` | — | **yes** | e.g. `http://connect:8083` in compose, `http://localhost:8083` locally |
| `KAFKA_CONNECT_TIMEOUT` | `30s` | — | per-request timeout |
| `KAFKA_CONNECT_RETRY_MAX_ATTEMPTS` | `5` | — | retries on rebalances / 5xx |
| `KAFKA_CONNECT_RETRY_INITIAL_DELAY` | `300ms` | — | exponential backoff base + jitter |

### `VALIDATE_DB_*` — pre-deploy `dbcheck` (Postgres)

Not present in `.env.example`; defaults apply unless set.

| Variable | Default | Required | Purpose |
|----------|---------|----------|---------|
| `VALIDATE_DB_STEP_TIMEOUT` | `10s` | — | per-check-stage DB timeout |
| `VALIDATE_DB_MAX_TABLE_CHECKS` | `50` | — | cap on per-table checks |

### `HEALTH_*` — health endpoint

Not present in `.env.example`.

| Variable | Default | Required | Purpose |
|----------|---------|----------|---------|
| `HEALTH_READY_TIMEOUT` | `2s` | — | readiness probe deadline (`GET /healthz`) |

### Misc — Alertmanager, Grafana

| Variable | Default | Required | Purpose |
|----------|---------|----------|---------|
| `ALERTMANAGER_SLACK_API_URL` | — | — | Slack webhook; consumed by `make alertmanager-config` into the gitignored `deploy/alertmanager/slack_api_url` file mount |
| `GF_SECURITY_ADMIN_USER` | `admin` (compose default) | — | Grafana admin login |
| `GF_SECURITY_ADMIN_PASSWORD` | `admin` (compose default) | — | Grafana admin password |
