import { createContext, useCallback, useContext, useEffect, useRef, useState } from "react";
import type { ButtonHTMLAttributes, ReactNode } from "react";
import { createPortal } from "react-dom";

// ---------- Toasts ----------

interface Toast {
  id: number;
  kind: "success" | "error" | "info";
  text: string;
}

const ToastContext = createContext<{ push: (kind: Toast["kind"], text: string) => void } | null>(null);

let toastSeq = 1;

export function ToastProvider({ children }: { children: ReactNode }) {
  const [toasts, setToasts] = useState<Toast[]>([]);

  const push = useCallback((kind: Toast["kind"], text: string) => {
    const id = toastSeq++;
    setToasts((t) => [...t, { id, kind, text }]);
    window.setTimeout(() => setToasts((t) => t.filter((x) => x.id !== id)), 5000);
  }, []);

  return (
    <ToastContext.Provider value={{ push }}>
      {children}
      <div className="toasts" aria-live="polite">
        {toasts.map((t) => (
          <div key={t.id} className={"toast toast-" + t.kind} role="status">
            {t.text}
          </div>
        ))}
      </div>
    </ToastContext.Provider>
  );
}

export function useToast() {
  const ctx = useContext(ToastContext);
  if (!ctx) throw new Error("useToast must be used within ToastProvider");
  return ctx.push;
}

// ---------- Button ----------

type ButtonVariant = "primary" | "secondary" | "danger" | "ghost";

export function Button({
  variant = "primary",
  loading = false,
  children,
  className = "",
  ...rest
}: ButtonHTMLAttributes<HTMLButtonElement> & { variant?: ButtonVariant; loading?: boolean }) {
  return (
    <button
      className={"btn btn-" + variant + " " + className}
      disabled={rest.disabled || loading}
      {...rest}
    >
      {loading ? <span className="spinner" aria-hidden="true" /> : null}
      {children}
    </button>
  );
}

// ---------- Modal ----------

export function Modal({
  title,
  onClose,
  children,
  width,
}: {
  title: string;
  onClose: () => void;
  children: ReactNode;
  width?: number;
}) {
  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if (e.key === "Escape") onClose();
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [onClose]);

  return createPortal(
    <div className="modal-backdrop" onMouseDown={(e) => e.target === e.currentTarget && onClose()}>
      <div className="modal" style={width ? { maxWidth: width } : undefined} role="dialog" aria-modal="true" aria-label={title}>
        <div className="modal-head">
          <h3>{title}</h3>
          <button className="icon-btn" onClick={onClose} aria-label="Close" type="button">
            ✕
          </button>
        </div>
        <div className="modal-body">{children}</div>
      </div>
    </div>,
    document.body,
  );
}

// ---------- Form field ----------

export function Field({
  label,
  hint,
  error,
  required,
  children,
}: {
  label: string;
  hint?: string | null;
  error?: string | null;
  required?: boolean;
  children: ReactNode;
}) {
  return (
    <label className="field">
      <span className="field-label">
        {label}
        {required ? <span className="req">*</span> : null}
      </span>
      {children}
      {hint ? <span className="field-hint">{hint}</span> : null}
      {error ? <span className="field-error">{error}</span> : null}
    </label>
  );
}

// ---------- Status badge ----------

const STATUS_LABEL: Record<string, string> = {
  running: "Running",
  paused: "Paused",
  failed: "Failed",
  starting: "Starting",
};

export function StatusBadge({ status }: { status: string }) {
  return (
    <span className={"badge badge-" + status}>
      <span className="dot" />
      {STATUS_LABEL[status] ?? status}
    </span>
  );
}

export function TaskStateBadge({ state }: { state: string }) {
  const s = state.toLowerCase();
  return <span className={"badge badge-" + s}>{state}</span>;
}

// ---------- Paginator ----------

const PAGE_SIZES = [10, 20, 50, 100];

export function Paginator({
  page,
  limit,
  hasNext,
  onPage,
  onLimit,
}: {
  page: number;
  limit?: number;
  hasNext: boolean;
  onPage: (p: number) => void;
  onLimit?: (n: number) => void;
}) {
  const hasPrev = page > 1;
  return (
    <div className="paginator">
      <Button variant="ghost" disabled={!hasPrev} onClick={() => onPage(page - 1)} type="button">
        ← Prev
      </Button>
      <span className="paginator-info">Page {page}</span>
      <Button variant="ghost" disabled={!hasNext} onClick={() => onPage(page + 1)} type="button">
        Next →
      </Button>
      {onLimit && limit !== undefined ? (
        <label className="paginator-limit">
          <select
            value={limit}
            onChange={(e) => {
              onLimit(Number(e.target.value));
              onPage(1);
            }}
            aria-label="Rows per page"
          >
            {PAGE_SIZES.map((n) => (
              <option key={n} value={n}>
                {n} / page
              </option>
            ))}
          </select>
        </label>
      ) : null}
    </div>
  );
}

// ---------- Spinner / Empty / Error ----------

export function Spinner({ label = "Loading…" }: { label?: string }) {
  return (
    <div className="center-row">
      <span className="spinner big" aria-hidden="true" />
      <span>{label}</span>
    </div>
  );
}

export function EmptyState({ title, hint }: { title: string; hint?: string }) {
  return (
    <div className="empty">
      <div className="empty-title">{title}</div>
      {hint ? <div className="empty-hint">{hint}</div> : null}
    </div>
  );
}

export function ErrorBanner({ error, onRetry }: { error: unknown; onRetry?: () => void }) {
  const message = error instanceof Error ? error.message : String(error);
  return (
    <div className="error-box">
      <div>{message}</div>
      {onRetry ? (
        <Button variant="secondary" onClick={onRetry} type="button">
          Retry
        </Button>
      ) : null}
    </div>
  );
}

// ---------- Confirm dialog ----------

export function useConfirm() {
  const [state, setState] = useState<{
    title: string;
    message: ReactNode;
    danger?: boolean;
    confirmText?: string;
    resolve: (ok: boolean) => void;
  } | null>(null);

  const confirm = useCallback(
    (title: string, message: ReactNode, opts?: { danger?: boolean; confirmText?: string }) =>
      new Promise<boolean>((resolve) => {
        setState({ title, message, resolve, ...opts });
      }),
    [],
  );

  const node = state ? (
    <Modal title={state.title} onClose={() => { state.resolve(false); setState(null); }} width={440}>
      <div className="confirm-body">{state.message}</div>
      <div className="modal-actions">
        <Button variant="secondary" onClick={() => { state.resolve(false); setState(null); }} type="button">
          Cancel
        </Button>
        <Button
          variant={state.danger ? "danger" : "primary"}
          onClick={() => { state.resolve(true); setState(null); }}
          type="button"
        >
          {state.confirmText ?? "Confirm"}
        </Button>
      </div>
    </Modal>
  ) : null;

  return { confirm, confirmNode: node };
}

// ---------- Debounced value ----------

export function useDebounced<T>(value: T, delay = 300): T {
  const [v, setV] = useState(value);
  const timer = useRef<number | undefined>(undefined);
  useEffect(() => {
    window.clearTimeout(timer.current);
    timer.current = window.setTimeout(() => setV(value), delay);
    return () => window.clearTimeout(timer.current);
  }, [value, delay]);
  return v;
}

export function shortClass(fqcn: string | null | undefined): string {
  if (!fqcn) return "—";
  const parts = fqcn.split(".");
  return parts[parts.length - 1] ?? fqcn;
}

export function fmtDate(iso: string | null | undefined): string {
  if (!iso) return "—";
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return iso;
  return d.toLocaleString();
}
