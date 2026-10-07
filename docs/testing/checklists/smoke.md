# Smoke (5 min, after every `make docker-up`)

Covers [`TC-CONN-02`](../test-cases.md#tc-conn-02-create-connector-from-validated-config)
and [`TC-AUTH-01`](../test-cases.md#tc-auth-01-login-returns-token-pair) at
the shallowest level: the stack answers and auth works. Anything failing
here blocks the critical path — do not proceed, fix the contour first.

- [ ] `GET http://127.0.0.1:8080/healthz → 204`
- [ ] `docker compose ps` — all services `Up`
- [ ] login `admin/admin12345 → 200` with `access_token`
  (`docs/demo.md:70-73`)
- [ ] `GET /api/v1/connectors → 200` (array, may be empty)
- [ ] `GET /api/v1/connector-plugins → 200`
  (`api/swagger.yaml:1032`; Debezium plugin listed)
- [ ] `GET /api/v1/smt-plugins → 200` (`api/swagger.yaml:1100`)
- [ ] web console `http://127.0.0.1:8080` loads, login works
- [ ] Grafana `http://127.0.0.1:3000` shows `Maestro RED` dashboard
- [ ] `GET /api/v1/audit-logs → 200`
- [ ] `GET /api/v1/users/me → 200` with admin Bearer
