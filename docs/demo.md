# Demo: Postgres → Kafka in 10 minutes

Goal: capture changes of `public.orders` in Postgres and see them land in
Kafka, entirely through Maestro.

Prerequisites: the stack is up ([Quick start](../README.md#quick-start)) and
you are logged into the web console as admin (see below if this is a fresh
install).

## Create the first admin user

There is no public self-registration: `POST /api/v1/users` requires the
`admin` role, so the very first user must be inserted directly. With the stack
running:

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

Sanity check (Connect is reachable through the Maestro API, no direct port):

```bash
curl -s -o /dev/null -w "%{http_code}\n" http://127.0.0.1:8080/healthz
docker compose ps
```

## Scenario

**0. Stack is up** ([Quick start](../README.md#quick-start)) and you are logged
into the web console as admin.

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
