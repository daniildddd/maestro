import { useCallback, useEffect, useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { Activity, AlertTriangle, Layers, PauseCircle, PlayCircle, ScrollText } from "lucide-react";
import * as api from "../api/endpoints";
import type { AuditLogItem } from "../api/types";
import { useAuth } from "../auth/AuthContext";
import { LiveToggle } from "../components/LiveToggle";
import { usePolling } from "../components/usePolling";
import { Button, EmptyState, ErrorBanner, Spinner, fmtDate } from "../components/ui";

interface Counters {
  total: number;
  running: number;
  paused: number;
  failed: number;
  starting: number;
}

const EMPTY: Counters = { total: 0, running: 0, paused: 0, failed: 0, starting: 0 };

export function DashboardPage() {
  const navigate = useNavigate();
  const { user } = useAuth();
  const isAdmin = user?.role === "admin";

  const [counters, setCounters] = useState<Counters>(EMPTY);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<unknown>(null);

  const [recent, setRecent] = useState<AuditLogItem[]>([]);
  const [recentError, setRecentError] = useState<unknown>(null);

  const load = useCallback(async () => {
    setError(null);
    try {
      const res = await api.getConnectors({ page: 1, limit: 100 });
      const next: Counters = { ...EMPTY, total: res.data?.length ?? 0 };
      for (const c of res.data ?? []) {
        if (c.status === "running") next.running++;
        else if (c.status === "paused") next.paused++;
        else if (c.status === "failed") next.failed++;
        else next.starting++;
      }
      setCounters(next);
    } catch (err) {
      setError(err);
    } finally {
      setLoading(false);
    }
  }, []);

  const loadRecent = useCallback(async () => {
    try {
      const res = await api.getAuditLogs({ page: 1, limit: 7 });
      setRecent(res.data ?? []);
      setRecentError(null);
    } catch (err) {
      setRecentError(err);
    }
  }, []);

  const { polling, setPolling } = usePolling(load, 7000);

  useEffect(() => {
    load();
  }, [load]);

  useEffect(() => {
    if (isAdmin) loadRecent();
  }, [isAdmin, loadRecent]);

  return (
    <div>
      <div className="page-head">
        <div>
          <h1 className="page-title">Dashboard</h1>
          <p className="page-sub">Overview of your connectors</p>
        </div>
        <LiveToggle active={polling} onToggle={() => setPolling((v) => !v)} />
      </div>

      {loading ? (
        <Spinner />
      ) : error ? (
        <ErrorBanner error={error} onRetry={load} />
      ) : (
        <div className="dash-cards">
          <Link to="/connectors" className="dash-card">
            <Layers size={18} aria-hidden="true" />
            <div className="dash-num">{counters.total}</div>
            <div className="dash-label">Total</div>
          </Link>
          <button type="button" className="dash-card" onClick={() => navigate("/connectors?status=running")}>
            <PlayCircle size={18} className="c-green" aria-hidden="true" />
            <div className="dash-num">{counters.running}</div>
            <div className="dash-label">Running</div>
          </button>
          <button type="button" className="dash-card" onClick={() => navigate("/connectors?status=paused")}>
            <PauseCircle size={18} className="c-amber" aria-hidden="true" />
            <div className="dash-num">{counters.paused}</div>
            <div className="dash-label">Paused</div>
          </button>
          <button type="button" className="dash-card" onClick={() => navigate("/connectors?status=failed")}>
            <AlertTriangle size={18} className="c-red" aria-hidden="true" />
            <div className="dash-num">{counters.failed}</div>
            <div className="dash-label">Failed</div>
          </button>
        </div>
      )}

      {isAdmin && (
        <div className="card" style={{ marginTop: 20 }}>
          <div className="card-head">
            <h2 className="section-title">
              <Activity size={15} aria-hidden="true" /> Recent activity
            </h2>
            <Link to="/audit">
              <Button variant="ghost" className="btn-sm" type="button">
                <ScrollText size={14} aria-hidden="true" /> Open audit log
              </Button>
            </Link>
          </div>
          {recentError ? (
            <div style={{ padding: "0 16px 14px" }}>
              <ErrorBanner error={recentError} onRetry={loadRecent} />
            </div>
          ) : recent.length === 0 ? (
            <EmptyState title="No recent activity" />
          ) : (
            <table className="table">
              <thead>
                <tr>
                  <th>Time</th>
                  <th>Action</th>
                  <th>Actor</th>
                  <th>Subject</th>
                </tr>
              </thead>
              <tbody>
                {recent.map((e) => (
                  <tr key={e.id} className="rowlink" onClick={() => navigate("/audit")}>
                    <td className="muted">{fmtDate(e.created_at)}</td>
                    <td className="mono">{e.action}</td>
                    <td className="mono">{e.actor?.username || "—"}</td>
                    <td className="mono">{e.subject.name || e.subject.id || "—"}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>
      )}
    </div>
  );
}