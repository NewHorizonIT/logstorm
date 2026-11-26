"use client";

import { useState, useEffect, useRef, useCallback } from "react";

import { type LogRecord } from "@/lib/api/log";
import { useAuthStore } from "@/store/auth";

const API_BASE = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

const BUFFER_MAX = 200;

const BACKOFF_BASE_MS = 1_000;
const BACKOFF_MAX_MS = 30_000;

export type LiveTailStatus = "connecting" | "connected" | "disconnected" | "error";

export interface LiveTailFilters {
  level?: string;
  service?: string;
}

export interface UseLiveTailResult {
  logs: LogRecord[];
  status: LiveTailStatus;
  error: string | null;
  isPaused: boolean;
  pause: () => void;
  resume: () => void;
  clear: () => void;
  reconnect: () => void;
}

function parseSseFrames(buffer: string): [Array<{ event: string; data: string }>, string] {
  const frames = buffer.split("\n\n");
  // Phần tử cuối là frame chưa kết thúc — giữ lại cho chunk tiếp theo
  const remaining = frames.pop() ?? "";
  const events: Array<{ event: string; data: string }> = [];

  for (const frame of frames) {
    if (!frame.trim()) continue;

    let eventType = "message";
    let data = "";

    for (const line of frame.split("\n")) {
      if (line.startsWith("event:")) {
        eventType = line.slice(6).trim();
      } else if (line.startsWith("data:")) {
        data = line.slice(5).trim();
      }
    }

    if (data) {
      events.push({ event: eventType, data });
    }
  }

  return [events, remaining];
}

export function useLiveTail(
  projectId: string | null,
  filters: LiveTailFilters = {},
): UseLiveTailResult {
  const isAuthReady = useAuthStore((s) => s.isAuthReady);
  const accessToken = useAuthStore((s) => s.accessToken);

  const [logs, setLogs] = useState<LogRecord[]>([]);
  const [status, setStatus] = useState<LiveTailStatus>("disconnected");
  const [error, setError] = useState<string | null>(null);
  const [isPaused, setIsPaused] = useState(false);
  const [reconnectKey, setReconnectKey] = useState(0);

  const pendingRef = useRef<LogRecord[]>([]);
  const isPausedRef = useRef(false);
  const retryCountRef = useRef(0);

  const pause = useCallback(() => {
    isPausedRef.current = true;
    setIsPaused(true);
  }, []);

  const resume = useCallback(() => {
    isPausedRef.current = false;
    setIsPaused(false);
    // Xả buffer pending vào state khi resume
    if (pendingRef.current.length > 0) {
      setLogs((prev) => {
        const merged = [...prev, ...pendingRef.current];
        pendingRef.current = [];
        return merged.slice(-BUFFER_MAX);
      });
    }
  }, []);

  const clear = useCallback(() => {
    pendingRef.current = [];
    setLogs([]);
  }, []);

  const reconnect = useCallback(() => {
    retryCountRef.current = 0;
    setReconnectKey((k) => k + 1);
  }, []);

  useEffect(() => {
    if (!isAuthReady || !accessToken || !projectId) {
      queueMicrotask(() => setStatus("disconnected"));
      return;
    }

    let retryTimeout: ReturnType<typeof setTimeout> | null = null;
    const controller = new AbortController();
    const { signal } = controller;

    const qs = new URLSearchParams();
    if (filters.level) qs.set("level", filters.level);
    if (filters.service) qs.set("service", filters.service);
    const query = qs.toString();
    const url = `${API_BASE}/api/v1/livetail/${projectId}${query ? `?${query}` : ""}`;

    // scheduleReconnect lên lịch kết nối lại với exponential backoff
    function scheduleReconnect() {
      if (signal.aborted) return;
      const delay = Math.min(BACKOFF_MAX_MS, BACKOFF_BASE_MS * Math.pow(2, retryCountRef.current));
      retryCountRef.current += 1;
      retryTimeout = setTimeout(() => {
        if (!signal.aborted) void connect();
      }, delay);
    }

    // connect thực hiện một lần kết nối SSE, đọc stream, gọi scheduleReconnect khi lỗi
    async function connect() {
      if (signal.aborted) return;

      setStatus("connecting");
      setError(null);

      let textBuffer = "";

      try {
        const response = await fetch(url, {
          headers: {
            Authorization: `Bearer ${accessToken}`,
            Accept: "text/event-stream",
          },
          signal,
        });

        if (!response.ok) {
          const text = await response.text().catch(() => `HTTP ${response.status}`);
          setStatus("error");
          setError(text || `HTTP ${response.status}`);
          scheduleReconnect();
          return;
        }

        if (!response.body) {
          setStatus("error");
          setError("Response body is null");
          scheduleReconnect();
          return;
        }

        // Connect thành công — reset retry count
        setStatus("connected");
        retryCountRef.current = 0;

        const reader = response.body.getReader();
        const decoder = new TextDecoder();

        // Đọc stream cho đến khi done (server đóng) hoặc bị abort (cleanup)
        for (;;) {
          const { done, value } = await reader.read();
          if (done) break;

          textBuffer += decoder.decode(value, { stream: true });
          const [events, remaining] = parseSseFrames(textBuffer);
          textBuffer = remaining;

          for (const { event, data } of events) {
            // Chỉ xử lý event "log" — bỏ qua heartbeat và unknown event
            if (event !== "log") continue;
            try {
              const record = JSON.parse(data) as LogRecord;
              if (isPausedRef.current) {
                // Buffer log khi pause — không re-render, không mất event
                pendingRef.current = [...pendingRef.current, record].slice(-BUFFER_MAX);
              } else {
                setLogs((prev) => [...prev, record].slice(-BUFFER_MAX));
              }
            } catch {
              // JSON parse error — bỏ qua event (live tail là best-effort)
            }
          }
        }

        // Stream kết thúc bình thường (server đóng connection) — reconnect
        setStatus("disconnected");
        scheduleReconnect();
      } catch (err) {
        // AbortError xảy ra khi cleanup chủ động huỷ → không reconnect
        if ((err as Error).name === "AbortError") return;
        setStatus("error");
        setError((err as Error).message);
        scheduleReconnect();
      }
    }

    void connect();

    return () => {
      controller.abort();
      if (retryTimeout) clearTimeout(retryTimeout);
    };
  }, [isAuthReady, accessToken, projectId, filters.level, filters.service, reconnectKey]);

  return { logs, status, error, isPaused, pause, resume, clear, reconnect };
}
