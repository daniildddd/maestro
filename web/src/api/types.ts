// Types mirroring api/swagger.yaml (Maestro API v1)

export interface ErrorResponse {
  code: string;
  message: string;
}

export interface User {
  id: string;
  username: string;
  role: "admin" | "user";
  created_at: string;
  updated_at: string | null;
}

export interface PaginationMeta {
  page: number;
  limit: number;
}

export interface UserListResponse {
  data: User[];
  meta: PaginationMeta;
  has_more: boolean;
}

export interface LoginResponse {
  access_token: string;
  username: string;
}

export type ConnectorStatus = "running" | "paused" | "failed" | "starting";

export interface ConnectorTask {
  id: number;
  state: string;
  worker_id: string | null;
}

// Config as returned by the backend (flattened source config DTO).
export interface ConnectorSourceConfig {
  database_hostname?: string;
  database_port?: string;
  database_user?: string;
  database_dbname?: string | null;
  plugin_name?: string | null;
  [key: string]: string | null | undefined;
}

export interface ConnectorListItem {
  name: string;
  plugin_type: string | null;
  status: ConnectorStatus;
  tasks_count: number;
}

export interface ConnectorListResponse {
  data: ConnectorListItem[];
  meta: PaginationMeta;
  total: number;
}

export interface ConnectorDetail {
  name: string;
  plugin_type: string | null;
  config: ConnectorSourceConfig | null;
  status: ConnectorStatus;
  worker_id: string | null;
  tasks_count: number;
  tasks: ConnectorTask[] | null;
}

export interface ConnectorCreateResponse {
  name: string;
  plugin_type: string;
  status: ConnectorStatus;
  config: ConnectorSourceConfig;
  tasks_count: number;
  tasks: ConnectorTask[];
}

export interface PluginListItem {
  id: string;
}

export interface PluginSchemaField {
  name: string;
  label?: string | null;
  description?: string | null;
  type: string;
  importance: string;
  required: boolean;
  default?: string | null;
  values?: string[];
}

export interface PluginSchema {
  fields: PluginSchemaField[];
}

export type ValidationSeverity = "ok" | "warning" | "error" | "skipped";

export interface ValidationCheck {
  id: string;
  severity: ValidationSeverity;
  message: string;
  fix_hint?: string;
  field?: string;
  table?: string;
}

export interface ValidationStep {
  id: string;
  status: ValidationSeverity;
  checks: ValidationCheck[];
}

export interface ValidationReport {
  valid: boolean;
  steps: ValidationStep[];
}

export interface TaskDetail {
  id: number;
  state: string;
  worker_id: string | null;
  trace: string | null;
}

export interface AuditActor {
  id: string;
  username: string;
}

export interface AuditSubject {
  type: string;
  id: string;
  name: string;
}

export interface AuditRequest {
  id: string;
  ip: string;
  user_agent: string;
}

export interface AuditLogItem {
  id: string;
  created_at: string;
  action: string;
  outcome: string;
  failure_reason: string | null;
  actor: AuditActor | null;
  subject: AuditSubject;
  request: AuditRequest | null;
}

export interface AuditLogEntry extends AuditLogItem {
  state_before?: Record<string, unknown> | null;
  state_after?: Record<string, unknown> | null;
}

export interface AuditLogListResponse {
  data: AuditLogItem[];
  meta: PaginationMeta;
  has_more: boolean;
}
