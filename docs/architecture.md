# Architecture

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

Only Maestro (`:8080`) and Grafana (`:3000`, embedded in the Metrics page via
iframe) publish host ports. Everything else lives inside the compose network.

Request flow for a typical "create connector":

1. Web console (or `curl`) calls `POST /api/v1/connectors/validate` with
   `plugin_type + name + config`.
2. Maestro validates types/required fields against the Connect plugin schema,
   then runs Postgres `dbcheck` (connection → permissions → CDC → tables).
3. On `valid: true` the UI enables **Create** → `POST /api/v1/connectors`
   proxies to Kafka Connect, writes an audit entry with the config snapshot.
4. The background collector exports connector/task states as
   `maestro_connector_info/tasks` for Prometheus/Grafana/alerts.

API contract: [`api/swagger.yaml`](../api/swagger.yaml) (OpenAPI, `vacuum`-linted).

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
│       ├── audit            # change history: who did what, with before/after snapshots
│       ├── auth             # JWT login / refresh / logout and expired-token cleanup
│       ├── connectors       # lifecycle: validation, dbcheck, Connect proxy, state collector
│       └── users            # user accounts: roles, passwords, profiles
├── migrations               # golang-migrate SQL versions
├── deploy                   # infra as code
│   ├── alertmanager         # alertmanager.yml + gitignored slack_api_url
│   ├── grafana              # provisioned datasource + dashboards (Maestro RED, …)
│   ├── jmx-exporter         # config.yml + gitignored agent jar
│   └── prometheus           # prometheus.yml + alert.rules.yml
├── docs
│   ├── assets               # logo + README screenshots
│   ├── architecture.md      # this file: service map, request flow, project structure
│   ├── configuration.md     # full env reference
│   ├── demo.md              # first admin user + end-to-end Postgres → Kafka scenario
│   ├── development.md       # make targets, conventions, CI
│   └── monitoring.md        # metrics, alerts, dashboards
├── test
│   └── integration          # testcontainers suites (audit, auth, dbcheck, pgxadapter, users)
└── web                      # React + Vite console (maestro-web)
    └── src
        ├── api              # client, endpoints, types
        ├── auth             # AuthContext
        ├── components       # Layout, SchemaField, ui, …
        └── pages            # Dashboard, Connectors, Create/Detail, Audit, Users, …
```
