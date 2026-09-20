import { createContext, useCallback, useContext, useEffect, useMemo, useRef, useState } from "react";
import type { ReactNode } from "react";
import * as api from "../api/endpoints";
import { logout as apiLogout } from "../api/client";
import { getAccessToken, onTokenChange, setAccessToken, tryRefresh } from "../api/client";
import type { User } from "../api/types";

interface AuthState {
  user: User | null;
  loading: boolean; // initial bootstrap
  signIn: (username: string, password: string) => Promise<void>;
  signOut: () => Promise<void>;
  reloadUser: () => Promise<void>;
}

const AuthContext = createContext<AuthState | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);

  const loadUser = useCallback(async () => {
    const me = await api.getMe();
    setUser(me);
  }, []);

  // Bootstrap: if we have a token, load profile; otherwise try silent refresh.
  useEffect(() => {
    let cancelled = false;
    (async () => {
      try {
        if (!getAccessToken()) {
          const ok = await tryRefresh();
          if (!ok) {
            if (!cancelled) setLoading(false);
            return;
          }
        }
        await loadUser();
      } catch {
        setAccessToken(null);
      } finally {
        if (!cancelled) setLoading(false);
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [loadUser]);

  // If the token is cleared elsewhere (expired refresh, 401), drop the user.
  useEffect(
    () =>
      onTokenChange((token) => {
        if (!token) setUser(null);
      }),
    [],
  );

  const signIn = useCallback(
    async (username: string, password: string) => {
      const res = await api.login(username, password);
      setAccessToken(res.access_token);
      await loadUser();
    },
    [loadUser],
  );

  const signingOut = useRef(false);

  const signOut = useCallback(async () => {
    if (signingOut.current) return;
    signingOut.current = true;
    try {
      await apiLogout();
      setUser(null);
    } finally {
      signingOut.current = false;
    }
  }, []);

  const value = useMemo(
    () => ({ user, loading, signIn, signOut, reloadUser: loadUser }),
    [user, loading, signIn, signOut, loadUser],
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth(): AuthState {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error("useAuth must be used within AuthProvider");
  return ctx;
}