# Development

Prerequisites: Go 1.26+, Docker with Compose v2, Node 20+ (only for web dev).

## Make targets

Run `make help` for the full list.

| Target | What it does |
|--------|--------------|
| `make build` | `go build -o bin/maestro ./cmd/maestro` |
| `make run` | build + run (exports `.env` via the Makefile) |
| `make test` | unit tests, `-race -cover` |
| `make test-integration` | testcontainers suites, needs a Docker daemon (`-tags integration`) |
| `make lint` / `make lint-fix` | `golangci-lint run ./...` (with autofix) |
| `make mocks` | regenerate `mockery` mocks (version pinned in `go.mod`) |
| `make validate-swagger` | `vacuum lint -r vacuum.yml api/swagger.yaml` |
| `make lint-actions` | `actionlint` for `.github/workflows` |
| `make lint-dockerfile` | `hadolint` on `cmd/maestro/Dockerfile` (via compose) |
| `make lint-trivy` | Trivy `HIGH,CRITICAL` filesystem scan |
| `make migrate-create seq=<name>` | new `golang-migrate` SQL version |
| `make migrate-up` / `make migrate-down` | apply / roll back migrations |
| `make docker-up` / `make docker-down` / `make ps` | compose lifecycle |
| `make jmx-exporter` | download the (gitignored) JMX agent jar with SHA256 check |
| `make alertmanager-config` | write the Slack webhook into gitignored `deploy/alertmanager/slack_api_url` |

Before pushing: `make lint test validate-swagger` (see
[CONTRIBUTING.md](../CONTRIBUTING.md)).

CI (`.github/workflows/`): `test.yml` (unit), `integration.yml`
(testcontainers), `lint.yml`, `security.yml` (Trivy), `docker.yml`.

## Conventions

- **Tests are table-driven** with full struct comparisons, parallel where
  possible; `goleak` verifies no goroutine leaks in suites.
- **Integration tests** live in `test/integration` (`audit`, `auth`, `dbcheck`,
  `pgxadapter`, `users`) and run against real PostgreSQL via testcontainers on
  an overlay network — not against mocks.
- **Slices, not layers:** `internal/features/<slice>/{repository,service,transport}`
  (`auth`, `users`, `connectors`, `audit`); shared kernel in `internal/core`
  (`domain`, `metrics`, `logger`, `security`, `transport`). The domain model
  lives in `internal/core/domain`, not inside slices.
- **Commits** follow Conventional Commits.
- **Mocks** are generated with mockery (`make mocks`) — do not hand-edit
  `mocks_test.go`.
- **API changes** must update `api/swagger.yaml` and pass `make validate-swagger`.
- **Web console** (`web/`, React + Vite): `npm run build` output is embedded
  into the Go binary via `internal/core/transport/webfs` (the Dockerfile copies
  `web/dist` there; the directory is otherwise gitignored at `web/dist`).
