# Dbcheck (extended)

Where the report carries a `fix_hint`, it must come with a `field` or
`table` pointer (`api/swagger.yaml:2236-2255`) — but not every branch has a
hint: connection failures and missing tables report `error` without one
(`dbcheck/postgres.go:80-90`, `dbcheck/tables.go:129-131`). Automated mirror:
`test/integration/dbcheck/dbcheck_test.go` (run with `make test-integration`).
Cases: [`TC-DBCHECK-01`–`TC-DBCHECK-03`](../test-cases.md#tc-dbcheck-01-missing-table-is-an-error-with-table-pointer),
[`TC-CONN-08`](../test-cases.md#tc-conn-08-validate-config-with-unreachable-database).

Run each item as `POST /api/v1/connectors/validate`
(`api/swagger.yaml:677`) with the fixture mutated accordingly
(`docs/demo.md:76-96` for the base payload).

- [ ] correct config → `valid: true`, all steps `ok`
- [ ] wrong password → `400` `CONNECTOR_CONFIG_INVALID`, `valid: false`,
  connection step `severity: error` with `field: database.hostname` and no `fix_hint`
- [ ] unreachable host → `valid: false`, connection step `severity: error` with no `fix_hint` (host/port only inside the pgx message)
- [ ] missing table in `table.include.list` → `severity: error`,
  `table: public.orders`, no `fix_hint` in this branch
- [ ] table with default replica identity, no PK and no usable unique index → step `warning` with `fix_hint` naming `ADD PRIMARY KEY` or `REPLICA IDENTITY FULL`
- [ ] `wal_level ≠ logical` → `severity: error`, `fix_hint` mentions
  `wal_level` (`dbcheck/cdc.go:62-64`)
- [ ] no free replication slot / no WAL sender → `severity: error` with hint
- [ ] user without `SELECT` on the table → `severity: error`, hint names
  the missing privilege (`dbcheck/permissions.go`)
- [ ] publication missing the table → default `severity: warning` ("added automatically"); `error` only when `publication.autocreate.mode=disabled` (`dbcheck/tables.go:171-184`)
- [ ] `VALIDATE_DB_MAX_TABLE_CHECKS` exceeded → report truncated gracefully, no `500`
  (`dbcheck/config.go`)
- [ ] step timeout (`VALIDATE_DB_STEP_TIMEOUT`, default `10s`) →
  step `skipped`, not a hung request
- [ ] unknown `plugin_type` → `400` before any DB probe is attempted
