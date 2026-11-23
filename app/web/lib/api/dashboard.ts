import { apiRequest } from "@/lib/api/client";

export interface VolumePoint {
  time: string;
  count: number;
}

export interface LevelBreakdown {
  level: string;
  count: number;
}

export interface RecentError {
  timestamp: string;
  level: string;
  service: string;
  message: string;
  environment: string;
}

export interface DashboardStats {
  total_logs: number;
  error_count: number;
  warn_count: number;
  volume: VolumePoint[];
  level_breakdown: LevelBreakdown[];
  recent_errors: RecentError[];
  from: string;
  to: string;
}

export function getDashboardStatsApi(projectId: string, from?: string, to?: string) {
  const qs = new URLSearchParams();
  if (from) qs.set("from", from);
  if (to) qs.set("to", to);
  const q = qs.toString();
  return apiRequest<{ data: DashboardStats }>(`dashboard/${projectId}${q ? `?${q}` : ""}`);
}
