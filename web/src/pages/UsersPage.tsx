import { useCallback, useEffect, useState } from "react";
import type { FormEvent } from "react";
import * as api from "../api/endpoints";
import type { User } from "../api/types";

import {
  Button,
  EmptyState,
  ErrorBanner,
  Field,
  Modal,
  Paginator,
  Spinner,
  fmtDate,
  useConfirm,
  useDebounced,
  useToast,
} from "../components/ui";

function generatePassword(): string {
  const alphabet = "abcdefghijkmnopqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ23456789";
  const bytes = new Uint8Array(16);
  crypto.getRandomValues(bytes);
  let out = "";
  for (const b of bytes) out += alphabet[b % alphabet.length];
  return out;
}

export function UsersPage() {
  const toast = useToast();
  const { confirm, confirmNode } = useConfirm();

  const [users, setUsers] = useState<User[]>([]);
  const [meta, setMeta] = useState({ page: 1, limit: 20 });
  const [hasMore, setHasMore] = useState(false);
  const [page, setPage] = useState(1);
  const [limit, setLimit] = useState(20);
  const [role, setRole] = useState("");
  const [usernameFilter, setUsernameFilter] = useState("");
  const debouncedUsername = useDebounced(usernameFilter);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<unknown>(null);

  const [createOpen, setCreateOpen] = useState(false);
  const [renameTarget, setRenameTarget] = useState<User | null>(null);
  const [resetTarget, setResetTarget] = useState<User | null>(null);

  const load = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const res = await api.getUsers({ page, limit, role, username: debouncedUsername });
      setUsers(res.data ?? []);
      setMeta(res.meta);
      setHasMore(res.has_more ?? false);
    } catch (err) {
      setError(err);
    } finally {
      setLoading(false);
    }
  }, [page, limit, role, debouncedUsername]);

  useEffect(() => {
    load();
  }, [load]);

  const onDelete = async (u: User) => {
    const ok = await confirm(
      "Delete user",
      <>
        User <strong className="mono">{u.username}</strong> ({u.role}) will be permanently deleted.
      </>,
      { danger: true, confirmText: "Delete" },
    );
    if (!ok) return;
    try {
      await api.deleteUser(u.id);
      toast("success", "User " + u.username + " deleted");
      await load();
    } catch (err) {
      toast("error", err instanceof Error ? err.message : "Delete failed");
    }
  };

  const onResetPassword = async (u: User, newPassword: string) => {
    try {
      await api.changeUserPassword(u.id, newPassword);
      toast("success", "Password updated for " + u.username);
      setResetTarget(null);
    } catch (err) {
      toast("error", err instanceof Error ? err.message : "Password change failed");
    }
  };

  return (
    <div>
      <div className="page-head">
        <div>
          <h1 className="page-title">Users</h1>
          <p className="page-sub">Administer Maestro accounts</p>
        </div>
        <Button onClick={() => setCreateOpen(true)}>Create user</Button>
      </div>

      <div className="toolbar">
        <input
          type="text"
          placeholder="Filter by exact username…"
          value={usernameFilter}
          onChange={(e) => setUsernameFilter(e.target.value)}
          style={{ maxWidth: 260 }}
          aria-label="Filter by username"
        />
        <select value={role} onChange={(e) => setRole(e.target.value)} style={{ maxWidth: 160 }} aria-label="Filter by role">
          <option value="">All roles</option>
          <option value="admin">admin</option>
          <option value="user">user</option>
        </select>
      </div>

      {error ? <ErrorBanner error={error} onRetry={load} /> : null}

      <div className="card">
        {loading ? (
          <Spinner />
        ) : users.length === 0 ? (
          <EmptyState title="No users found" />
        ) : (
          <table className="table">
            <thead>
              <tr>
                <th>Username</th>
                <th>Role</th>
                <th>Created</th>
                <th style={{ textAlign: "right" }}>Actions</th>
              </tr>
            </thead>
            <tbody>
              {users.map((u) => (
                <tr key={u.id}>
                  <td className="mono">{u.username}</td>
                  <td>
                    <span className={"badge badge-" + (u.role === "admin" ? "starting" : "running")}>{u.role}</span>
                  </td>
                  <td className="muted">{fmtDate(u.created_at)}</td>
                  <td style={{ textAlign: "right", whiteSpace: "nowrap" }}>
                    <Button variant="ghost" className="btn-sm" onClick={() => setResetTarget(u)} type="button">
                      Reset password
                    </Button>{" "}
                    <Button variant="ghost" className="btn-sm" onClick={() => setRenameTarget(u)} type="button">
                      Rename
                    </Button>{" "}
                    <Button variant="ghost" className="btn-sm" onClick={() => onDelete(u)} type="button">
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

      {createOpen && <CreateUserModal onClose={() => setCreateOpen(false)} onCreated={load} />}
      {resetTarget && (
        <ResetPasswordModal
          user={resetTarget}
          onClose={() => setResetTarget(null)}
          onReset={(pwd) => onResetPassword(resetTarget, pwd)}
        />
      )}
      {renameTarget && (
        <RenameUserModal
          user={renameTarget}
          onClose={() => setRenameTarget(null)}
          onRenamed={() => {
            setRenameTarget(null);
            void load();
          }}
        />
      )}
      {confirmNode}
    </div>
  );
}

function CreateUserModal({ onClose, onCreated }: { onClose: () => void; onCreated: () => void }) {
  const toast = useToast();
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [role, setRole] = useState("user");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<unknown>(null);

  const onSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setBusy(true);
    setError(null);
    try {
      await api.createUser(username, password, role);
      toast("success", "User " + username + " created");
      onCreated();
      onClose();
    } catch (err) {
      setError(err);
      setBusy(false);
    }
  };

  return (
    <Modal title="Create user" onClose={onClose}>
      <form onSubmit={onSubmit}>
        <Field label="Username" required hint="3–32 chars: letters, digits, _ or -">
          <input type="text" value={username} onChange={(e) => setUsername(e.target.value)} required minLength={3} maxLength={32} pattern="[a-zA-Z0-9_\-]+" autoFocus />
        </Field>
        <Field label="Password" required hint="Min 8 characters">
          <input type="password" value={password} onChange={(e) => setPassword(e.target.value)} required minLength={8} maxLength={128} />
        </Field>
        <Field label="Role" required>
          <select value={role} onChange={(e) => setRole(e.target.value)}>
            <option value="user">user</option>
            <option value="admin">admin</option>
          </select>
        </Field>
        {error ? <ErrorBanner error={error} /> : null}
        <div className="modal-actions">
          <Button type="button" variant="secondary" onClick={onClose}>
            Cancel
          </Button>
          <Button type="submit" loading={busy}>
            Create user
          </Button>
        </div>
      </form>
    </Modal>
  );
}

function RenameUserModal({ user, onClose, onRenamed }: { user: User; onClose: () => void; onRenamed: () => void }) {
  const toast = useToast();
  const [username, setUsername] = useState(user.username);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<unknown>(null);

  const onSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setBusy(true);
    setError(null);
    try {
      await api.updateUser(user.id, username);
      toast("success", "User renamed to " + username);
      onRenamed();
    } catch (err) {
      setError(err);
      setBusy(false);
    }
  };

  return (
    <Modal title={"Rename " + user.username} onClose={onClose}>
      <form onSubmit={onSubmit}>
        <Field label="New username" required hint="3–32 chars: letters, digits, _ or -">
          <input
            type="text"
            value={username}
            onChange={(e) => setUsername(e.target.value)}
            required
            minLength={3}
            maxLength={32}
            pattern="[a-zA-Z0-9_\-]+"
            autoFocus
          />
        </Field>
        {error ? <ErrorBanner error={error} /> : null}
        <div className="modal-actions">
          <Button type="button" variant="secondary" onClick={onClose}>
            Cancel
          </Button>
          <Button type="submit" loading={busy}>
            Rename
          </Button>
        </div>
      </form>
    </Modal>
  );
}

function ResetPasswordModal({
  user,
  onClose,
  onReset,
}: {
  user: User;
  onClose: () => void;
  onReset: (newPassword: string) => void;
}) {
  const [password, setPassword] = useState(generatePassword);
  const [busy, setBusy] = useState(false);

  const onSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setBusy(true);
    try {
      onReset(password);
    } finally {
      setBusy(false);
    }
  };

  return (
    <Modal title={"Reset password for " + user.username} onClose={onClose} width={460}>
      <form onSubmit={onSubmit}>
        <Field label="New password" required hint="Min 8 characters; a strong password is pre-generated">
          <div className="pwd-row">
            <input
              type="text"
              className="mono"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
              minLength={8}
              maxLength={128}
              autoFocus
            />
            <Button type="button" variant="secondary" onClick={() => setPassword(generatePassword())}>
              Generate
            </Button>
          </div>
        </Field>
        <div className="modal-actions">
          <Button type="button" variant="secondary" onClick={onClose}>
            Cancel
          </Button>
          <Button type="submit" loading={busy}>
            Update password
          </Button>
        </div>
      </form>
    </Modal>
  );
}
