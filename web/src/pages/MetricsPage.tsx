import { useState } from "react";
import { ExternalLink, RefreshCw } from "lucide-react";
import { Button } from "../components/ui";

// Grafana runs as a sidecar service (docker-compose, :3000). The dashboard is
// provisioned read-only; anonymous Viewer access lets the iframe render
// without a separate login. Full kiosk (?kiosk) hides both the sidebar and
// the top nav in Grafana 12; our own controls drive from/to/refresh.
const GRAFANA_BASE = "http://localhost:3000";
const DASH_URL = GRAFANA_BASE + "/d/maestro-red/maestro-red";

const RANGES = [
  { label: "1h", value: "now-1h" },
  { label: "6h", value: "now-6h" },
  { label: "24h", value: "now-24h" },
  { label: "7d", value: "now-7d" },
];

export function MetricsPage() {
  const [range, setRange] = useState("now-1h");
  const [refresh, setRefresh] = useState("15s");
  const [iframeKey, setIframeKey] = useState(0);

  const src = `${DASH_URL}?kiosk&from=${range}&to=now&refresh=${refresh}`;

  return (
    <div className="metrics-page">
      <div className="page-head">
        <div>
          <h1 className="page-title">Metrics</h1>
          <p className="page-sub">RED metrics and connector health, powered by Grafana</p>
        </div>
        <div className="metrics-tools">
          <div className="range-toggle" role="group" aria-label="Time range">
            {RANGES.map((r) => (
              <button
                key={r.value}
                type="button"
                className={"range-btn" + (range === r.value ? " active" : "")}
                onClick={() => setRange(r.value)}
              >
                {r.label}
              </button>
            ))}
          </div>
          <select
            value={refresh}
            onChange={(e) => setRefresh(e.target.value)}
            aria-label="Auto refresh"
            className="refresh-select"
          >
            {["5s", "15s", "30s", "1m", "off"].map((v) => (
              <option key={v} value={v}>
                {v === "off" ? "no auto-refresh" : "refresh " + v}
              </option>
            ))}
          </select>
          <Button variant="ghost" className="btn-sm" onClick={() => setIframeKey((k) => k + 1)} type="button">
            <RefreshCw size={14} aria-hidden="true" /> Reload
          </Button>
          <a href={DASH_URL} target="_blank" rel="noreferrer" className="external-link">
            <Button variant="ghost" className="btn-sm" type="button">
              <ExternalLink size={14} aria-hidden="true" /> Open in Grafana
            </Button>
          </a>
        </div>
      </div>

      <div className="card metrics-frame-card">
        <iframe
          key={iframeKey}
          className="metrics-frame"
          src={src}
          title="Grafana — Maestro RED dashboard"
        />
      </div>
    </div>
  );
}