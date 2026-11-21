"use client";

import { useState, useEffect, useCallback } from "react";

import { searchLogsApi, type LogRecord } from "@/lib/api/log";
import { useAuthStore } from "@/store/auth";

function timeRangeToFromTo(range: string): { from: string; to: string } {
  const now = new Date();
  const hours = ({ "1h": 1, "6h": 6, "24h": 24, "7d": 168 } as Record<string, number>)[range] ?? 1;
  return {
    from: new Date(now.getTime() - hours * 3_600_000).toISOString(),
    to: now.toISOString(),
  };
}

export interface LogQuery {
  level: string;
  service: string;
  text: string;
  timeRange: string;
}

export function useLogs(projectId: string | null, query: LogQuery) {
  const isAuthReady = useAuthStore((s) => s.isAuthReady);
  const accessToken = useAuthStore((s) => s.accessToken);
  const [state, setState] = useState<{
    logs: LogRecord[];
    total: number;
    isLoading: boolean;
    isError: string | null;
  }>({ logs: [], total: 0, isLoading: true, isError: null });
  const [tick, setTick] = useState(0);
  const [debouncedText, setDebouncedText] = useState(query.text);

  useEffect(() => {
    const timer = setTimeout(() => setDebouncedText(query.text), 300);
    return () => clearTimeout(timer);
  }, [query.text]);

  useEffect(() => {
    if (!isAuthReady) return;
    if (!accessToken) {
      queueMicrotask(() => setState({ logs: [], total: 0, isLoading: false, isError: null }));
      return;
    }
    if (!projectId) return;

    const { from, to } = timeRangeToFromTo(query.timeRange);
    searchLogsApi(projectId, {
      ...(query.level ? { level: query.level } : {}),
      ...(query.service ? { service: query.service } : {}),
      ...(debouncedText ? { text: debouncedText } : {}),
      from,
      to,
    })
      .then(({ data }) =>
        setState({ logs: data.logs ?? [], total: data.total, isLoading: false, isError: null }),
      )
      .catch((err: Error) =>
        setState({ logs: [], total: 0, isLoading: false, isError: err.message }),
      );
  }, [
    isAuthReady,
    accessToken,
    projectId,
    query.level,
    query.service,
    query.timeRange,
    debouncedText,
    tick,
  ]);

  // Auto-refresh every 30 seconds
  useEffect(() => {
    const interval = setInterval(() => setTick((t) => t + 1), 30_000);
    return () => clearInterval(interval);
  }, []);

  const refresh = useCallback(() => setTick((t) => t + 1), []);
  return {
    logs: state.logs,
    total: state.total,
    loading: state.isLoading,
    error: state.isError,
    refresh,
  };
}
