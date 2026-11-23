"use client";

import { useCallback, useEffect, useState } from "react";

import { getDashboardStatsApi, type DashboardStats } from "@/lib/api/dashboard";
import { useAuthStore } from "@/store/auth";

function timeRangeToFromTo(range: string): { from: string; to: string } {
  const now = new Date();
  const hours = ({ "1h": 1, "6h": 6, "24h": 24, "7d": 168 } as Record<string, number>)[range] ?? 24;
  return {
    from: new Date(now.getTime() - hours * 3_600_000).toISOString(),
    to: now.toISOString(),
  };
}

export function useDashboard(projectId: string | null, timeRange: string) {
  const isAuthReady = useAuthStore((s) => s.isAuthReady);
  const accessToken = useAuthStore((s) => s.accessToken);
  const [state, setState] = useState<{
    stats: DashboardStats | null;
    isLoading: boolean;
    isError: string | null;
  }>({ stats: null, isLoading: true, isError: null });
  const [tick, setTick] = useState(0);

  useEffect(() => {
    if (!isAuthReady) return;
    if (!accessToken) {
      queueMicrotask(() => setState({ stats: null, isLoading: false, isError: null }));
      return;
    }
    if (!projectId) return;

    const { from, to } = timeRangeToFromTo(timeRange);
    getDashboardStatsApi(projectId, from, to)
      .then(({ data }) => setState({ stats: data, isLoading: false, isError: null }))
      .catch((err: Error) => setState({ stats: null, isLoading: false, isError: err.message }));
  }, [isAuthReady, accessToken, projectId, timeRange, tick]);

  const refresh = useCallback(() => setTick((t) => t + 1), []);
  return {
    stats: state.stats,
    loading: state.isLoading,
    error: state.isError,
    refresh,
  };
}
