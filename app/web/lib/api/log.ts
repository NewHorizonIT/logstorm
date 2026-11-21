import { apiRequest } from "@/lib/api/client";

export interface LogRecord {
  id: string;
  timestamp: string;
  project_id: string;
  environment: string;
  level: string;
  service: string;
  message: string;
  trace_id: string;
  span_id: string;
  source: string;
  attributes: Record<string, string> | null;
}

export interface LogSearchParams {
  level?: string;
  service?: string;
  text?: string;
  from?: string;
  to?: string;
  limit?: number;
  offset?: number;
}

export interface LogSearchResult {
  logs: LogRecord[] | null;
  total: number;
  limit: number;
  offset: number;
}

export function searchLogsApi(projectId: string, params: LogSearchParams) {
  const qs = new URLSearchParams();
  if (params.level) qs.set("level", params.level);
  if (params.service) qs.set("service", params.service);
  if (params.text) qs.set("text", params.text);
  if (params.from) qs.set("from", params.from);
  if (params.to) qs.set("to", params.to);
  if (params.limit !== undefined) qs.set("limit", String(params.limit));
  if (params.offset !== undefined) qs.set("offset", String(params.offset));
  const q = qs.toString();
  return apiRequest<{ data: LogSearchResult }>(`/search/${projectId}/logs${q ? `?${q}` : ""}`);
}
