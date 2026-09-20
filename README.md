<p align="center">
  <img src="docs/assets/logo.png" alt="Maestro — Orchestrating Debezium" width="600">
</p>

# Maestro

[![go](https://img.shields.io/badge/go-1.26-00ADD8)](https://go.dev)
[![license](https://img.shields.io/badge/license-MIT-blue)](LICENSE)
[![test](https://github.com/daniildddd/maestro/actions/workflows/test.yml/badge.svg)](https://github.com/daniildddd/maestro/actions/workflows/test.yml)
[![lint](https://github.com/daniildddd/maestro/actions/workflows/lint.yml/badge.svg)](https://github.com/daniildddd/maestro/actions/workflows/lint.yml)

Maestro is a control plane for Debezium CDC connectors running on Kafka Connect.
It wraps the Kafka Connect REST API with an authenticated backend, a web console,
pre-deploy database checks, config validation, audit history, and a full
Prometheus / Grafana observability stack — so you never have to touch
`curl localhost:8083` again.

> **TL;DR:** `cp .env.example .env && make alertmanager-config && make docker-up && make migrate-up` →
> open `http://127.0.0.1:8080`, validate a Postgres config, create a connector,
> watch rows flow into Kafka. Full walkthrough in [Demo scenario](#demo-scenario-postgres--kafka-in-10-minutes).

## Contents

- [What is Maestro?](#what-is-maestro)
- [Features](#features)
- [Architecture](#architecture)
- [Quick start (5 minutes)](#quick-start-5-minutes)
- [Demo scenario: Postgres → Kafka in 10 minutes](#demo-scenario-postgres--kafka-in-10-minutes)
- [Screenshots](#screenshots)
- [Compatibility](#compatibility)
- [Configuration](#configuration)
- [Observability](#observability)
- [Development](#development)
- [Project structure](#project-structure)
- [Contributing](CONTRIBUTING.md)
- [License](#license)

## What is Maestro?

Running Debezium in production usually means juggling three things at once:
Kafka Connect REST (`:8083`), Postgres-side prerequisites (`wal_level`,
publications, `REPLICA IDENTITY`, privileges, slots), and the question
"who changed this connector config and when?".

Maestro puts a single authenticated surface in front of all three:

- **For developers:** a web wizard to create a connector from a live plugin
  schema, with validation before deploy.
- **For operators:** pause / resume / restart (including failed-tasks-only),
  audit log with before/after snapshots, Prometheus alerts + Grafana.
- **For the database:** a pre-deploy `dbcheck` that fails fast with a fix hint
  instead of a dead connector at 3 AM.

Non-goals: Maestro is not a replacement for Kafka Connect or Debezium, not a
schema registry, and not a data pipeline orchestrator. It manages connector
lifecycle — Debezium still does the CDC.

## Features

- **Connector lifecycle** — create, update, delete, pause / resume and hot-restart
  connectors and individual tasks, including isolated restart of failed tasks.
- **Rebalance-aware Kafka Connect client** — HTTP client with timeouts and
  automatic retries with exponential backoff and jitter, built to survive
  Kafka Connect cluster rebalances.
- **Pre-deploy PostgreSQL checks (`dbcheck`)** — before a connector starts,
  Maestro verifies connectivity, PostgreSQL version compatibility, logical
  replication readiness, user privileges, free replication slots and WAL senders,
  publication correctness, and captured tables (existence, `SELECT` access,
  primary key or `REPLICA IDENTITY`, CDC readability). Verdicts come as an
  `ok / warning / error` report with fix recommendations.
- **Schema-based config validation** — Debezium / Kafka Connect configs are
  validated by type, required params and recommended values, with support for
  custom properties.
- **Audit log** — every config change stores JSONB before/after snapshots with
  full change history.
- **Users & RBAC** — JWT access/refresh auth, admin/user roles.
- **Observability as code** — RED metrics on an internal `:9100` listener,
  Kafka Connect metrics via JMX Exporter, 6 Prometheus alerts with Slack
  delivery, and a provisioned Grafana dashboard (`Maestro RED`).
- **Web console** — React + Vite UI in the style of Debezium UI: dashboard with
  live statuses, connector list with task drill-down, schema-driven config
  editor, audit and users pages.

## Architecture

```
                    ┌───────────────┐
                    │  Web console  │  React + Vite (served by :8080)
                    └───────┬───────┘
                            │ JWT
                    ┌───────▼───────┐      ┌───────────────┐
                    │  Maestro API  │─────▶│ Kafka Connect │──▶ Kafka ──▶ Debezium ──▶ Postgres
                    │     :8080     │      │     :8083     │
                    └───────┬───────┘      └───────┬───────┘
                            │                      │ JMX :8084
                    ┌───────▼───────┐      ┌───────▼───────┐
                    │   Postgres    │      │  Prometheus   │──▶ Grafana :3000
                    │  (app + CDC)  │      │     :9090     │──▶ Alertmanager :9093 ──▶ Slack
                    └───────────────┘      └───────────────┘
```

Scrape targets: `maestro:9100` (app RED metrics), `connect:8084` (JMX),
plus Prometheus self-scrape and Alertmanager.

Request flow for a typical "create connector":

1. Web console (or `curl`) calls `POST /api/v1/connectors/validate` with
   `plugin_type + name + config`.
2. Maestro validates types/required fields against the Connect plugin schema,
   then runs Postgres `dbcheck` (connection → permissions → CDC → tables).
3. On `valid: true` the UI enables **Create** → `POST /api/v1/connectors`
   proxies to Kafka Connect, writes an audit entry with the config snapshot.
4. The background collector exports connector/task states as
   `maestro_connector_info/tasks` for Prometheus/Grafana/alerts.

API contract: [`api/swagger.yaml`](api/swagger.yaml) (OpenAPI, `vacuum`-linted).

## Quick start (5 minutes)

Prerequisites: Go 1.26+, Docker with Compose v2, Node 20+ (only for web dev).

```bash
# 1. Configure environment
cp .env.example .env
# Minimal working .env for local demo (edit values for real use):
#   POSTGRES_USER=postgres POSTGRES_PASSWORD=pass POSTGRES_DB=postgres
#   POSTGRES_HOST=postgres (in compose) / localhost (for `make run`)
#   JWT_SECRET=<random 32+ chars>  MIDDLEWARE_ALLOWED_ORIGINS=http://127.0.0.1:8080
#   ALERTMANAGER_SLACK_API_URL=https://hooks.slack.com/... (or leave empty for local)

# 2. JMX exporter jar (gitignored, required for the `connect` service)
make jmx-exporter

# 3. Alertmanager Slack secret (needs ALERTMANAGER_SLACK_API_URL in .env;
#    skip if you only want local metrics without Slack delivery)
make alertmanager-config

# 4. Start the whole stack
make docker-up

# 5. Apply DB migrations
make migrate-up
```

| Service      | URL                      | Credentials (defaults) |
|--------------|--------------------------|------------------------|
| API / Web UI | http://127.0.0.1:8080    | create first user, see below |
| Grafana      | http://127.0.0.1:3000    | `GF_SECURITY_ADMIN_USER` / `GF_SECURITY_ADMIN_PASSWORD` (default `admin`/`admin`) |
| Prometheus   | http://127.0.0.1:9090    | — |
| Alertmanager | http://127.0.0.1:9093    | — |
| Kafka Connect| http://127.0.0.1:8083    | — (proxied by Maestro, do not expose publicly) |

### Create the first admin user

There is no public self-registration: `POST /api/v1/users` requires the
`admin` role, so the very first user must be inserted directly (bootstrap
problem, see [#roadmap](#roadmap--limitations)). With the stack running:

```bash
# pgcrypto crypt() with Blowfish is bcrypt-compatible (default HASHER_COST=10)
docker exec -i maestro-postgres psql -U postgres -d postgres -c \
  "CREATE EXTENSION IF NOT EXISTS pgcrypto;
   INSERT INTO users (username, password_hash, role)
   VALUES ('admin', crypt('admin12345', gen_salt('bf', 10)), 'admin')
   ON CONFLICT (username) DO NOTHING;"
```

Then open `http://127.0.0.1:8080`, sign in as `admin` / `admin12345`,
change the password under Profile, and create the rest of the team via
Users → Create user (or `POST /api/v1/users` with the admin token).

Sanity check:

```bash
curl -s http://127.0.0.1:8080/healthz; echo
curl -s http://127.0.0.1:8083/connectors | head -c 200; echo
```

## Demo scenario: Postgres → Kafka in 10 minutes

Goal: capture changes of `public.orders` in Postgres and see them land in
Kafka, entirely through Maestro.

**0. Stack is up** (see Quick start) and you are logged into the web console
as admin.

**1. Prepare the source table** (run against the compose Postgres):

```bash
docker exec -i maestro-postgres psql -U postgres -d postgres <<'SQL'
CREATE TABLE IF NOT EXISTS public.orders (
  id SERIAL PRIMARY KEY,
  status TEXT NOT NULL DEFAULT 'new',
  total NUMERIC NOT NULL DEFAULT 0
);
ALTER TABLE public.orders REPLICA IDENTITY DEFAULT;
DROP PUBLICATION IF EXISTS maestro_pub;
CREATE PUBLICATION maestro_pub FOR TABLE public.orders;
SQL
```

Why: `dbcheck` will verify the table exists, is readable, has a PK /
`REPLICA IDENTITY`, and is covered by a publication. If you skip this step,
validation returns the exact fix hint instead of failing silently later.

**2. Validate the config before creating anything.**

UI path (recommended): Connectors → Create connector → pick
`io.debezium.connector.postgresql.PostgresConnector` → fill
`topic.prefix=maestro-demo`, `database.hostname=postgres`,
`database.port=5432`, `database.user/postgres/password/dbname`,
`slot.name=maestro_demo`, `publication.name=maestro_pub`,
`table.include.list=public.orders` → Review & validate.

API path:

```bash
TOKEN=$(curl -s -X POST http://127.0.0.1:8080/api/v1/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"admin","password":"admin12345"}' | python3 -c 'import sys,json; print(json.load(sys.stdin)["access_token"])')

curl -s -X POST http://127.0.0.1:8080/api/v1/connectors/validate \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{
    "plugin_type": "io.debezium.connector.postgresql.PostgresConnector",
    "name": "pg-orders-cdc",
    "config": {
      "connector.class": "io.debezium.connector.postgresql.PostgresConnector",
      "topic.prefix": "maestro-demo",
      "database.hostname": "postgres",
      "database.port": "5432",
      "database.user": "postgres",
      "database.password": "pass",
      "database.dbname": "postgres",
      "database.server.name": "maestro-demo",
      "plugin.name": "pgoutput",
      "slot.name": "maestro_demo",
      "publication.name": "maestro_pub",
      "table.include.list": "public.orders"
    }
  }' | python3 -m json.tool
```

Expected: `{"valid": true, "steps": [{"id":"config","status":"ok",...},
{"id":"connection",...}, {"id":"permissions",...}, {"id":"cdc",...},
{"id":"tables",...}]}`. On error you get HTTP 400 with
`CONNECTOR_CONFIG_INVALID` and the same report with `valid: false`,
per-check `fix_hint`, `field` and `table` pointers.

**3. Create the connector.**

```bash
curl -s -X POST http://127.0.0.1:8080/api/v1/connectors \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{"name":"pg-orders-cdc","config":{
    "connector.class":"io.debezium.connector.postgresql.PostgresConnector",
    "topic.prefix":"maestro-demo",
    "database.hostname":"postgres","database.port":"5432",
    "database.user":"postgres","database.password":"pass","database.dbname":"postgres",
    "plugin.name":"pgoutput","slot.name":"maestro_demo",
    "publication.name":"maestro_pub","table.include.list":"public.orders"}}' \
  | python3 -m json.tool
```

The connector appears on the Dashboard as `starting` → `running`
(live-polling, ~7s). Open its detail page for per-task state and trace.

**4. Generate CDC traffic and watch it.**

```bash
docker exec -i maestro-postgres psql -U postgres -d postgres -c \
  "INSERT INTO public.orders (status, total) VALUES ('paid', 42.50);"

docker exec maestro-kafka /opt/kafka/bin/kafka-console-consumer.sh \
  --bootstrap-server localhost:9092 --topic maestro-demo.public.orders --from-beginning --max-messages 1
```

You should see the Debezium envelope (`before`/`after`/`op: c`).
Every create/update/delete is also written to the Audit log page.

**5. Operate it:** pause → resume → restart failed tasks only.

```bash
curl -s -X POST http://127.0.0.1:8080/api/v1/connectors/pg-orders-cdc/pause \
  -H "Authorization: Bearer $TOKEN" | head -c 300; echo
curl -s -X POST http://127.0.0.1:8080/api/v1/connectors/pg-orders-cdc/resume \
  -H "Authorization: Bearer $TOKEN" | head -c 300; echo
# restart connector + tasks, or only failed ones:
curl -s -X POST 'http://127.0.0.1:8080/api/v1/connectors/pg-orders-cdc/restart?include_tasks=true&only_failed=true' \
  -H "Authorization: Bearer $TOKEN" | head -c 300; echo
```

**6. Observe:** Grafana → `Maestro RED` dashboard (latency, 5xx, connector/task
states, Debezium lag); Prometheus alerts fire to Slack on disconnect/lag/down.
Audit log shows who did what with before/after diffs.

**7. Clean up:** `DELETE /api/v1/connectors/pg-orders-cdc` (or Delete button),
`DROP PUBLICATION maestro_pub; DROP TABLE public.orders;`.

## Screenshots

> Screenshots live in `docs/assets/` and are embedded here so the repo page
> works as a product landing. Captured from a live stack
> (`make docker-up`, Postgres 17 + Debezium 3.5.2): dashboard with one running
> connector, the 8-step create wizard, a validation report showing a failing
> `dbcheck` with fix hints, connector detail with task drill-down, audit log,
> and the provisioned `Maestro RED` Grafana dashboard.

| Dashboard (live statuses) | Create wizard (schema-driven) | Validation report (dbcheck) |
|---------------------------|-------------------------------|-----------------------------|
| ![Dashboard](docs/assets/screenshot-dashboard.png) | ![Create wizard](docs/assets/screenshot-create.png) | ![Validation](docs/assets/screenshot-validation.png) |

| Connector detail (tasks) | Audit log (before/after) | Grafana (Maestro RED) |
|--------------------------|--------------------------|-----------------------|
| ![Detail](docs/assets/screenshot-detail.png) | ![Audit](docs/assets/screenshot-audit.png) | ![Grafana](docs/assets/screenshot-grafana.png) |

To refresh them: `make docker-up && make migrate-up`, create the admin user
(see Quick start), run the demo scenario, and re-capture the pages above.

## Compatibility

Pinned in [`docker-compose.yml`](docker-compose.yml) — CDC stacks break silently
on version drift, so check this table before changing images. `dbcheck`
additionally verifies the live database at connector deploy time.

| Component     | Version                |
|---------------|------------------------|
| Go            | 1.26                   |
| Kafka Connect | Debezium 3.5.2         |
| Apache Kafka  | 4.3.1                  |
| PostgreSQL    | 17 (`wal_level=logical`) |
| Prometheus    | 3.5.0                  |
| Alertmanager  | 0.28.1                 |
| Grafana       | 12.2                   |
| Node (web)    | 20+                    |

## Configuration

All settings come from the environment (see [`.env.example`](.env.example),
`kelseyhightower/envconfig` under the hood):

| Prefix            | Purpose                                  |
|-------------------|------------------------------------------|
| `POSTGRES_*`      | app database + CDC source checks         |
| `JWT_*`, `AUTH_*` | auth, sessions, cookie                   |
| `SERVER_*`        | public HTTP server (`:8080`)             |
| `METRICS_*`       | internal metrics listener (`:9100`)      |
| `KAFKA_CONNECT_*` | Connect base URL, timeouts, retry policy |

## Observability

- **Metrics** — `http_server_requests_total`, `http_server_request_duration_seconds`,
  `maestro_connector_info/tasks` on the internal listener (never exposed publicly).
- **Alerts** — [`deploy/prometheus/alert.rules.yml`](deploy/prometheus/alert.rules.yml):
  Debezium lag / disconnect, Connect down, Maestro down, 5xx rate, failed connectors.
- **Dashboards** — [`deploy/grafana/`](deploy/grafana/): datasource + `Maestro RED`
  dashboard are provisioned from git on Grafana startup (datasource, RED, P95,
  connector/task states, Debezium lag). No clicking in the UI — edit JSON, restart.

## Development

```bash
make build             # go build -o bin/maestro ./cmd/maestro
make run               # build + run
make test              # unit tests, -race -cover
make test-integration  # testcontainers suites (needs Docker)
make lint              # golangci-lint
make mocks             # regenerate mockery mocks
make validate-swagger  # vacuum lint api/swagger.yaml
make lint-dockerfile   # hadolint via compose
make lint-trivy        # Trivy HIGH/CRITICAL scan
```

Conventions: table-driven tests with full comparisons, `goleak` in suites,
parallel tests where possible; integration tests live in `test/integration`
(.overlay network via testcontainers, real PostgreSQL); conventional commits.

## Project structure

```
maestro
├── api                      # OpenAPI contract (swagger.yaml, vacuum-linted)
├── cmd
│   └── maestro              # entrypoint (main.go), Dockerfile
├── internal
│   ├── core                 # shared kernel
│   │   ├── domain           # entities: users, connectors, audit, validation
│   │   ├── errs             # common errors
│   │   ├── logger           # zap logger + config
│   │   ├── metrics          # RED metrics, :9100 server
│   │   ├── repository
│   │   │   └── postgres     # pgx pool + adapter
│   │   ├── security         # access (JWT), hasher (bcrypt), refresh tokens
│   │   └── transport        # health, middleware, reqctx, request, response, server, webfs
│   └── features             # vertical slices: auth, users, connectors, audit
│       ├── audit            # repository / service / transport
│       ├── auth             # cleanup / repository / service / transport
│       ├── connectors       # collector / dbcheck / kafkaconnect / plugins / service / transport
│       └── users            # repository / service / transport
├── migrations               # golang-migrate SQL versions
├── deploy                   # infra as code
│   ├── alertmanager         # alertmanager.yml + gitignored slack_api_url
│   ├── grafana              # provisioned datasource + dashboards (Maestro RED, …)
│   ├── jmx-exporter         # config.yml + gitignored agent jar
│   └── prometheus           # prometheus.yml + alert.rules.yml
├── docs
│   └── assets               # logo + README screenshots
├── test
│   └── integration          # testcontainers suites (audit, auth, dbcheck, pgxadapter, users)
└── web                      # React + Vite console (maestro-web)
    └── src
        ├── api              # client, endpoints, types
        ├── auth             # AuthContext
        ├── components       # Layout, SchemaField, ui, …
        └── pages            # Dashboard, Connectors, Create/Detail, Audit, Users, …
```

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).

## License

MIT — see [LICENSE](LICENSE).
