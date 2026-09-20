import { useCallback, useEffect, useState } from "react";
import * as api from "../api/endpoints";
import type { AuditLogEntry, AuditLogItem } from "../api/types";
import { ApiError } from "../api/client";
import {
  Button,
  EmptyState,
  ErrorBanner,
  Modal,
  Paginator,
  Spinner,
  fmtDate,
  useConfirm,
  useToast,
} from "../components/ui";

const ACTION_LABELS: Record<string, string> = {
  "auth.login": "Login",
  "auth.logout": "Logout",
  "user.created": "User created",
  "user.updated": "User updated",
  "user.deleted": "User deleted",
  "user.password_changed": "Password changed",
  "connector.created": "Connector created",
  "connector.updated": "Connector updated",
  "connector.deleted": "Connector deleted",
  "connector.paused": "Connector paused",
  "connector.resumed": "Connector resumed",
  "connector.restarted": "Connector restarted",
  "connector.task_restarted": "Task restarted",
};

const KNOWN_ACTIONS = Object.keys(ACTION_LABELS);

function actionLabel(action: string): string {
  return ACTION_LABELS[action] ?? action;
}

function subjectLabel(e: AuditLogItem): string {
  if (e.subject.type === "connector") return e.subject.name || "—";
  return e.subject.name || e.subject.id || "—";
}

function actorLabel(e: AuditLogItem): string {
  if (!e.actor) return "—";
  return e.actor.username || e.actor.id;
}

// NOTE: /api/v1/audit-logs is documented in swagger but may be absent on older
// servers. This page degrades gracefully: if the server answers 404/405 we
// show a friendly "not available" state instead of a crash.
export function AuditPage() {
  const toast = useToast();
  const { confirm, confirmNode } = useConfirm();

  const [items, setItems] = useState<AuditLogItem[] | null>(null);
  const [meta, setMeta] = useState({ page: 1, limit: 20 });
  const [hasMore, setHasMore] = useState(false);
  const [page, setPage] = useState(1);
  const [limit, setLimit] = useState(20);
  const [dateFrom, setDateFrom] = useState("");
  const [dateTo, setDateTo] = useState("");
  const [action, setAction] = useState("");
  const [customAction, setCustomAction] = useState("");
  const [actor, setActor] = useState("");
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<unknown>(null);
  const [detail, setDetail] = useState<AuditLogEntry | null>(null);

  const effectiveAction = action === "__custom__" ? customAction.trim() : action;

  const load = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const res = await api.getAuditLogs({ page, limit, action: effectiveAction, actor });
      setItems(res.data ?? []);
      setMeta(res.meta);
      setHasMore(res.has_more);
    } catch (err) {
      setError(err);
      setItems([]);
    } finally {
      setLoading(false);
    }
  }, [page, limit, effectiveAction, actor]);

  useEffect(() => {
    load();
  }, [load]);

  const openDetail = async (id: string) => {
    try {
      setDetail(await api.getAuditLog(id));
    } catch (err) {
      toast("error", err instanceof Error ? err.message : "Failed to load entry");
    }
  };

  const onDelete = async (id: string) => {
    const ok = await confirm("Delete log entry", <>Audit entry will be permanently removed.</>, {
      danger: true,
      confirmText: "Delete",
    });
    if (!ok) return;
    try {
      await api.deleteAuditLog(id);
      toast("success", "Entry deleted");
      setDetail(null);
      await load();
    } catch (err) {
      toast("error", err instanceof Error ? err.message : "Delete failed");
    }
  };

  const notAvailable = error instanceof ApiError && (error.status === 404 || error.status === 405);

  // client-side date filter: the API exposes only action/actor filters, so the
  // date range trims the current page after loading
  const filteredItems = (items ?? []).filter((e) => {
    const t = new Date(e.created_at).getTime();
    if (Number.isNaN(t)) return true;
    if (dateFrom && t < new Date(dateFrom + "T00:00:00").getTime()) return false;
    if (dateTo && t > new Date(dateTo + "T23:59:59.999").getTime()) return false;
    return true;
  });

  return (
    <div>
      <div className="page-head">
        <div>
          <h1 className="page-title">Audit log</h1>
          <p className="page-sub">Who changed what, and when</p>
        </div>
      </div>

      <div className="toolbar">
        <select
          value={action}
          onChange={(e) => setAction(e.target.value)}
          style={{ maxWidth: 220 }}
          aria-label="Filter by action"
        >
          <option value="">All actions</option>
          {KNOWN_ACTIONS.map((a) => (
            <option key={a} value={a}>
              {actionLabel(a)} ({a})
            </option>
          ))}
          <option value="__custom__">Custom…</option>
        </select>
        {action === "__custom__" && (
          <input
            type="text"
            placeholder="Custom action (e.g. connector.validate)…"
            value={customAction}
            onChange={(e) => setCustomAction(e.target.value)}
            style={{ maxWidth: 240 }}
            aria-label="Custom action filter"
          />
        )}
        <input
          type="text"
          placeholder="Actor…"
          value={actor}
          onChange={(e) => setActor(e.target.value)}
          style={{ maxWidth: 200 }}
          aria-label="Filter by actor"
        />
        <input
          type="date"
          value={dateFrom}
          onChange={(e) => setDateFrom(e.target.value)}
          max={dateTo || undefined}
          aria-label="From date"
        />
        <input
          type="date"
          value={dateTo}
          onChange={(e) => setDateTo(e.target.value)}
          min={dateFrom || undefined}
          aria-label="To date"
        />
        {(dateFrom || dateTo) && (
          <Button variant="ghost" className="btn-sm" onClick={() => { setDateFrom(""); setDateTo(""); }} type="button">
            Clear dates
          </Button>
        )}
      </div>

      {notAvailable ? (
        <EmptyState
          title="Audit log is not available on this server"
          hint="The backend build does not expose /api/v1/audit-logs yet. The UI will pick it up automatically once deployed."
        />
      ) : (
        <>
          {error && !notAvailable ? <ErrorBanner error={error} onRetry={load} /> : null}
          <div className="card">
            {loading ? (
              <Spinner />
            ) : !items || items.length === 0 ? (
              <EmptyState title="No audit entries" />
            ) : filteredItems.length === 0 ? (
              <EmptyState title="No entries in the selected date range" hint="Adjust or clear the date filters." />
            ) : (
              <table className="table">
                <thead>
                  <tr>
                    <th>Time</th>
                    <th>Action</th>
                    <th>Outcome</th>
                    <th>Actor</th>
                    <th>Subject</th>
                    <th />
                  </tr>
                </thead>
                <tbody>
                  {filteredItems.map((e) => (
                    <tr key={e.id} className="rowlink" onClick={() => openDetail(e.id)}>
                      <td className="muted">{fmtDate(e.created_at)}</td>
                      <td className="mono">{actionLabel(e.action)}</td>
                      <td>
                        <span className={`badge badge-${e.outcome}`}>{e.outcome}</span>
                      </td>
                      <td className="mono">{actorLabel(e)}</td>
                      <td className="mono">{subjectLabel(e)}</td>
                      <td style={{ textAlign: "right" }}>
                        <Button
                          variant="ghost"
                          className="btn-sm"
                          type="button"
                          onClick={(ev) => {
                            ev.stopPropagation();
                            void onDelete(e.id);
                          }}
                        >
                          Delete
                        </Button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            )}
          </div>
          <Paginator page={meta.page} limit={limit} hasNext={hasMore} onPage={setPage} onLimit={setLimit} />
        </>
      )}

      {detail ? (
        <Modal title={actionLabel(detail.action)} onClose={() => setDetail(null)} width={680}>
          <dl className="kv" style={{ marginBottom: 12 }}>
            <dt>Time</dt>
            <dd>{fmtDate(detail.created_at)}</dd>
            <dt>Action</dt>
            <dd className="mono">{detail.action}</dd>
            <dt>Outcome</dt>
            <dd>
              <span className={`badge badge-${detail.outcome}`}>{detail.outcome}</span>
            </dd>
            {detail.failure_reason ? (
              <>
                <dt>Failure reason</dt>
                <dd className="mono">{detail.failure_reason}</dd>
              </>
            ) : null}
            <dt>Actor</dt>
            <dd className="mono">{detail.actor ? `${detail.actor.username} (${detail.actor.id})` : "—"}</dd>
            <dt>Subject</dt>
            <dd className="mono">
              {detail.subject.type}: {detail.subject.name || detail.subject.id || "—"}
            </dd>
            {detail.request ? (
              <>
                <dt>Request id</dt>
                <dd className="mono">{detail.request.id}</dd>
                <dt>IP</dt>
                <dd className="mono">{detail.request.ip}</dd>
                <dt>User agent</dt>
                <dd className="mono">{detail.request.user_agent}</dd>
              </>
            ) : null}
          </dl>
          <h3 className="section-title">State before</h3>
          <pre className="trace">{detail.state_before ? JSON.stringify(detail.state_before, null, 2) : "null"}</pre>
          <h3 className="section-title">State after</h3>
          <pre className="trace">{detail.state_after ? JSON.stringify(detail.state_after, null, 2) : "null"}</pre>
        </Modal>
      ) : null}
      {confirmNode}
    </div>
  );
}
