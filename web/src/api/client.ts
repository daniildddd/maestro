// Low-level API client with JWT access token handling and single-flight
// refresh on 401 (refresh token rides in an HttpOnly cookie, same origin).

export class ApiError extends Error {
  code: string;
  status: number;
  details: unknown;

  constructor(status: number, code: string, message: string, details?: unknown) {
    super(message);
    this.status = status;
    this.code = code;
    this.details = details;
  }
}

let accessToken: string | null = null;
const tokenListeners = new Set<(token: string | null) => void>();

export function getAccessToken(): string | null {
  return accessToken;
}

export function setAccessToken(token: string | null) {
  accessToken = token;
  tokenListeners.forEach((fn) => fn(token));
}

export function onTokenChange(fn: (token: string | null) => void): () => void {
  tokenListeners.add(fn);
  return () => tokenListeners.delete(fn);
}

export class UnauthorizedError extends ApiError {
  constructor(code = "UNAUTHORIZED", message = "Not authenticated") {
    super(401, code, message);
  }
}

let refreshPromise: Promise<boolean> | null = null;

async function doRefresh(): Promise<boolean> {
  try {
    const res = await fetch("/api/v1/refresh", {
      method: "POST",
      credentials: "same-origin",
    });
    if (res.status === 204) return false; // no session — nothing to refresh
    if (!res.ok) return false;
    const body = (await res.json()) as { access_token: string };
    setAccessToken(body.access_token);
    return true;
  } catch {
    return false;
  }
}

export function tryRefresh(): Promise<boolean> {
  if (!refreshPromise) {
    refreshPromise = doRefresh().finally(() => {
      refreshPromise = null;
    });
  }
  return refreshPromise;
}

export async function logout(): Promise<void> {
  try {
    await fetch("/api/v1/logout", {
      method: "POST",
      credentials: "same-origin",
    });
  } finally {
    setAccessToken(null);
  }
}

interface RequestOptions {
  method?: string;
  body?: unknown;
  raw?: boolean; // don't auto-refresh / don't throw UnauthorizedError
  skipAuth?: boolean;
}

export async function api<T = unknown>(
  path: string,
  opts: RequestOptions = {},
): Promise<T> {
  const { method = "GET", body, raw = false, skipAuth = false } = opts;

  const headers: Record<string, string> = {};
  if (body !== undefined) headers["Content-Type"] = "application/json";
  if (!skipAuth && accessToken) headers["Authorization"] = "Bearer " + accessToken;

  const send = () =>
    fetch(path, {
      method,
      headers,
      credentials: "same-origin",
      body: body === undefined ? null : JSON.stringify(body),
    });

  let res = await send();

  if (res.status === 401 && !raw && !skipAuth) {
    const refreshed = await tryRefresh();
    if (refreshed) {
      if (accessToken) headers["Authorization"] = "Bearer " + accessToken;
      res = await send();
    }
  }

  if (res.status === 204) return undefined as T;

  let payload: unknown = null;
  const text = await res.text();
  if (text) {
    try {
      payload = JSON.parse(text);
    } catch {
      payload = { code: "BAD_RESPONSE", message: text.slice(0, 300) };
    }
  }

  if (!res.ok) {
    const err = payload as { code?: string; message?: string; details?: unknown } | null;
    const apiErr = new ApiError(
      res.status,
      err?.code ?? "HTTP_" + res.status,
      err?.message ?? "Request failed (" + res.status + ")",
      err?.details,
    );
    if (res.status === 401 && !raw) throw new UnauthorizedError(apiErr.code, apiErr.message);
    throw apiErr;
  }

  return payload as T;
}

export function qs(params: Record<string, string | number | boolean | undefined | null>): string {
  const search = new URLSearchParams();
  for (const [k, v] of Object.entries(params)) {
    if (v !== undefined && v !== null && v !== "") search.set(k, String(v));
  }
  const s = search.toString();
  return s ? "?" + s : "";
}
