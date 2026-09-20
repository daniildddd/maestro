import { useCallback, useEffect, useState } from "react";
import { Link, useNavigate, useSearchParams } from "react-router-dom";
import { ChevronDown, ChevronRight, Plus, RotateCcw } from "lucide-react";
import * as api from "../api/endpoints";
import type { ConnectorDetail, ConnectorListItem } from "../api/types";
import { ActionMenu } from "../components/ActionMenu";
import { LiveToggle } from "../components/LiveToggle";
import { usePolling } from "../components/usePolling";
import {
  Button,
  EmptyState,
  ErrorBanner,
  Paginator,
  Spinner,
  TaskStateBadge,
  shortClass,
  useConfirm,
  useDebounced,
  useToast,
} from "../components/ui";

const STATUSES = ["", "running", "paused", "failed", "starting"] as const;

type SortKey = "name" | "plugin" | "status" | "tasks";
type SortDir = "asc" | "desc";

const STATUS_ORDER: Record<string, number> = { running: 0, starting: 1, paused: 2, failed: 3 };

function compareConnectors(a: ConnectorListItem, b: ConnectorListItem, key: SortKey, dir: SortDir): number {
  let cmp: number;
  if (key === "tasks") {
    cmp = a.tasks_count - b.tasks_count;
  } else if (key === "status") {
    cmp = (STATUS_ORDER[a.status] ?? 9) - (STATUS_ORDER[b.status] ?? 9) || a.status.localeCompare(b.status);
  } else {
    const av = key === "name" ? a.name : (a.plugin_type ?? "");
    const bv = key === "name" ? b.name : (b.plugin_type ?? "");
    cmp = av.localeCompare(bv);
  }
  // name tie-break keeps row order stable across live refreshes
  if (cmp === 0) cmp = a.name.localeCompare(b.name);
  return dir === "asc" ? cmp : -cmp;
}

export function ConnectorsPage() {
  const [searchParams, setSearchParams] = useSearchParams();
  const status = searchParams.get("status") ?? "";
  const [statusInput, setStatusInput] = useState(status);
  const [search, setSearch] = useState("");
  const debouncedSearch = useDebounced(search);
  const [items, setItems] = useState<ConnectorListItem[]>([]);
  const [meta, setMeta] = useState({ page: 1, limit: 20 });
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [limit, setLimit] = useState(20);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<unknown>(null);
  const [sortKey, setSortKey] = useState<SortKey>("name");
  const [sortDir, setSortDir] = useState<SortDir>("asc");

  const load = useCallback(async (opts?: { silent?: boolean }) => {
    if (!opts?.silent) setLoading(true);
    setError(null);
    try {
      const res = await api.getConnectors({ page, limit, status, search: debouncedSearch });
      setItems(res.data ?? []);
      setMeta(res.meta);
      setTotal(res.total ?? 0);
    } catch (err) {
      setError(err);
    } finally {
      if (!opts?.silent) setLoading(false);
    }
  }, [page, limit, status, debouncedSearch]);

  useEffect(() => {
    load();
  }, [load]);

  // background refresh must not flash the spinner or collapse expanded rows
  const { polling, setPolling } = usePolling(() => void load({ silent: true }), 7000);

  // keep the select in sync when the status filter is changed from the Dashboard
  useEffect(() => {
    setStatusInput(status);
  }, [status]);

  // reset to first page when filters change
  useEffect(() => {
    setPage(1);
  }, [status, debouncedSearch]);

  const onStatusChange = (next: string) => {
    setStatusInput(next);
    const params = new URLSearchParams(searchParams);
    if (next) params.set("status", next);
    else params.delete("status");
    setSearchParams(params, { replace: true });
  };

  return (
    <div>
      <div className="page-head">
        <div>
          <h1 className="page-title">Connectors</h1>
          <p className="page-sub">Debezium connectors managed through Kafka Connect</p>
        </div>
        <div className="head-actions">
          <LiveToggle active={polling} onToggle={() => setPolling((v) => !v)} />
          <Link to="/connectors/new">
            <Button>
              <Plus size={15} aria-hidden="true" /> Create connector
            </Button>
          </Link>
        </div>
      </div>

      <div className="toolbar">
        <input
          type="text"
          placeholder="Search by name…"
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          style={{ maxWidth: 280 }}
          aria-label="Search connectors"
        />
        <select
          value={statusInput}
          onChange={(e) => onStatusChange(e.target.value)}
          style={{ maxWidth: 180 }}
          aria-label="Filter by status"
        >
          {STATUSES.map((s) => (
            <option key={s} value={s}>
              {s === "" ? "All statuses" : s}
            </option>
          ))}
        </select>
      </div>

      {error ? <ErrorBanner error={error} onRetry={load} /> : null}

      <div className="card">
        {loading ? (
          <Spinner />
        ) : items.length === 0 ? (
          <EmptyState
            title="No connectors found"
            hint={search || status ? "Try adjusting the filters." : "Create your first connector to get started."}
          />
        ) : (
          <table className="table">
            <thead>
              <tr>
                <th style={{ width: 36 }} />
                {(["name", "plugin", "status", "tasks"] as const).map((k) => (
                  <SortableTh
                    key={k}
                    label={k === "name" ? "Name" : k === "plugin" ? "Plugin" : k === "status" ? "Status" : "Tasks"}
                    sortKey={k}
                    activeKey={sortKey}
                    dir={sortDir}
                    onSort={(key, d) => {
                      setSortKey(key);
                      setSortDir(d);
                    }}
                  />
                ))}
                <th style={{ width: 48 }} />
              </tr>
            </thead>
            <tbody>
              {[...items]
                .sort((a, b) => compareConnectors(a, b, sortKey, sortDir))
                .map((c) => (
                  <ConnectorRow key={c.name} item={c} onChanged={load} />
                ))}
            </tbody>
          </table>
        )}
      </div>

      <Paginator page={meta.page} limit={limit} hasNext={page * meta.limit < total} onPage={setPage} onLimit={setLimit} />
    </div>
  );
}

function ConnectorRow({ item, onChanged }: { item: ConnectorListItem; onChanged: () => void }) {
  const navigate = useNavigate();
  const toast = useToast();
  const { confirm, confirmNode } = useConfirm();
  const [expanded, setExpanded] = useState(false);
  const [detail, setDetail] = useState<ConnectorDetail | null>(null);
  const [detailLoading, setDetailLoading] = useState(false);
  const [detailError, setDetailError] = useState<unknown>(null);
  const [busy, setBusy] = useState<string | null>(null);

  const loadDetail = useCallback(async () => {
    setDetailLoading(true);
    setDetailError(null);
    try {
      setDetail(await api.getConnector(item.name));
    } catch (err) {
      setDetailError(err);
    } finally {
      setDetailLoading(false);
    }
  }, [item.name]);

  useEffect(() => {
    if (expanded && !detail && !detailLoading && !detailError) loadDetail();
  }, [expanded, detail, detailLoading, detailError, loadDetail]);

  const runAction = async (key: string, fn: () => Promise<unknown>, okMessage: string) => {
    setBusy(key);
    try {
      await fn();
      toast("success", okMessage);
      onChanged();
      if (expanded) await loadDetail();
      return true;
    } catch (err) {
      toast("error", err instanceof Error ? err.message : "Action failed");
      return false;
    } finally {
      setBusy(null);
    }
  };

  const onPause = () => runAction("pause", () => api.pauseConnector(item.name), "Connector paused");
  const onResume = () => runAction("resume", () => api.resumeConnector(item.name), "Connector resumed");

  const onDelete = async () => {
    const ok = await confirm(
      "Delete connector",
      <>
        Connector <strong className="mono">{item.name}</strong> and all its tasks will be removed from Kafka Connect.
        This cannot be undone.
      </>,
      { danger: true, confirmText: "Delete" },
    );
    if (!ok) return;
    await runAction("delete", () => api.deleteConnector(item.name), "Connector deleted");
  };

  const restartTask = async (taskId: number) => {
    setBusy("task-" + taskId);
    try {
      await api.restartTask(item.name, taskId);
      toast("success", "Task #" + taskId + " restarted");
      await loadDetail();
    } catch (err) {
      toast("error", err instanceof Error ? err.message : "Restart failed");
    } finally {
      setBusy(null);
    }
  };

  return (
    <>
      <tr className="rowlink" onClick={() => navigate("/connectors/" + encodeURIComponent(item.name))}>
        <td>
          <button
            type="button"
            className="icon-btn expand-btn"
            aria-expanded={expanded}
            aria-label={expanded ? "Collapse tasks" : "Expand tasks"}
            onClick={(e) => {
              e.stopPropagation();
              setExpanded((v) => !v);
            }}
          >
            {expanded ? <ChevronDown size={15} /> : <ChevronRight size={15} />}
          </button>
        </td>
        <td className="mono">{item.name}</td>
        <td className="mono muted">{shortClass(item.plugin_type)}</td>
        <td>
          <span className={"badge badge-" + item.status}>
            <span className="dot" />
            {item.status}
          </span>
        </td>
        <td>{item.tasks_count}</td>
        <td onClick={(e) => e.stopPropagation()}>
          <ActionMenu
            label={"Actions for " + item.name}
            items={[
              item.status !== "paused"
                ? { label: "Pause", onSelect: onPause, disabled: busy === "pause" }
                : { label: "Resume", onSelect: onResume, disabled: busy === "resume" },
              {
                label: "Restart…",
                onSelect: () => navigate("/connectors/" + encodeURIComponent(item.name) + "?restart=1"),
              },
              {
                label: "Edit config",
                onSelect: () => navigate("/connectors/" + encodeURIComponent(item.name) + "?edit=1"),
              },
              { label: "Delete", onSelect: onDelete, danger: true, disabled: busy === "delete" },
            ]}
          />
        </td>
      </tr>
      {expanded && (
        <tr className="expand-row">
          <td colSpan={6}>
            {detailLoading && !detail ? (
              <Spinner label="Loading tasks…" />
            ) : detailError ? (
              <ErrorBanner error={detailError} onRetry={loadDetail} />
            ) : detail && detail.tasks && detail.tasks.length > 0 ? (
              <table className="table inner-table">
                <thead>
                  <tr>
                    <th>Task</th>
                    <th>State</th>
                    <th>Worker</th>
                    <th style={{ width: 120 }} />
                  </tr>
                </thead>
                <tbody>
                  {detail.tasks.map((t) => (
                    <tr key={t.id}>
                      <td>#{t.id}</td>
                      <td>
                        <TaskStateBadge state={t.state} />
                      </td>
                      <td className="mono muted">{t.worker_id || "—"}</td>
                      <td style={{ textAlign: "right" }}>
                        <Button
                          variant="ghost"
                          className="btn-sm"
                          loading={busy === "task-" + t.id}
                          onClick={() => restartTask(t.id)}
                          title={t.state.toLowerCase() === "running" ? "Restart is only needed for failed or paused tasks" : undefined}
                          type="button"
                        >
                          <RotateCcw size={13} aria-hidden="true" /> Restart
                        </Button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            ) : (
              <p className="muted" style={{ margin: "4px 0 6px" }}>
                No tasks yet — tasks appear when the connector starts.
              </p>
            )}
          </td>
        </tr>
      )}
      {confirmNode}
    </>
  );
}

function SortableTh({
  label,
  sortKey,
  activeKey,
  dir,
  onSort,
}: {
  label: string;
  sortKey: SortKey;
  activeKey: SortKey;
  dir: SortDir;
  onSort: (key: SortKey, dir: SortDir) => void;
}) {
  const active = sortKey === activeKey;
  const nextDir: SortDir = active && dir === "asc" ? "desc" : "asc";
  return (
    <th>
      <button
        type="button"
        className={"sort-btn" + (active ? " active" : "")}
        onClick={() => onSort(sortKey, nextDir)}
        aria-label={"Sort by " + label + (active ? ", currently " + dir : "")}
      >
        {label}
        <span className="sort-arrow" aria-hidden="true">
          {active ? (dir === "asc" ? "↑" : "↓") : "↕"}
        </span>
      </button>
    </th>
  );
}