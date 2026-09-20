import { useEffect, useMemo, useState } from "react";
import type { FormEvent } from "react";
import { useNavigate } from "react-router-dom";
import { Copy, Check, Download } from "lucide-react";
import * as api from "../api/endpoints";
import type { PluginSchema, PluginSchemaField, ValidationReport, ValidationStep } from "../api/types";
import { ApiError } from "../api/client";
import { CustomValueInput, SchemaFieldInput } from "../components/SchemaField";
import { Button, ErrorBanner, Field, Spinner, useToast } from "../components/ui";

const STEP_LABELS: Record<string, string> = {
  config: "Configuration",
  connection: "Connection",
  permissions: "Permissions",
  cdc: "CDC readiness",
  tables: "Tables",
};

const SEV_ICON: Record<string, string> = { ok: "✓", warning: "⚠", error: "✕", skipped: "–" };

const WIZARD_STEPS = [
  { n: 1, label: "Connector" },
  { n: 2, label: "Configuration" },
  { n: 3, label: "Filters" },
  { n: 4, label: "Data options" },
  { n: 5, label: "Transformations" },
  { n: 6, label: "Runtime options" },
  { n: 7, label: "Custom properties" },
  { n: 8, label: "Review & create" },
];

// Kafka Connect infrastructure / internal / experimental fields that are not
// connector properties. Mirrors the field filter Debezium UI applies when
// generating its connector metadata (146 ConfigDef fields -> ~80 shown).
// Hidden fields can still be set via the "Custom properties" step.
const isInfraField = (name: string): boolean =>
  name === "connector.class" ||
  name === "name" ||
  name === "tasks.max" ||
  name === "tasks.max.enforce" ||
  name === "connector.plugin.version" ||
  name.endsWith(".converter") ||
  name.endsWith(".converter.plugin.version") ||
  name.startsWith("errors.") ||
  name.startsWith("internal.") ||
  name.startsWith("openlineage.") ||
  name.startsWith("guardrail.") ||
  name.startsWith("notification.") ||
  name.startsWith("lsn.flush.") ||
  name.startsWith("transaction.boundary") ||
  name === "transaction.metadata.factory" ||
  name.startsWith("snapshot.mode.configuration.based") ||
  name.startsWith("snapshot.query.mode") ||
  name.startsWith("snapshot.locking.mode") ||
  name === "snapshot.isolation.mode" ||
  name === "snapshot.max.threads.multiplier" ||
  name === "snapshot.mode.custom.name" ||
  name === "transforms" ||
  name === "predicates" ||
  name === "post.processors" ||
  name === "offsets.storage.topic" ||
  name === "topic.creation.groups" ||
  name === "config.action.reload" ||
  name === "exactly.once.support" ||
  name === "executor.shutdown.timeout.ms" ||
  name === "connection.validation.timeout.ms" ||
  name === "custom.sanitize.pattern" ||
  name === "extended.headers.enabled" ||
  name === "incremental.snapshot.watermarking.strategy" ||
  name === "slot.failover" ||
  name === "streaming.delay.ms" ||
  name === "database.query.timeout.ms" ||
  name === "sourceinfo.struct.maker" ||
  name === "publish.via.partition.root";

// Group advanced fields into semantic sections (mirrors Debezium UI steps:
// Filters / Data options / Runtime options). Name-based heuristic until the
// backend exposes the real ConfigDef group.
const categoryOf = (name: string): string => {
  if (/^(schema|table|column)\.(include|exclude)\.list|^table\.ignore\.builtin/.test(name)) return "filters";
  if (name.startsWith("snapshot.")) return "snapshot";
  if (name.startsWith("heartbeat.")) return "heartbeat";
  if (name.startsWith("topic.creation.")) return "topic-creation";
  if (name.startsWith("database.")) return "connection";
  if (
    /^(tombstones|decimal|binary|time|interval|hstore|include\.schema|schema\.name|schema\.refresh|include\.unknown|message\.|converters|column\.(truncate|mask|propagate)|datatype\.|provide\.transaction|unavailable\.|flush\.lsn)/.test(name)
  )
    return "connector";
  return "advanced";
};

const CATEGORY_LABELS: Record<string, string> = {
  connection: "Connection",
  filters: "Filters",
  snapshot: "Snapshot",
  connector: "Connector",
  heartbeat: "Heartbeat",
  "topic-creation": "Topic creation",
  advanced: "Advanced",
};



interface TransformEntry {
  alias: string;
  type: string;
  config: Record<string, string>;
}

interface CustomProp {
  key: string;
  value: string;
}

export function ConnectorCreatePage() {
  const navigate = useNavigate();
  const toast = useToast();

  const [plugins, setPlugins] = useState<string[] | null>(null);
  const [pluginsError, setPluginsError] = useState<unknown>(null);

  const [step, setStep] = useState(1);

  const [name, setName] = useState("");
  const [pluginId, setPluginId] = useState("");
  const [schema, setSchema] = useState<PluginSchema | null>(null);
  const [schemaLoading, setSchemaLoading] = useState(false);
  const [config, setConfig] = useState<Record<string, string>>({});

  const [smtPlugins, setSmtPlugins] = useState<string[] | null>(null);
  const [smtSchemas, setSmtSchemas] = useState<Record<string, PluginSchema>>({});
  const [transforms, setTransforms] = useState<TransformEntry[]>([]);

  const [customProps, setCustomProps] = useState<CustomProp[]>([]);

  const [validating, setValidating] = useState(false);
  const [report, setReport] = useState<ValidationReport | null>(null);
  const [validationNote, setValidationNote] = useState<string | null>(null);

  const [creating, setCreating] = useState(false);
  const [createError, setCreateError] = useState<unknown>(null);
  const [ackRisk, setAckRisk] = useState(false);

  useEffect(() => {
    api
      .getConnectorPlugins()
      .then((list) => {
        const ids = (list ?? []).map((p) => p.id).sort();
        setPlugins(ids);
        if (ids.length > 0) setPluginId((cur) => cur || ids[0]);
      })
      .catch(setPluginsError);
  }, []);

  useEffect(() => {
    if (!pluginId) return;
    let cancelled = false;
    setSchemaLoading(true);
    setSchema(null);
    api
      .getPluginSchema(pluginId, "all")
      .then((s) => {
        if (cancelled) return;
        setSchema(s);
        const next: Record<string, string> = {};
        for (const f of s.fields ?? []) {
          if (isInfraField(f.name)) continue;
          if (f.default !== null && f.default !== undefined) next[f.name] = f.default;
        }
        setConfig(next);
      })
      .catch((err) => {
        if (!cancelled) setSchema({ fields: [] });
        console.warn("schema load failed", err);
      })
      .finally(() => !cancelled && setSchemaLoading(false));
    return () => {
      cancelled = true;
    };
  }, [pluginId]);

  // load SMT plugin list once, when the wizard mounts
  useEffect(() => {
    api
      .getSmtPlugins()
      .then((list) => setSmtPlugins((list ?? []).map((p) => p.id).sort()))
      .catch(() => setSmtPlugins([]));
  }, []);

  // any edit invalidates the previous validation result
  useEffect(() => {
    setReport(null);
    setValidationNote(null);
    setAckRisk(false);
  }, [pluginId, name, config, transforms, customProps]);

  const compact = (cfg: Record<string, string>): Record<string, string> => {
    const out: Record<string, string> = {};
    for (const [k, v] of Object.entries(cfg)) {
      if (v !== "" && v !== undefined && v !== null) out[k] = v;
    }
    return out;
  };

  const buildConfig = (): Record<string, string> => {
    const cfg: Record<string, string> = { "connector.class": pluginId };
    const trimmedName = name.trim();
    if (trimmedName) cfg["name"] = trimmedName;

    const merged: Record<string, string> = { ...compact(config) };

    for (const t of transforms) {
      const alias = t.alias.trim();
      if (!alias || !t.type) continue;
      merged["transforms." + alias + ".type"] = t.type;
      for (const [k, v] of Object.entries(compact(t.config))) {
        merged["transforms." + alias + "." + k] = v;
      }
    }

    for (const p of customProps) {
      const key = p.key.trim();
      if (key && p.value !== "") merged[key] = p.value;
    }

    return { ...cfg, ...merged };
  };

  const runValidate = async () => {
    if (!pluginId) return;
    setValidating(true);
    setReport(null);
    setValidationNote(null);
    setAckRisk(false);
    try {
      const res = await api.validateConnector(pluginId, name.trim(), buildConfig());
      setReport(res);
    } catch (err) {
      if (err instanceof ApiError && err.code === "CONNECTOR_CONFIG_INVALID" && err.details) {
        setReport(err.details as ValidationReport);
      } else if (err instanceof ApiError && err.code === "CONNECTOR_VALIDATION_UNSUPPORTED") {
        setValidationNote(err.message);
      } else {
        toast("error", err instanceof Error ? err.message : "Validation failed");
      }
    } finally {
      setValidating(false);
    }
  };

  const goReview = () => {
    setStep(8);
    void runValidate();
  };

  const requiredFields = useMemo(
    () => (schema?.fields ?? []).filter((f) => f.required && f.name !== "name" && f.name !== "connector.class"),
    [schema],
  );

  const missingRequired = requiredFields.filter((f) => !(config[f.name] ?? "").trim());

  // "most important" = connection essentials (like Debezium UI basic properties):
  // topic prefix + connection fields of the database. Everything else goes under Advanced.
  const isConnectionField = (f: PluginSchemaField): boolean =>
    f.name === "topic.prefix" ||
    f.name.startsWith("database.") ||
    f.name.startsWith("mongodb.") ||
    f.name.startsWith("schema.registry");

  const visibleFields = useMemo(
    () => (schema?.fields ?? []).filter((f) => !isInfraField(f.name)),
    [schema],
  );

  const basicFields = useMemo(
    () =>
      visibleFields.filter(
        (f) => f.required || (f.importance === "HIGH" && isConnectionField(f)),
      ),
    [visibleFields],
  );

  const advancedFields = useMemo(
    () =>
      visibleFields.filter(
        (f) => !(f.required || (f.importance === "HIGH" && isConnectionField(f))),
      ),
    [visibleFields],
  );

  const advancedByCategory = useMemo(() => {
    const groups: Record<string, PluginSchemaField[]> = {};
    for (const f of advancedFields) {
      const cat = categoryOf(f.name);
      (groups[cat] ??= []).push(f);
    }
    return groups;
  }, [advancedFields]);

  const renderCategoryFields = (cat: string) => {
    const fields = advancedByCategory[cat] ?? [];
    if (fields.length === 0) return null;
    return (
      <AdvSection title={`${CATEGORY_LABELS[cat]} (${fields.length})`}>
        <div className="form-grid-2">
          {fields.map((f) => (
            <SchemaFieldInput
              key={f.name}
              field={f}
              value={config[f.name] ?? ""}
              error={fieldErrors[f.name] ?? null}
              onChange={(v) => setConfig((prev) => ({ ...prev, [f.name]: v }))}
            />
          ))}
        </div>
      </AdvSection>
    );
  };

  // fields hidden from the form (Kafka Connect infrastructure) — still settable
  // via the Custom properties step, with proper input types from their schema
  const hiddenFields = useMemo(
    () => (schema?.fields ?? []).filter((f) => isInfraField(f.name)),
    [schema],
  );

  const fieldErrors = useMemo(() => {
    const m: Record<string, string> = {};
    for (const s of report?.steps ?? []) {
      for (const c of s.checks) {
        if (c.severity === "error" && c.field && !m[c.field]) m[c.field] = c.message;
      }
    }
    return m;
  }, [report]);

  const reportFailed = report !== null && !report.valid;

  const createBlocked =
    validating ||
    creating ||
    !name.trim() ||
    !pluginId ||
    missingRequired.length > 0 ||
    (reportFailed && !ackRisk);

  // Secret values must never render in plain text on the review step.
  // Copy/Download below intentionally keep raw values for reuse.
  const maskedConfig = useMemo(() => {
    const secretNames = new Set(
      (schema?.fields ?? []).filter((f) => f.type === "PASSWORD").map((f) => f.name),
    );
    const out = buildConfig();
    for (const k of Object.keys(out)) {
      const leaf = k.slice(k.lastIndexOf(".") + 1).toLowerCase();
      if (
        secretNames.has(k) ||
        leaf.includes("password") ||
        leaf.includes("secret") ||
        leaf.includes("passwd")
      ) {
        out[k] = "***";
      }
    }
    return out;
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [schema, config, transforms, customProps, name, pluginId]);

  const onSubmit = async (e: FormEvent) => {
    e.preventDefault();
    if (step !== 8) return;
    setCreateError(null);
    setCreating(true);
    try {
      const res = await api.createConnector({
        name: name.trim(),
        config: buildConfig(),
      });
      toast("success", "Connector " + res.name + " created");
      navigate("/connectors/" + encodeURIComponent(res.name));
    } catch (err) {
      setCreateError(err);
    } finally {
      setCreating(false);
    }
  };

  // ---- SMT helpers ----
  const addTransform = () => {
    setTransforms((prev) => [...prev, { alias: "", type: "", config: {} }]);
  };

  const removeTransform = (index: number) => {
    setTransforms((prev) => prev.filter((_, i) => i !== index));
  };

  const updateTransform = (index: number, patch: Partial<TransformEntry>) => {
    setTransforms((prev) => prev.map((t, i) => (i === index ? { ...t, ...patch } : t)));
  };

  const loadSmtSchema = (type: string) => {
    if (!type || smtSchemas[type]) return;
    api
      .getSmtPluginSchema(type, "all")
      .then((s) => {
        setSmtSchemas((prev) => ({ ...prev, [type]: s }));
        // prefill defaults into every transform that uses this type
        setTransforms((prev) =>
          prev.map((t) => {
            if (t.type !== type) return t;
            const next = { ...t.config };
            for (const f of s.fields ?? []) {
              if (f.default !== null && f.default !== undefined && next[f.name] === undefined) {
                next[f.name] = f.default;
              }
            }
            return { ...t, config: next };
          }),
        );
      })
      .catch(() => {
        setSmtSchemas((prev) => ({ ...prev, [type]: { fields: [] } }));
      });
  };

  const onTransformTypeChange = (index: number, type: string) => {
    updateTransform(index, { type, config: {} });
    loadSmtSchema(type);
  };

  // ---- custom properties helpers ----
  const addCustomProp = () => {
    setCustomProps((prev) => [...prev, { key: "", value: "" }]);
  };

  const addKnownField = (name: string) => {
    const f = hiddenFields.find((x) => x.name === name);
    if (!f) return;
    setCustomProps((prev) => [...prev, { key: f.name, value: f.default ?? "" }]);
  };

  const removeCustomProp = (index: number) => {
    setCustomProps((prev) => prev.filter((_, i) => i !== index));
  };

  const updateCustomProp = (index: number, patch: Partial<CustomProp>) => {
    setCustomProps((prev) => prev.map((p, i) => (i === index ? { ...p, ...patch } : p)));
  };

  if (pluginsError) {
    return (
      <div>
        <h1 className="page-title">Create connector</h1>
        <ErrorBanner error={pluginsError} />
      </div>
    );
  }

  if (!plugins) {
    return <Spinner label="Loading plugins…" />;
  }

  return (
    <div style={{ maxWidth: 860 }}>
      <div className="page-head">
        <div>
          <h1 className="page-title">Create connector</h1>
          <p className="page-sub">Pick a connector, configure it, then review the validation before creating.</p>
        </div>
      </div>

      <div className="wizard-steps">
        {WIZARD_STEPS.map((s) => (
          <div
            key={s.n}
            className={"wizard-step" + (step === s.n ? " active" : "") + (step > s.n ? " done" : "")}
          >
            <span className="wizard-num">{step > s.n ? "✓" : s.n}</span>
            {s.label}
          </div>
        ))}
      </div>

      <form className="card" style={{ padding: 20 }} onSubmit={onSubmit}>
        {step === 1 ? (
          <>
            <h2 className="section-title">Choose a connector type</h2>
            <Field label="Connector plugin" required hint="Debezium connector class">
              <select value={pluginId} onChange={(e) => setPluginId(e.target.value)}>
                {plugins.map((p) => (
                  <option key={p} value={p}>
                    {p}
                  </option>
                ))}
              </select>
            </Field>
            <div className="modal-actions">
              <Button type="button" variant="secondary" onClick={() => navigate("/connectors")}>
                Cancel
              </Button>
              <Button type="button" disabled={!pluginId} onClick={() => setStep(2)}>
                Next
              </Button>
            </div>
          </>
        ) : null}

        {step === 2 ? (
          <>
            <h2 className="section-title">Configuration</h2>
            <div className="form-grid-2">
              <Field
                label="Connector name"
                required
                hint="Unique name, 1–128 characters"
                error={fieldErrors["name"] ?? null}
              >
                <input
                  type="text"
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                  placeholder="pg-orders-cdc"
                  required
                  maxLength={128}
                  minLength={1}
                />
              </Field>
              <Field label="Plugin" required hint="Debezium connector class">
                <select value={pluginId} onChange={(e) => setPluginId(e.target.value)}>
                  {plugins.map((p) => (
                    <option key={p} value={p}>
                      {p}
                    </option>
                  ))}
                </select>
              </Field>
            </div>

            {schemaLoading ? (
              <Spinner label="Loading plugin schema…" />
            ) : schema && schema.fields.length > 0 ? (
              <>
                <div className="form-grid-2">
                  {basicFields.map((f) => (
                    <SchemaFieldInput
                      key={f.name}
                      field={f}
                      value={config[f.name] ?? ""}
                      error={fieldErrors[f.name] ?? null}
                      onChange={(v) => setConfig((prev) => ({ ...prev, [f.name]: v }))}
                    />
                  ))}
                </div>

                {(advancedByCategory["connection"]?.length ?? 0) > 0 ? (
                  <AdvSection title={`${CATEGORY_LABELS["connection"]} (${advancedByCategory["connection"].length})`}>
                    <div className="form-grid-2">
                      {advancedByCategory["connection"].map((f) => (
                        <SchemaFieldInput
                          key={f.name}
                          field={f}
                          value={config[f.name] ?? ""}
                          error={fieldErrors[f.name] ?? null}
                          onChange={(v) => setConfig((prev) => ({ ...prev, [f.name]: v }))}
                        />
                      ))}
                    </div>
                  </AdvSection>
                ) : null}
              </>
            ) : (
              <p className="schema-note">
                No schema available for this plugin — you can still review and create with an empty config.
              </p>
            )}

            <div className="modal-actions">
              <Button type="button" variant="secondary" onClick={() => setStep(1)}>
                Back
              </Button>
              <Button
                type="button"
                disabled={!name.trim() || missingRequired.length > 0}
                title={
                  !name.trim()
                    ? "Enter a connector name to continue"
                    : missingRequired.length > 0
                      ? "Fill required fields: " + missingRequired.map((f) => f.name).join(", ")
                      : undefined
                }
                onClick={() => setStep(3)}
              >
                Next
              </Button>
            </div>
          </>
        ) : null}

        {step === 3 ? (
          <>
            <h2 className="section-title">Filters</h2>
            <p className="schema-note">
              Which schemas, tables and columns to capture or exclude from replication.
            </p>
            {schemaLoading ? <Spinner label="Loading plugin schema…" /> : renderCategoryFields("filters")}
            <div className="modal-actions">
              <Button type="button" variant="secondary" onClick={() => setStep(2)}>
                Back
              </Button>
              <Button type="button" onClick={() => setStep(4)}>
                Next
              </Button>
            </div>
          </>
        ) : null}

        {step === 4 ? (
          <>
            <h2 className="section-title">Data options</h2>
            <p className="schema-note">
              Snapshot behaviour and data type mapping.
            </p>
            {schemaLoading ? (
              <Spinner label="Loading plugin schema…" />
            ) : (
              <>
                {renderCategoryFields("snapshot")}
                {renderCategoryFields("connector")}
              </>
            )}
            <div className="modal-actions">
              <Button type="button" variant="secondary" onClick={() => setStep(3)}>
                Back
              </Button>
              <Button type="button" onClick={() => setStep(5)}>
                Next
              </Button>
            </div>
          </>
        ) : null}

        {step === 5 ? (
          <>
            <h2 className="section-title">Transformations (SMT)</h2>
            <p className="schema-note">
              Optional single message transforms applied to records before they are written to Kafka.
            </p>

            {smtPlugins === null ? (
              <Spinner label="Loading SMT plugins…" />
            ) : (
              <div className="smt-list">
                {transforms.map((t, i) => {
                  const tSchema = t.type ? smtSchemas[t.type] : undefined;
                  return (
                    <div key={i} className="smt-entry">
                      <div className="form-grid-2">
                        <Field label="Alias" required hint={"transforms." + (t.alias || "<alias>") + ".type"}>
                          <input
                            type="text"
                            value={t.alias}
                            onChange={(e) => updateTransform(i, { alias: e.target.value })}
                            placeholder="route"
                            className="mono"
                          />
                        </Field>
                        <Field label="Type" required hint="SMT class">
                          <select
                            value={t.type}
                            onChange={(e) => onTransformTypeChange(i, e.target.value)}
                          >
                            <option value="">— select —</option>
                            {smtPlugins.map((p) => (
                              <option key={p} value={p}>
                                {p}
                              </option>
                            ))}
                          </select>
                        </Field>
                      </div>

                      {t.type ? (
                        tSchema && tSchema.fields.length > 0 ? (
                          <div className="form-grid-2 smt-fields">
                            {tSchema.fields
                              .filter((f) => f.name !== "type")
                              .map((f) => (
                                <SchemaFieldInput
                                  key={f.name}
                                  field={f}
                                  value={t.config[f.name] ?? ""}
                                  error={null}
                                  onChange={(v) =>
                                    updateTransform(i, {
                                      config: { ...t.config, [f.name]: v },
                                    })
                                  }
                                />
                              ))}
                          </div>
                        ) : (
                          <p className="schema-note">No schema available for this SMT.</p>
                        )
                      ) : null}

                      <div className="modal-actions">
                        <Button type="button" variant="secondary" onClick={() => removeTransform(i)}>
                          Remove
                        </Button>
                      </div>
                    </div>
                  );
                })}

                <Button type="button" variant="secondary" onClick={addTransform}>
                  Add transformation
                </Button>
              </div>
            )}

            <div className="modal-actions">
              <Button type="button" variant="secondary" onClick={() => setStep(4)}>
                Back
              </Button>
              <Button type="button" onClick={() => setStep(6)}>
                Next
              </Button>
            </div>
          </>
        ) : null}

        {step === 6 ? (
          <>
            <h2 className="section-title">Runtime options</h2>
            <p className="schema-note">
              Engine behaviour, heartbeats and signal channels.
            </p>
            {schemaLoading ? (
              <Spinner label="Loading plugin schema…" />
            ) : (
              <>
                {renderCategoryFields("heartbeat")}
                {renderCategoryFields("advanced")}
              </>
            )}
            <div className="modal-actions">
              <Button type="button" variant="secondary" onClick={() => setStep(5)}>
                Back
              </Button>
              <Button type="button" onClick={() => setStep(7)}>
                Next
              </Button>
            </div>
          </>
        ) : null}

        {step === 7 ? (
          <>
            <h2 className="section-title">Custom properties</h2>
            <p className="schema-note">
              Additional connector properties not present in the form — e.g. transforms.*, topic.creation.* or
              custom keys of your connector. Hidden Kafka Connect fields (tasks.max, converters, errors.*, …)
              can be added from the list below with proper input types.
            </p>

            {hiddenFields.length > 0 ? (
              <Field
                label="Add hidden field"
                hint="Fields hidden from the form — added with their schema (type, enum, default)"
              >
                <select
                  value=""
                  onChange={(e) => {
                    if (e.target.value) {
                      addKnownField(e.target.value);
                      e.target.value = "";
                    }
                  }}
                >
                  <option value="">— add known field —</option>
                  {hiddenFields.map((f) => (
                    <option key={f.name} value={f.name}>
                      {f.label || f.name} ({f.name})
                    </option>
                  ))}
                </select>
              </Field>
            ) : null}

            {customProps.map((p, i) => {
              const known = hiddenFields.find((f) => f.name === p.key.trim());
              return (
                <div key={i} className="kv-row">
                  <div className="form-grid-2">
                    <Field label="Key" hint={known?.description ?? undefined}>
                      <input
                        type="text"
                        value={p.key}
                        onChange={(e) => updateCustomProp(i, { key: e.target.value })}
                        placeholder="custom.key"
                        className="mono"
                      />
                    </Field>
                    <Field
                      label="Value"
                      hint={
                        known
                          ? "type: " + known.type + (known.default ? ", default: " + known.default : "")
                          : undefined
                      }
                    >
                      {known ? (
                        <CustomValueInput
                          field={known}
                          value={p.value}
                          onChange={(v) => updateCustomProp(i, { value: v })}
                        />
                      ) : (
                        <input
                          type="text"
                          value={p.value}
                          onChange={(e) => updateCustomProp(i, { value: e.target.value })}
                          placeholder="value"
                          className="mono"
                        />
                      )}
                    </Field>
                  </div>
                  <div className="modal-actions">
                    <Button type="button" variant="secondary" onClick={() => removeCustomProp(i)}>
                      Remove
                    </Button>
                  </div>
                </div>
              );
            })}

            <Button type="button" variant="secondary" onClick={addCustomProp}>
              Add property
            </Button>

            <div className="modal-actions">
              <Button type="button" variant="secondary" onClick={() => setStep(6)}>
                Back
              </Button>
              <Button
                type="button"
                disabled={!name.trim() || missingRequired.length > 0}
                title={
                  !name.trim()
                    ? "Enter a connector name to continue"
                    : missingRequired.length > 0
                      ? "Fill required fields: " + missingRequired.map((f) => f.name).join(", ")
                      : undefined
                }
                onClick={goReview}
              >
                Review & validate
              </Button>
            </div>
          </>
        ) : null}

        {step === 8 ? (
          <>
            <h2 className="section-title">Review & create</h2>
            <p className="schema-note">
              {name.trim() || "—"} · {pluginId}
            </p>

            <pre className="review-json">{JSON.stringify(maskedConfig, null, 2)}</pre>

            <div className="review-tools">
              <Button
                type="button"
                variant="ghost"
                className="btn-sm"
                onClick={() => {
                  void navigator.clipboard.writeText(JSON.stringify(buildConfig(), null, 2));
                  toast("info", "Configuration JSON copied");
                }}
              >
                <Copy size={13} aria-hidden="true" /> Copy JSON
              </Button>
              <Button
                type="button"
                variant="ghost"
                className="btn-sm"
                onClick={() => {
                  const blob = new Blob([JSON.stringify(buildConfig(), null, 2)], { type: "application/json" });
                  const url = URL.createObjectURL(blob);
                  const a = document.createElement("a");
                  a.href = url;
                  a.download = (name.trim() || "connector") + ".json";
                  a.click();
                  URL.revokeObjectURL(url);
                }}
              >
                <Download size={13} aria-hidden="true" /> Download JSON
              </Button>
            </div>

            {validating ? <Spinner label="Validating configuration…" /> : null}

            {validationNote ? <div className="val-verdict warn">{validationNote}</div> : null}

            {report ? (
              <>
                <div className={"val-verdict " + (report.valid ? "ok" : "bad")}>
                  {report.valid
                    ? "Validation passed — the connector can be created"
                    : "Validation found errors — fix them before creating"}
                </div>
                <div className="val-steps">
                  {report.steps.map((s) => (
                    <StepCard key={s.id} step={s} />
                  ))}
                </div>
              </>
            ) : null}

            {createError ? <ErrorBanner error={createError} /> : null}

            {reportFailed ? (
              <div className="val-verdict warn">
                Validation found errors — the connector will likely fail at runtime.
                <label className={"check-card risk" + (ackRisk ? " checked" : "")} style={{ marginTop: 8 }}>
                  <input
                    type="checkbox"
                    checked={ackRisk}
                    onChange={(e) => setAckRisk(e.target.checked)}
                  />
                  <span className="box" aria-hidden="true">
                    <Check size={13} strokeWidth={3.5} />
                  </span>
                  <span className="check-text">
                    <span className="check-title">Create anyway, I understand the risk</span>
                  </span>
                </label>
              </div>
            ) : null}

            <div className="modal-actions">
              <Button type="button" variant="secondary" onClick={() => setStep(7)}>
                Back
              </Button>
              <Button type="button" variant="secondary" onClick={runValidate} loading={validating}>
                Re-validate
              </Button>
              <Button
                type="submit"
                loading={creating}
                disabled={createBlocked}
                title={
                  reportFailed && !ackRisk
                    ? "Validation errors block creation — acknowledge the risk to proceed"
                    : missingRequired.length > 0
                      ? "Fill required fields: " + missingRequired.map((f) => f.name).join(", ")
                      : undefined
                }
              >
                Create connector
              </Button>
            </div>
          </>
        ) : null}
      </form>
    </div>
  );
}

function StepCard({ step }: { step: ValidationStep }) {
  return (
    <div className="val-step">
      <div className="val-step-head">
        <span>{STEP_LABELS[step.id] ?? step.id}</span>
        <span className={"sev-" + step.status}>{SEV_ICON[step.status] ?? step.status}</span>
      </div>
      {step.checks.map((c, i) => (
        <div key={i} className="val-check">
          <span className={"sev-" + c.severity}>{SEV_ICON[c.severity] ?? c.severity}</span>
          <div className="val-msg">
            <div>{c.message}</div>
            {c.fix_hint ? <code className="val-fix">{c.fix_hint}</code> : null}
            {c.field || c.table ? (
              <div className="val-chips">
                {c.field ? <span className="val-chip">field: {c.field}</span> : null}
                {c.table ? <span className="val-chip">table: {c.table}</span> : null}
              </div>
            ) : null}
          </div>
        </div>
      ))}
    </div>
  );
}

// Collapsible section for advanced fields: closed by default, opens on click.
function AdvSection({ title, children }: { title: string; children: React.ReactNode }) {
  const [open, setOpen] = useState(false);

  return (
    <div className="adv-section">
      <button
        type="button"
        className={"adv-section-title" + (open ? " open" : "")}
        onClick={() => setOpen((v) => !v)}
        aria-expanded={open}
      >
        {title}
        <span className="adv-chevron">{open ? "▾" : "▸"}</span>
      </button>
      {open ? <div className="adv-section-body">{children}</div> : null}
    </div>
  );
}
