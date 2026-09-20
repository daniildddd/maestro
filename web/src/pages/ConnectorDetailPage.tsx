import { useCallback, useEffect, useMemo, useState } from "react";
import { useNavigate, useParams, useSearchParams } from "react-router-dom";
import { Copy } from "lucide-react";
import * as api from "../api/endpoints";
import type { ConnectorDetail, PluginSchema, TaskDetail } from "../api/types";
import { SchemaFieldInput } from "../components/SchemaField";
import { CheckCard } from "../components/CheckCard";
import { usePolling } from "../components/usePolling";
import {
  Button,
  EmptyState,
  ErrorBanner,
  Field,
  Modal,
  Spinner,
  StatusBadge,
  TaskStateBadge,
  shortClass,
  useConfirm,
  useToast,
} from "../components/ui";

// DTO keys (database_hostname) mapped back to raw Kafka Connect keys
const dtoToRaw: Record<string, string> = {
  database_hostname: "database.hostname",
  database_port: "database.port",
  database_user: "database.user",
  database_dbname: "database.dbname",
  plugin_name: "plugin.name",
};

export function ConnectorDetailPage() {
  const { name = "" } = useParams();
  const navigate = useNavigate();
  const [searchParams, setSearchParams] = useSearchParams();
  const toast = useToast();
  const { confirm, confirmNode } = useConfirm();

  const [connector, setConnector] = useState<ConnectorDetail | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<unknown>(null);

  const [actionBusy, setActionBusy] = useState<string | null>(null);
  const [restartOpen, setRestartOpen] = useState(false);
  const [editOpen, setEditOpen] = useState(false);

  const load = useCallback(async () => {
    setError(null);
    try {
      setConnector(await api.getConnector(name));
    } catch (err) {
      setError(err);
    } finally {
      setLoading(false);
    }
  }, [name]);

  useEffect(() => {
    load();
  }, [load]);

  // allow the connectors list menu to open these modals directly
  useEffect(() => {
    if (searchParams.get("restart")) setRestartOpen(true);
    if (searchParams.get("edit")) setEditOpen(true);
    if (searchParams.get("restart") || searchParams.get("edit")) {
      const params = new URLSearchParams(searchParams);
      params.delete("restart");
      params.delete("edit");
      setSearchParams(params, { replace: true });
    }
  }, [searchParams, setSearchParams]);

  // live status updates; suspended while a modal is open to avoid flicker
  usePolling(load, 10000, !restartOpen && !editOpen);

  const runAction = async (key: string, fn: () => Promise<unknown>, okMessage: string) => {
    setActionBusy(key);
    setError(null);
    try {
      await fn();
      toast("success", okMessage);
      await load();
      return true;
    } catch (err) {
      toast("error", err instanceof Error ? err.message : "Action failed");
      return false;
    } finally {
      setActionBusy(null);
    }
  };

  const onPause = () => runAction("pause", () => api.pauseConnector(name), "Connector paused");
  const onResume = () => runAction("resume", () => api.resumeConnector(name), "Connector resumed");

  const onDelete = async () => {
    const ok = await confirm(
      "Delete connector",
      <>
        Connector <strong className="mono">{name}</strong> and all its tasks will be removed from Kafka Connect. This
        cannot be undone.
      </>,
      { danger: true, confirmText: "Delete" },
    );
    if (!ok) return;
    const done = await runAction("delete", () => api.deleteConnector(name), "Connector deleted");
    if (done) navigate("/connectors");
  };

  if (loading) return <Spinner label="Loading connector…" />;
  if (error) return <ErrorBanner error={error} onRetry={load} />;
  if (!connector) return <EmptyState title="Connector not found" />;

  const status = connector.status;

  return (
    <div>
      <div className="page-head">
        <div>
          <h1 className="page-title mono">{connector.name}</h1>
          <p className="page-sub mono">{connector.plugin_type ?? "unknown plugin"}</p>
        </div>
        <StatusBadge status={status} />
      </div>

      <div className="actions-row">
        {status !== "paused" && (
          <Button variant="secondary" onClick={onPause} loading={actionBusy === "pause"} type="button">
            Pause
          </Button>
        )}
        {status === "paused" && (
          <Button variant="primary" onClick={onResume} loading={actionBusy === "resume"} type="button">
            Resume
          </Button>
        )}
        <Button variant="secondary" onClick={() => setRestartOpen(true)} type="button">
          Restart…
        </Button>
        <Button variant="secondary" onClick={() => setEditOpen(true)} type="button">
          Edit config
        </Button>
        <Button variant="danger" onClick={onDelete} loading={actionBusy === "delete"} type="button">
          Delete
        </Button>
        <Button variant="ghost" onClick={load} loading={actionBusy === "refresh"} type="button">
          Refresh
        </Button>
      </div>

      <div className="detail-grid">
        <div className="card" style={{ padding: 20 }}>
          <h2 className="section-title">Overview</h2>
          <dl className="kv">
            <dt>Status</dt>
            <dd>
              <StatusBadge status={status} />
            </dd>
            <dt>Worker</dt>
            <dd className="mono">{connector.worker_id || "—"}</dd>
            <dt>Plugin</dt>
            <dd className="mono">{shortClass(connector.plugin_type)}</dd>
            <dt>Tasks count</dt>
            <dd>{connector.tasks_count}</dd>
            <dt>database.hostname</dt>
            <dd className="mono">{connector.config?.database_hostname || "—"}</dd>
            <dt>database.port</dt>
            <dd className="mono">{connector.config?.database_port || "—"}</dd>
            <dt>database.user</dt>
            <dd className="mono">{connector.config?.database_user || "—"}</dd>
            <dt>database.dbname</dt>
            <dd className="mono">{connector.config?.database_dbname || "—"}</dd>
            <dt>plugin.name</dt>
            <dd className="mono">{connector.config?.plugin_name || "—"}</dd>
          </dl>
        </div>

        <div className="card" style={{ padding: 20 }}>
          <h2 className="section-title">Tasks</h2>
          {!connector.tasks || connector.tasks.length === 0 ? (
            <EmptyState title="No tasks" hint="Tasks appear when the connector starts." />
          ) : (
            <table className="table">
              <thead>
                <tr>
                  <th>Task</th>
                  <th>State</th>
                  <th>Worker</th>
                  <th />
                </tr>
              </thead>
              <tbody>
                {connector.tasks.map((t) => (
                  <tr key={t.id}>
                    <td>#{t.id}</td>
                    <td>
                      <span className={"badge badge-" + t.state.toLowerCase()}>
                        <span className="dot" />
                        {t.state.toLowerCase()}
                      </span>
                    </td>
                    <td className="mono muted">{t.worker_id || "—"}</td>
                    <td style={{ textAlign: "right" }}>
                      <TaskActions connectorName={name} taskId={t.id} onAfter={load} />
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>
      </div>

      {restartOpen && (
        <RestartModal
          name={name}
          onClose={() => setRestartOpen(false)}
          onDone={async () => {
            setRestartOpen(false);
            await load();
          }}
        />
      )}

      {editOpen && (
        <EditConfigModal
          connector={connector}
          onClose={() => setEditOpen(false)}
          onDone={async () => {
            setEditOpen(false);
            await load();
          }}
        />
      )}

      {confirmNode}
    </div>
  );
}

function TaskActions({ connectorName, taskId, onAfter }: { connectorName: string; taskId: number; onAfter: () => void }) {
  const [open, setOpen] = useState(false);
  return (
    <>
      <Button variant="ghost" className="btn-sm" onClick={() => setOpen(true)} type="button">
        Details
      </Button>
      {open && <TaskDetailModal connectorName={connectorName} taskId={taskId} onClose={() => setOpen(false)} onAfter={onAfter} />}
    </>
  );
}

function TaskDetailModal({ connectorName, taskId, onClose, onAfter }: { connectorName: string; taskId: number; onClose: () => void; onAfter: () => void }) {
  const toast = useToast();
  const [task, setTask] = useState<TaskDetail | null>(null);
  const [error, setError] = useState<unknown>(null);
  const [restarting, setRestarting] = useState(false);

  useEffect(() => {
    let cancelled = false;
    api
      .getTask(connectorName, taskId)
      .then((t) => !cancelled && setTask(t))
      .catch((e) => !cancelled && setError(e));
    return () => {
      cancelled = true;
    };
  }, [connectorName, taskId]);

  const restart = async () => {
    setRestarting(true);
    try {
      await api.restartTask(connectorName, taskId);
      toast("success", "Task #" + taskId + " restarted");
      onAfter();
      onClose();
    } catch (err) {
      toast("error", err instanceof Error ? err.message : "Restart failed");
    } finally {
      setRestarting(false);
    }
  };

  return (
    <Modal title={"Task #" + taskId + " — " + connectorName} onClose={onClose} width={640}>
      {error ? <ErrorBanner error={error} /> : null}
      {!task && !error ? <Spinner /> : null}
      {task ? (
        <>
          <dl className="kv" style={{ marginBottom: 14 }}>
            <dt>State</dt>
            <dd>
              <TaskStateBadge state={task.state} />
            </dd>
            <dt>Worker</dt>
            <dd className="mono">{task.worker_id || "—"}</dd>
          </dl>
          <h3 className="section-title">Last error trace</h3>
          {task.trace ? <pre className="trace">{task.trace}</pre> : <p className="muted">No errors recorded for this task.</p>}
          <div className="modal-actions">
            <Button variant="secondary" onClick={onClose} type="button">
              Close
            </Button>
            <Button
              variant="primary"
              onClick={restart}
              loading={restarting}
              title={task.state.toLowerCase() === "running" ? "Restart is only needed for failed or paused tasks" : undefined}
              type="button"
            >
              Restart task
            </Button>
          </div>
        </>
      ) : null}
    </Modal>
  );
}

function RestartModal({ name, onClose, onDone }: { name: string; onClose: () => void; onDone: () => void }) {
  const toast = useToast();
  const [includeTasks, setIncludeTasks] = useState(false);
  const [onlyFailed, setOnlyFailed] = useState(false);
  const [busy, setBusy] = useState(false);

  const go = async () => {
    setBusy(true);
    try {
      await api.restartConnector(name, includeTasks, onlyFailed);
      toast("success", "Connector restart requested");
      onDone();
    } catch (err) {
      toast("error", err instanceof Error ? err.message : "Restart failed");
      setBusy(false);
    }
  };

  return (
    <Modal title="Restart connector" onClose={onClose} width={460}>
      <p className="muted" style={{ marginTop: 0 }}>
        Restart <strong className="mono">{name}</strong>. Tasks can be included in the restart.
      </p>
      <div className="check-row">
        <CheckCard
          checked={includeTasks}
          onChange={(v) => {
            setIncludeTasks(v);
            if (!v) setOnlyFailed(false);
          }}
          title="Include tasks"
          desc="Restart the connector and all of its tasks"
        />
        <CheckCard
          checked={onlyFailed}
          disabled={!includeTasks}
          onChange={setOnlyFailed}
          title="Only failed"
          desc="Restart only tasks that are in the FAILED state"
        />
      </div>
      <div className="modal-actions">
        <Button variant="secondary" onClick={onClose} type="button">
          Cancel
        </Button>
        <Button onClick={go} loading={busy} type="button">
          Restart
        </Button>
      </div>
    </Modal>
  );
}

function EditConfigModal({ connector, onClose, onDone }: { connector: ConnectorDetail; onClose: () => void; onDone: () => void }) {
  const toast = useToast();
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<unknown>(null);
  const [topicPrefix, setTopicPrefix] = useState("");
  const [dbPassword, setDbPassword] = useState("");
  const [schema, setSchema] = useState<PluginSchema | null>(null);

  // current config as raw Kafka Connect keys
  const initial: Record<string, string> = {};
  for (const [k, v] of Object.entries(connector.config ?? {})) {
    if (v === null || v === undefined) continue;
    initial[dtoToRaw[k] ?? k] = v;
  }

  // fields editable through the schema form (basic connection fields);
  // everything else stays in the raw JSON textarea
  const [formValues, setFormValues] = useState<Record<string, string>>({});
  const [rawText, setRawText] = useState("");

  // load the plugin schema and split config into form fields vs raw
  useEffect(() => {
    let cancelled = false;
    if (!connector.plugin_type) {
      setSchema(null);
      setRawText(JSON.stringify(initial, null, 2));
      return;
    }
    api
      .getPluginSchema(connector.plugin_type, "all")
      .then((s) => {
        if (cancelled) return;
        setSchema(s);
        const editable = (s.fields ?? []).filter(
          (f) => f.name === "topic.prefix" || f.name.startsWith("database."),
        );
        const next: Record<string, string> = {};
        for (const f of editable) {
          if (f.name === "database.password") continue; // masked by the API, entered separately
          if (initial[f.name] !== undefined) next[f.name] = initial[f.name];
        }
        setFormValues(next);
        const formKeys = new Set(editable.map((f) => f.name));
        const rest = Object.entries(initial).filter(([k]) => !formKeys.has(k) && k !== "connector.class");
        setRawText(JSON.stringify(Object.fromEntries(rest), null, 2));
      })
      .catch(() => {
        if (!cancelled) {
          setSchema(null);
          setRawText(JSON.stringify(initial, null, 2));
        }
      });
    return () => {
      cancelled = true;
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [connector.plugin_type]);

  const buildConfig = (): Record<string, string> | null => {
    let rest: Record<string, string>;
    try {
      rest = rawText.trim() ? JSON.parse(rawText) : {};
    } catch {
      return null;
    }
    return { ...initial, ...formValues, ...rest, "topic.prefix": topicPrefix.trim(), "database.password": dbPassword };
  };

  const save = async () => {
    const cfg = buildConfig();
    if (cfg === null) {
      setError(new Error("Raw JSON section must be a valid JSON object"));
      return;
    }
    if (connector.plugin_type) cfg["connector.class"] = connector.plugin_type;
    delete cfg["name"];
    setBusy(true);
    setError(null);
    try {
      await api.updateConnector(connector.name, cfg);
      toast("success", "Configuration updated");
      onDone();
    } catch (err) {
      setError(err);
      setBusy(false);
    }
  };

  const schemaFields = useMemo(() => {
    if (!schema) return [];
    return (schema.fields ?? []).filter(
      (f) => (f.name === "topic.prefix" || f.name.startsWith("database.")) && f.name !== "database.password",
    );
  }, [schema]);

  return (
    <Modal title="Edit connector config" onClose={onClose} width={680}>
      <p className="schema-note">
        This replaces the entire configuration (PUT semantics). The API does not return secrets or{" "}
        <span className="mono">topic.prefix</span>, so provide them below.
      </p>
      <div className="form-grid-2">
        <Field label="topic.prefix" required hint="Required by Kafka Connect; not returned by the API">
          <input type="text" className="mono" value={topicPrefix} onChange={(e) => setTopicPrefix(e.target.value)} placeholder="ui-pg-cdc" />
        </Field>
        <Field label="database.password" required hint="Masked by the API; re-enter to keep">
          <input type="password" className="mono" value={dbPassword} onChange={(e) => setDbPassword(e.target.value)} placeholder="password" />
        </Field>
      </div>

      {schema && schemaFields.length > 0 ? (
        <>
          <h3 className="section-title">Connection settings</h3>
          <div className="form-grid-2">
            {schemaFields.map((f) => (
              <SchemaFieldInput
                key={f.name}
                field={f}
                value={formValues[f.name] ?? ""}
                error={null}
                onChange={(v) => setFormValues((prev) => ({ ...prev, [f.name]: v }))}
              />
            ))}
          </div>
          <h3 className="section-title">Other properties (JSON)</h3>
          <textarea
            className="mono"
            style={{ width: "100%", minHeight: 180 }}
            value={rawText}
            onChange={(e) => setRawText(e.target.value)}
            spellCheck={false}
          />
        </>
      ) : (
        <p className="schema-note">
          No schema available for this plugin — only topic.prefix, password and the raw JSON below are editable.
        </p>
      )}

      {error ? <ErrorBanner error={error} /> : null}
      <div className="modal-actions">
        <Button
          variant="ghost"
          type="button"
          onClick={() => {
            const cfg = buildConfig();
            if (cfg === null) {
              toast("error", "Fix the raw JSON first");
              return;
            }
            void navigator.clipboard.writeText(JSON.stringify(cfg, null, 2));
            toast("info", "Configuration JSON copied");
          }}
        >
          <Copy size={13} aria-hidden="true" /> Copy JSON
        </Button>
        <Button variant="secondary" onClick={onClose} type="button">
          Cancel
        </Button>
        <Button onClick={save} loading={busy} type="button" disabled={!topicPrefix.trim() || !dbPassword}>
          Save config
        </Button>
      </div>
    </Modal>
  );
}