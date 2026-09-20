# Monitoring

Observability is provisioned from git — no clicking in UIs. Prometheus,
Alertmanager and Grafana configs live in [`deploy/`](../deploy/); dashboards
and datasource are mounted read-only into the containers.

## Metrics

### Maestro app (`maestro:9100`, internal listener)

Scraped by Prometheus as job `maestro`. The listener is `expose`d inside the
compose network only — never published to the host.

| Metric | Type | Labels | Meaning |
|--------|------|--------|---------|
| `http_server_requests_total` | counter | `method`, `route`, `status` | total HTTP requests (RED: rate) |
| `http_server_request_duration_seconds` | histogram | `method`, `route` | request latency, buckets from `METRICS_BUCKETS` (RED: errors/duration → P95) |
| `maestro_connector_info` | gauge | `connector`, `state` | `1` per connector in its current state (`running`, `paused`, `failed`, `starting`) |
| `maestro_connector_tasks` | gauge | `connector`, `state` | task count per connector/state |

The connector/task gauges are refreshed by the background collector every
`METRICS_COLLECTOR_INTERVAL` (`15s` default).

### Kafka Connect (`connect:8084`, JMX Exporter)

Scraped as job `kafka-connect`. The `connect` service runs with
`KAFKA_OPTS=-javaagent:/opt/jmx-exporter/jmx_prometheus_javaagent.jar=8084:...`,
using [`deploy/jmx-exporter/config.yml`](../deploy/jmx-exporter/config.yml).
Key series: `debezium_milliseconds_behind_source`, `debezium_connected`.
The agent jar itself is gitignored — fetch it with `make jmx-exporter`.

### Scrape targets (`deploy/prometheus/prometheus.yml`)

| Job | Target | What |
|-----|--------|------|
| `prometheus` | `localhost:9090` | self |
| `alertmanager` | `alertmanager:9093` | Alertmanager |
| `kafka-connect` | `connect:8084` | Debezium JMX |
| `maestro` | `maestro:9100` | app RED + connector states |

> The Prometheus (`:9090`) and Alertmanager (`:9093`) UIs are not published to
> the host. Running Maestro via `make run` outside Docker breaks the `maestro:9100`
> target (name only resolves inside the compose network). Point Prometheus at
> `host.docker.internal:9100` for local runs.

## Alerts (`deploy/prometheus/alert.rules.yml`)

Delivered to Slack via Alertmanager (`deploy/alertmanager/`). The webhook URL
is stored in the gitignored `deploy/alertmanager/slack_api_url` file — generate
it with `make alertmanager-config` (needs `ALERTMANAGER_SLACK_API_URL`).

| Alert | Severity | Expression | Fires when |
|-------|----------|------------|------------|
| `DebeziumLagHigh` | warning | `debezium_milliseconds_behind_source > 5000` for `1m` | connector lags the source |
| `DebeziumDisconnected` | critical | `debezium_connected == 0` for `1m` | connector lost the DB connection |
| `ConnectTargetDown` | critical | `up{job="kafka-connect"} == 0` for `1m` | JMX target unreachable (all `debezium_*` stale) |
| `MaestroDown` | critical | `up{job="maestro"} == 0` for `1m` | app metrics unreachable |
| `MaestroHighErrorRate` | critical | 5xx ratio `> 0.01` over `5m`, for `5m` | >1% of responses are 5xx |
| `FailedConnectors` | critical | `sum(maestro_connector_info{state="failed"}) > 0` for `1m` | any connector in `FAILED` |

## Dashboards (`deploy/grafana/`)

Provisioned on Grafana startup from `dashboards/` + `datasources/` (read-only
mounts). To change a dashboard, edit the JSON and restart Grafana.

| Dashboard | File | Content |
|-----------|------|---------|
| Maestro RED | `dashboards/Application/maestro-red.json` | request rate, 5xx rate, P95 overall and per route |
| Maestro Overview | `maestro-overview.json` | connector/task states |
| Kafka Connect Cluster | `maestro-cluster.json` | Connect workers, Debezium lag |
| Observability Health | `maestro-observability.json` | pipeline self-health |

Login with `GF_SECURITY_ADMIN_USER` / `GF_SECURITY_ADMIN_PASSWORD`
(compose defaults `admin`/`admin`).
