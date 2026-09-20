import { useState } from "react";
import type { FormEvent } from "react";
import { useNavigate } from "react-router-dom";
import { useAuth } from "../auth/AuthContext";
import * as api from "../api/endpoints";
import { Button, Field, fmtDate, useToast } from "../components/ui";

function Card({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <div className="card" style={{ padding: 20, marginBottom: 16 }}>
      <h2 className="section-title">{title}</h2>
      {children}
    </div>
  );
}

export function ProfilePage() {
  const { user } = useAuth();
  const toast = useToast();
  const navigate = useNavigate();

  return (
    <div style={{ maxWidth: 640 }}>
      <div className="page-head">
        <div>
          <h1 className="page-title">Profile</h1>
          <p className="page-sub">Your account settings</p>
        </div>
      </div>

      <Card title="Account">
        <dl className="kv">
          <dt>Username</dt>
          <dd className="mono">{user?.username}</dd>
          <dt>Role</dt>
          <dd>{user?.role}</dd>
          <dt>Created</dt>
          <dd>{fmtDate(user?.created_at)}</dd>
          <dt>Updated</dt>
          <dd>{fmtDate(user?.updated_at)}</dd>
        </dl>
      </Card>

      <Card title="Change password">
        <ChangePasswordForm
          onDone={async () => {
            toast("success", "Password changed");
          }}
        />
      </Card>

      <Card title="Danger zone">
        <p className="muted" style={{ marginTop: 0 }}>
          Deleting your account is permanent. You will be signed out immediately.
        </p>
        <DeleteAccountForm
          onDeleted={() => {
            toast("info", "Account deleted");
            navigate("/login", { replace: true });
          }}
        />
      </Card>
    </div>
  );
}

function ChangePasswordForm({ onDone }: { onDone: () => void }) {
  const [oldPassword, setOld] = useState("");
  const [newPassword, setNew] = useState("");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<unknown>(null);

  const onSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setBusy(true);
    setError(null);
    try {
      await api.changeOwnPassword(oldPassword, newPassword);
      setOld("");
      setNew("");
      onDone();
    } catch (err) {
      setError(err);
    } finally {
      setBusy(false);
    }
  };

  return (
    <form onSubmit={onSubmit}>
      <Field label="Current password" required>
        <input type="password" value={oldPassword} onChange={(e) => setOld(e.target.value)} required minLength={8} />
      </Field>
      <Field label="New password" required hint="Min 8 characters">
        <input type="password" value={newPassword} onChange={(e) => setNew(e.target.value)} required minLength={8} />
      </Field>
      {error ? <div className="error-box">{error instanceof Error ? error.message : String(error)}</div> : null}
      <Button type="submit" loading={busy}>
        Change password
      </Button>
    </form>
  );
}

function DeleteAccountForm({ onDeleted }: { onDeleted: () => void }) {
  const [password, setPassword] = useState("");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<unknown>(null);

  const onSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setBusy(true);
    setError(null);
    try {
      await api.deleteMe(password);
      onDeleted();
    } catch (err) {
      setError(err);
      setBusy(false);
    }
  };

  return (
    <form onSubmit={onSubmit}>
      <Field label="Confirm with your password" required>
        <input type="password" value={password} onChange={(e) => setPassword(e.target.value)} required minLength={8} />
      </Field>
      {error ? <div className="error-box">{error instanceof Error ? error.message : String(error)}</div> : null}
      <Button type="submit" variant="danger" loading={busy}>
        Delete my account
      </Button>
    </form>
  );
}