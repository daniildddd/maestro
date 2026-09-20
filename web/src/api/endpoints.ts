import { api, qs } from "./client";
import type {
  AuditLogEntry,
  AuditLogListResponse,
  ConnectorCreateResponse,
  ConnectorDetail,
  ConnectorListResponse,
  LoginResponse,
  PluginListItem,
  PluginSchema,
  TaskDetail,
  User,
  UserListResponse,
  ValidationReport,
} from "./types";

// ---- auth ----
export const login = (username: string, password: string) =>
  api<LoginResponse>("/api/v1/login", { method: "POST", body: { username, password }, raw: true });

// ---- users ----
export const getMe = () => api<User>("/api/v1/users/me");
export const getUsers = (params: { page?: number; limit?: number; username?: string; role?: string }) =>
  api<UserListResponse>("/api/v1/users" + qs(params));
export const createUser = (username: string, password: string, role: string) =>
  api<User>("/api/v1/users", { method: "POST", body: { username, password, role } });
export const updateUser = (id: string, username: string) =>
  api<User>("/api/v1/users/" + id, { method: "PATCH", body: { username } });
export const deleteUser = (id: string) =>
  api<void>("/api/v1/users/" + id, { method: "DELETE" });
export const changeUserPassword = (id: string, newPassword: string) =>
  api<void>("/api/v1/users/" + id + "/password", { method: "PATCH", body: { new_password: newPassword } });
export const changeOwnPassword = (oldPassword: string, newPassword: string) =>
  api<void>("/api/v1/users/me/password", {
    method: "PATCH",
    body: { old_password: oldPassword, new_password: newPassword },
  });
export const deleteMe = (password: string) =>
  api<void>("/api/v1/users/me", { method: "DELETE", body: { password } });

// ---- connectors ----
export const getConnectors = (params: { page?: number; limit?: number; status?: string; search?: string }) =>
  api<ConnectorListResponse>("/api/v1/connectors" + qs(params));
export const getConnector = (name: string) => api<ConnectorDetail>("/api/v1/connectors/" + encodeURIComponent(name));
export const createConnector = (payload: { name: string; config: Record<string, string> }) =>
  api<ConnectorCreateResponse>("/api/v1/connectors", { method: "POST", body: payload });
export const updateConnector = (name: string, config: Record<string, string>) =>
  api<ConnectorDetail>("/api/v1/connectors/" + encodeURIComponent(name), { method: "PATCH", body: { config } });
export const deleteConnector = (name: string) =>
  api<void>("/api/v1/connectors/" + encodeURIComponent(name), { method: "DELETE" });
export const pauseConnector = (name: string) =>
  api<ConnectorDetail>("/api/v1/connectors/" + encodeURIComponent(name) + "/pause", { method: "POST" });
export const resumeConnector = (name: string) =>
  api<ConnectorDetail>("/api/v1/connectors/" + encodeURIComponent(name) + "/resume", { method: "POST" });
export const restartConnector = (name: string, includeTasks: boolean, onlyFailed: boolean) =>
  api<ConnectorDetail>(
    "/api/v1/connectors/" + encodeURIComponent(name) + "/restart" + qs({ include_tasks: includeTasks, only_failed: onlyFailed }),
    { method: "POST" },
  );
export const getTask = (name: string, taskId: number) =>
  api<TaskDetail>("/api/v1/connectors/" + encodeURIComponent(name) + "/tasks/" + taskId);
export const restartTask = (name: string, taskId: number) =>
  api<void>("/api/v1/connectors/" + encodeURIComponent(name) + "/tasks/" + taskId + "/restart", { method: "POST" });

// ---- plugins ----
export const getConnectorPlugins = () => api<PluginListItem[]>("/api/v1/connector-plugins");
export const getPluginSchema = (pluginId: string, importance?: string) =>
  api<PluginSchema>("/api/v1/connector-plugins/" + encodeURIComponent(pluginId) + "/schema" + qs({ importance }));
export const getSmtPlugins = () => api<PluginListItem[]>("/api/v1/smt-plugins");
export const getSmtPluginSchema = (pluginId: string, importance?: string) =>
  api<PluginSchema>("/api/v1/smt-plugins/" + encodeURIComponent(pluginId) + "/schema" + qs({ importance }));
export const validateConnector = (pluginType: string, name: string, config: Record<string, string>) =>
  api<ValidationReport>("/api/v1/connectors/validate", {
    method: "POST",
    body: { plugin_type: pluginType, name, config },
  });

// ---- audit logs (documented in swagger; may be absent on older servers) ----
export const getAuditLogs = (params: {
  page?: number;
  limit?: number;
  action?: string;
  connector?: string;
  actor?: string;
}) => api<AuditLogListResponse>("/api/v1/audit-logs" + qs(params));
export const getAuditLog = (id: string) => api<AuditLogEntry>("/api/v1/audit-logs/" + encodeURIComponent(id));
export const deleteAuditLog = (id: string) =>
  api<void>("/api/v1/audit-logs/" + encodeURIComponent(id), { method: "DELETE" });
