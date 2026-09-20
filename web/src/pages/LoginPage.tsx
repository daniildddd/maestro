import { useState } from "react";
import type { FormEvent } from "react";
import { useNavigate } from "react-router-dom";
import { useAuth } from "../auth/AuthContext";
import { Button, Field } from "../components/ui";

export function LoginPage() {
  const { signIn } = useAuth();
  const navigate = useNavigate();
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);

  const onSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setError(null);
    setBusy(true);
    try {
      await signIn(username, password);
      navigate("/dashboard", { replace: true });
    } catch (err) {
      setError(err instanceof Error ? err.message : "Login failed");
    } finally {
      setBusy(false);
    }
  };

  return (
    <div className="login-wrap">
      <form className="login-card" onSubmit={onSubmit}>
        <div className="login-brand">
          <img className="brand-mark" src="/maestro.png" alt="Maestro" /> Maestro
        </div>
        <p className="login-sub">Sign in to manage Debezium connectors</p>

        <Field label="Username" required>
          <input
            type="text"
            value={username}
            onChange={(e) => setUsername(e.target.value)}
            placeholder="username"
            autoFocus
            required
            minLength={3}
            maxLength={32}
          />
        </Field>
        <Field label="Password" required>
          <input
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            placeholder="password"
            required
            minLength={8}
            maxLength={128}
          />
        </Field>

        {error ? <div className="error-box">{error}</div> : null}

        <Button type="submit" loading={busy} style={{ width: "100%", marginTop: 8 }}>
          Sign in
        </Button>
      </form>
    </div>
  );
}
