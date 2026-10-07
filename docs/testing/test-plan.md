# Test plan: integration contour

Scope: manual verification of the Maestro stack on a live contour
(Postgres + Kafka + Connect + app). Automated unit and integration
coverage lives in code — this plan covers what only a running stack shows.

## Scope

In:

- connector lifecycle via the API (`POST /api/v1/connectors`,
  `api/swagger.yaml:565`), config validation (`.../validate`, `:677`),
  pause/resume/restart (`:905`-`:1011`), task restart (`:1011`);
- `dbcheck` report: per-step `status` with `ok / warning / error / skipped` checks,
  `fix_hint`, `field` and `table` pointers (`api/swagger.yaml:2236-2255`);
- authN/authZ on the contour: login, RBAC (`admin` vs `user`), audit trail
  of connector operations.

Out:

- Kafka internals, Debezium snapshots and exactly-once semantics;
- web console E2E (covered by QA-agents with Playwright);
- load / soak / chaos testing.

## Environment

`make docker-up` (service versions in `README.md:125-140`: Postgres 17
`wal_level=logical`, Debezium 3.5.2, Kafka 4.3.1), then `make migrate-up`.
First admin per `docs/demo.md:10-27` (direct SQL insert, `admin/admin12345`).

Automated baseline before any manual run:

| Command | What it proves |
|---------|----------------|
| `make test` | unit tests, `-race -cover` (`Makefile:79-80`) |
| `make test-integration` | testcontainers suites (`audit`, `auth`, `dbcheck`, `pgxadapter`, `users`), `-tags integration` (`Makefile:82-83`) |
| `make lint test validate-swagger` | pre-push gate (`docs/development.md:27-28`) |

CI mirrors this: `test.yml` (unit), `integration.yml` (testcontainers),
`lint.yml`, `security.yml`, `docker.yml` (`.github/workflows/`).

## Levels

1. **Smoke** (`checklists/smoke.md`) — ~5 min after every `make docker-up`.
   Stack answers, auth works, plugin list non-empty, Grafana reachable.
2. **Critical path** (`checklists/connector-lifecycle.md`) — validate →
   create → running → pause → resume → restart → update → delete, with an
   audit record on every step. Must pass before any demo or release.
3. **Extended** (`checklists/dbcheck.md` + `test-cases.md`) — negative
   `dbcheck` verdicts, RBAC denials, audit filters. Run before releases.

Formal cases live in [`test-cases.md`](test-cases.md): `TC-<area>-<NN>`
(`CONN`, `AUTH`, `AUDIT`, `DBCHECK`, `USERS`), priority
`smoke` ⊂ `critical path` ⊂ `extended`.

## Entry/exit criteria

Entry: `GET /healthz → 204`
(`internal/core/transport/health/handler.go:44-45`), `make test` green.

Exit: smoke 100% PASS, critical path 100% PASS, extended ≥ 90% PASS.
Failures are filed with: failing checklist item or `TC-*` id, request +
response (or UI step), connector state, and the audit-log entry if present.
