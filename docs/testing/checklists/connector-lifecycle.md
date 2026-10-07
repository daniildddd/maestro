# Connector lifecycle (critical path)

Full pass mirrors `docs/demo.md:60-152`. Every step must leave an audit
record — a state change without an audit entry is a FAIL of that item, not
a note. Cases: [`TC-CONN-01`–`TC-CONN-07`](../test-cases.md#tc-conn-01-validate-correct-postgres-config),
[`TC-AUDIT-01`](../test-cases.md#tc-audit-01-connector-pause-leaves-audit-trail).

Preconditions: smoke green, fixture table ready
(`docs/demo.md:41-54`), admin Bearer exported as `$TOKEN`
(`docs/demo.md:70-73`).

- [ ] validate correct config → `200`, `valid: true`, all steps `ok`
- [ ] create `pg-orders-cdc` → `201`, Dashboard `starting` → `running`
- [ ] `GET /api/v1/connectors/pg-orders-cdc → 200`, task state visible
- [ ] insert row into `public.orders` → event lands in
  `maestro-demo.public.orders` (`docs/demo.md:125-129`)
- [ ] pause → `200`, state `paused` + audit entry with actor
- [ ] resume → `200`, state `running` + audit entry
- [ ] restart `?include_tasks=true&only_failed=true` → `200` (+ audit entry)
- [ ] restart single task `POST .../tasks/0/restart` → `204`
  (`api/swagger.yaml:1011`)
- [ ] update config → `200`, audit holds before/after snapshots
- [ ] `GET /api/v1/audit-logs?action=connector.paused` shows the pause entry (no connector filter in `getAuditLogs`)
- [ ] `GET /api/v1/connector-plugins/{id}/schema → 200`
  (schema-driven editor source, `api/swagger.yaml:1057`)
- [ ] delete connector → gone from list + audit entry
- [ ] delete audit log entry (admin) → `204`
  (`api/swagger.yaml:549`)
- [ ] cleanup: `DROP PUBLICATION maestro_pub; DROP TABLE public.orders;`
  (`docs/demo.md:151-152`)
