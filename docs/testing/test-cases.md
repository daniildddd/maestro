# Test cases

Conventions: ID `TC-<area>-<NN>` (`CONN`, `AUTH`, `AUDIT`, `DBCHECK`,
`USERS`). Priority `smoke` ⊂ `critical path` ⊂ `extended` (see
[`test-plan.md`](test-plan.md)). `Covers` binds each case to the swagger
operation and the service/dbcheck unit — if the code moves, the case moves
with it. Preconditions assume the seed admin (`admin/admin12345`,
`docs/demo.md:10-27`) and the `public.orders` fixture
(`docs/demo.md:41-54`).

## Connectors

### TC-CONN-01: validate correct Postgres config

| Field | Value |
|-------|-------|
| Priority | critical path |
| Preconditions | stack up, fixture table exists |
| Steps | `POST /api/v1/connectors/validate` with the `pg-orders-cdc` payload from `docs/demo.md:76-96` |
| Expected | `200`, `valid: true`, every step `status: ok` (`config`, `connection`, `permissions`, `cdc`, `tables`) |
| Covers | `validateConnectorConfig` (`api/swagger.yaml:681`), `service/validate_connector.go` |

### TC-CONN-02: create connector from validated config

| Field | Value |
|-------|-------|
| Priority | critical path |
| Preconditions | TC-CONN-01 PASS |
| Steps | `POST /api/v1/connectors` (`docs/demo.md:107-117`) |
| Expected | `201`, connector appears on the Dashboard as `starting` → `running` (~7s live-polling) |
| Covers | `createConnector` (`api/swagger.yaml:606`), `service/create_connector.go` |

### TC-CONN-03: pause running connector

| Field | Value |
|-------|-------|
| Priority | critical path |
| Preconditions | TC-CONN-02 PASS, state `running` |
| Steps | `POST /api/v1/connectors/pg-orders-cdc/pause` with admin Bearer |
| Expected | `200`, state → `paused`; audit log has a `pause` entry with actor |
| Covers | `pauseConnector` (`api/swagger.yaml:909`), `service/pause.go` |

### TC-CONN-04: resume paused connector

| Field | Value |
|-------|-------|
| Priority | critical path |
| Preconditions | TC-CONN-03 PASS, state `paused` |
| Steps | `POST /api/v1/connectors/pg-orders-cdc/resume` with admin Bearer |
| Expected | `200`, state → `running` |
| Covers | `resumeConnector` (`api/swagger.yaml:933`), `service/resume.go` |

### TC-CONN-05: restart with failed-tasks-only flags

| Field | Value |
|-------|-------|
| Priority | critical path |
| Preconditions | TC-CONN-04 PASS |
| Steps | `POST .../restart?include_tasks=true&only_failed=true` (`docs/demo.md:143-144`) |
| Expected | `200`; connector accepts the restart (on a healthy connector without failed tasks the call succeeds but may change nothing observable) |
| Covers | `restartConnector` (`api/swagger.yaml:957`), `service/restart.go` |

### TC-CONN-06: update connector config

| Field | Value |
|-------|-------|
| Priority | critical path |
| Preconditions | TC-CONN-04 PASS |
| Steps | `PUT /api/v1/connectors/pg-orders-cdc` with changed `table.include.list` |
| Expected | `200`, new config persisted; audit log stores before/after snapshots |
| Covers | `updateConnector` (`api/swagger.yaml:829`) |

### TC-CONN-07: delete connector

| Field | Value |
|-------|-------|
| Priority | critical path |
| Preconditions | TC-CONN-02 PASS |
| Steps | `DELETE /api/v1/connectors/pg-orders-cdc` |
| Expected | `204`, connector gone from `GET /api/v1/connectors`; audit entry present |
| Covers | `deleteConnector` (`api/swagger.yaml:889`), `service/delete.go` |

### TC-CONN-08: validate config with unreachable database

| Field | Value |
|-------|-------|
| Priority | extended |
| Preconditions | stack up |
| Steps | `POST /api/v1/connectors/validate` with wrong `database.hostname` |
| Expected | `400` `CONNECTOR_CONFIG_INVALID`, `valid: false`; connection step reports `severity: error` with `field: database.hostname` (`dbcheck/postgres.go:80-90`) — no `fix_hint` in this branch |
| Covers | `validateConnectorConfig` (`api/swagger.yaml:681`), `dbcheck/postgres.go` |

## Auth

### TC-AUTH-01: login returns token pair

| Field | Value |
|-------|-------|
| Priority | smoke |
| Preconditions | seed admin exists |
| Steps | `POST /api/v1/login` `{"username":"admin","password":"admin12345"}` |
| Expected | `200` with `access_token`; refresh cookie set |
| Covers | `login` (`api/swagger.yaml:49`) |

### TC-AUTH-02: wrong password rejected

| Field | Value |
|-------|-------|
| Priority | extended |
| Preconditions | seed admin exists |
| Steps | `POST /api/v1/login` with wrong password |
| Expected | `400` `INVALID_CREDENTIALS`, no tokens issued |
| Covers | `login` (`api/swagger.yaml:49`) |

### TC-AUTH-03: user role cannot manage users

| Field | Value |
|-------|-------|
| Priority | extended |
| Preconditions | non-admin user exists |
| Steps | `POST /api/v1/users` with user Bearer |
| Expected | `403`; admin Bearer on the same payload → `201` |
| Covers | `createUser` (`api/swagger.yaml:293`) |

## Audit

### TC-AUDIT-01: connector pause leaves audit trail

| Field | Value |
|-------|-------|
| Priority | critical path |
| Preconditions | TC-CONN-03 executed |
| Steps | `GET /api/v1/audit-logs?action=connector.paused` and select the entry for `pg-orders-cdc` (no connector filter in `getAuditLogs`, `api/swagger.yaml:462-480`) |
| Expected | `200`, `pause` entry with actor, timestamp, `StateAfter` (pause writes no `StateBefore`, `service/pause.go:47`) |
| Covers | `getAuditLogs` (`api/swagger.yaml:466`) |

### TC-AUDIT-02: audit entry detail readable

| Field | Value |
|-------|-------|
| Priority | extended |
| Preconditions | TC-AUDIT-01 PASS |
| Steps | `GET /api/v1/audit-logs/{id}` for the pause entry |
| Expected | `200`, `StateAfter` snapshot present (before/after pair exists only on update entries, `service/update_connector.go:55-56`) |
| Covers | `getAuditLogById` (`api/swagger.yaml:511`) |

## Dbcheck

### TC-DBCHECK-01: missing table is an error with table pointer

| Field | Value |
|-------|-------|
| Priority | extended |
| Preconditions | stack up, `public.orders` dropped |
| Steps | validate config with `table.include.list=public.orders` |
| Expected | `valid: false`, failing check `severity: error` with `table: public.orders` (`dbcheck/tables.go:129-131`) — no `fix_hint` in this branch |
| Covers | `dbcheck/tables.go`, `ValidationCheck` (`api/swagger.yaml:2236`) |

### TC-DBCHECK-02: non-logical wal_level is an error

| Field | Value |
|-------|-------|
| Priority | extended |
| Preconditions | stack up (or a scratch Postgres with `wal_level=replica`) |
| Steps | validate a correct-looking config against that server |
| Expected | `valid: false`, `severity: error`, `fix_hint` mentions `wal_level` |
| Covers | `dbcheck/cdc.go:62-64` |

### TC-DBCHECK-03: table without PK degrades the verdict

| Field | Value |
|-------|-------|
| Priority | extended |
| Preconditions | table with default replica identity, no PK and no usable unique index |
| Steps | validate config including that table |
| Expected | step `status: warning` with `fix_hint` naming `ADD PRIMARY KEY` or `REPLICA IDENTITY FULL` (`dbcheck/tables.go`, default branch) |
| Covers | `dbcheck/tables.go` |

## Users

### TC-USERS-01: admin creates a user

| Field | Value |
|-------|-------|
| Priority | extended |
| Preconditions | admin Bearer |
| Steps | `POST /api/v1/users` with unique username, role `user` |
| Expected | `201`; `GET /api/v1/users/{id}` returns it |
| Covers | `createUser` (`api/swagger.yaml:293`), `getUserById` (`:330`) |

### TC-USERS-02: admin deletes a user

| Field | Value |
|-------|-------|
| Priority | extended |
| Preconditions | TC-USERS-01 PASS |
| Steps | `DELETE /api/v1/users/{id}` |
| Expected | `204`; subsequent `GET` → `404` |
| Covers | `deleteUser` (`api/swagger.yaml:417`) |
