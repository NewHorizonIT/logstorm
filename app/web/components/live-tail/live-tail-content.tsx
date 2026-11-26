"use client";

import { Fragment, useCallback, useEffect, useRef, useState } from "react";
import {
  ArrowClockwise,
  Broadcast,
  CaretDown,
  CaretRight,
  Pause,
  Play,
  Trash,
} from "@phosphor-icons/react";

import { type LogRecord } from "@/lib/api/log";
import { useLiveTail, type LiveTailFilters, type LiveTailStatus } from "@/lib/hooks/use-live-tail";
import { useProject } from "@/lib/hooks/use-project";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Select } from "@/components/ui/select";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { cn } from "@/lib/utils";

// LevelVariant map từ log level sang Badge variant — giống logs-content.tsx
type LevelVariant = "neutral" | "info" | "warning" | "danger";

function levelVariant(level: string): LevelVariant {
  switch (level) {
    case "info":
      return "info";
    case "warn":
      return "warning";
    case "error":
    case "fatal":
      return "danger";
    default:
      return "neutral";
  }
}

// formatTimestamp hiển thị thời gian theo định dạng "Nov 26 14:30:00"
function formatTimestamp(iso: string) {
  const d = new Date(iso);
  return (
    d.toLocaleDateString("en-US", { month: "short", day: "numeric" }) +
    " " +
    d.toLocaleTimeString("en-US", {
      hour: "2-digit",
      minute: "2-digit",
      second: "2-digit",
      hour12: false,
    })
  );
}

// StatusDot hiển thị indicator trạng thái kết nối SSE dưới dạng vòng tròn nhỏ.
//
// connected    → xanh (#14b797)
// connecting   → vàng nhấp nháy (#fbbf24)
// disconnected → xám (#9ca3af)
// error        → đỏ (#f87171)
function StatusDot({ status }: { status: LiveTailStatus }) {
  return (
    <span
      className={cn(
        "inline-block size-2 shrink-0 rounded-full",
        status === "connected" && "bg-status-success",
        status === "connecting" && "bg-status-warning animate-pulse",
        status === "disconnected" && "bg-status-neutral",
        status === "error" && "bg-status-danger",
      )}
    />
  );
}

// statusLabel trả về chuỗi mô tả trạng thái hiển thị cạnh StatusDot
function statusLabel(status: LiveTailStatus): string {
  switch (status) {
    case "connected":
      return "Connected";
    case "connecting":
      return "Connecting…";
    case "disconnected":
      return "Disconnected";
    case "error":
      return "Error";
  }
}

// ExpandedRow hiển thị chi tiết log khi click vào row
function ExpandedRow({ log }: { log: LogRecord }) {
  const attrs = log.attributes ? Object.entries(log.attributes) : [];
  return (
    <TableRow className="bg-surface hover:bg-surface">
      <TableCell colSpan={4} className="py-3">
        <div className="space-y-2 pl-2 font-mono text-xs">
          <div>
            <span className="text-text-muted">message </span>
            <span className="text-text-primary break-all">{log.message}</span>
          </div>
          {log.trace_id && (
            <div>
              <span className="text-text-muted">trace_id </span>
              <span className="text-text-subtle">{log.trace_id}</span>
            </div>
          )}
          {log.span_id && (
            <div>
              <span className="text-text-muted">span_id </span>
              <span className="text-text-subtle">{log.span_id}</span>
            </div>
          )}
          {log.environment && (
            <div>
              <span className="text-text-muted">environment </span>
              <span className="text-text-subtle">{log.environment}</span>
            </div>
          )}
          {log.source && (
            <div>
              <span className="text-text-muted">source </span>
              <span className="text-text-subtle">{log.source}</span>
            </div>
          )}
          {attrs.length > 0 && (
            <div>
              <div className="text-text-muted mb-1">attributes</div>
              <div className="space-y-1 pl-4">
                {attrs.map(([k, v]) => (
                  <div key={k}>
                    <span className="text-accent-brand">{k}</span>
                    <span className="text-text-muted">: </span>
                    <span className="text-text-primary">{v}</span>
                  </div>
                ))}
              </div>
            </div>
          )}
        </div>
      </TableCell>
    </TableRow>
  );
}

interface Props {
  slug: string;
}

// LiveTailContent là component chính của trang Live Tail.
//
// Trách nhiệm:
//   - Resolve project từ slug, lấy project.id để truyền vào useLiveTail.
//   - Render toolbar: pause/resume, reconnect (khi lỗi), level/service filter, clear.
//   - Render bảng log stream với expand row khi click.
//   - Auto-scroll xuống cuối khi có log mới (trừ khi đang pause).
export function LiveTailContent({ slug }: Props) {
  const { project, loading: projectLoading, error: projectError } = useProject(slug);

  const [level, setLevel] = useState("");
  const [serviceInput, setServiceInput] = useState(""); // raw value bound to input
  const [service, setService] = useState(""); // debounced value used in filter
  const [expandedId, setExpandedId] = useState<string | null>(null);

  const filters: LiveTailFilters = {
    ...(level ? { level } : {}),
    ...(service ? { service } : {}),
  };

  const { logs, status, error, isPaused, pause, resume, clear, reconnect } = useLiveTail(
    project?.id ?? null,
    filters,
  );

  // bottomRef là anchor để auto-scroll, đặt sau bảng log
  const bottomRef = useRef<HTMLDivElement>(null);

  // Auto-scroll xuống cuối khi có log mới và không pause
  useEffect(() => {
    if (!isPaused && logs.length > 0) {
      bottomRef.current?.scrollIntoView({ behavior: "smooth" });
    }
  }, [logs, isPaused]);

  // Debounce service filter 400ms — tránh reconnect SSE trên từng keystroke
  useEffect(() => {
    const timer = setTimeout(() => {
      setService(serviceInput);
      clear();
    }, 400);
    return () => clearTimeout(timer);
  }, [serviceInput, clear]);

  // Khi level thay đổi, clear buffer để tránh hiển thị log cũ không khớp filter
  const handleLevelChange = useCallback(
    (e: React.ChangeEvent<HTMLSelectElement>) => {
      setLevel(e.target.value);
      clear();
    },
    [clear],
  );

  const handleServiceChange = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    setServiceInput(e.target.value);
  }, []);

  if (projectLoading) {
    return (
      <div className="flex h-full flex-col">
        <div className="border-border-divider shrink-0 border-b p-6">
          <div className="bg-surface-active h-7 w-16 animate-pulse rounded" />
          <div className="bg-surface-active mt-2 h-4 w-48 animate-pulse rounded" />
        </div>
      </div>
    );
  }

  if (projectError || !project) {
    return (
      <div className="p-8">
        <p className="text-status-danger text-sm">{projectError ?? "Project not found."}</p>
      </div>
    );
  }

  return (
    <div className="flex h-full flex-col">
      {/* Header */}
      <div className="border-border-divider flex shrink-0 items-center justify-between border-b px-6 py-4">
        <div>
          <h1 className="text-text-primary flex items-center gap-2 text-xl font-semibold">
            <Broadcast size={20} weight="duotone" className="text-accent-brand" />
            Live Tail
          </h1>
          <p className="text-text-muted mt-0.5 font-mono text-xs">{project.name}</p>
        </div>

        {/* Trạng thái kết nối + số log đang hiển thị */}
        <div className="flex items-center gap-2">
          <StatusDot status={status} />
          <span className="text-text-muted text-xs">{statusLabel(status)}</span>
          {logs.length > 0 && (
            <span className="text-text-disabled text-xs">
              · {logs.length} log{logs.length !== 1 ? "s" : ""}
            </span>
          )}
          {isPaused && (
            <span className="bg-status-warning-bg text-status-warning rounded-badge px-2 py-0.5 text-[11px] font-medium">
              Paused
            </span>
          )}
        </div>
      </div>

      {/* Toolbar */}
      <div className="border-border-divider flex shrink-0 items-center gap-2 border-b px-6 py-3">
        {/* Pause / Resume */}
        {isPaused ? (
          <Button variant="outline" size="sm" onClick={resume} className="shrink-0 gap-1.5">
            <Play size={12} weight="fill" />
            Resume
          </Button>
        ) : (
          <Button variant="outline" size="sm" onClick={pause} className="shrink-0 gap-1.5">
            <Pause size={12} weight="fill" />
            Pause
          </Button>
        )}

        {/* Reconnect — chỉ hiện khi disconnected hoặc error */}
        {(status === "disconnected" || status === "error") && (
          <Button variant="outline" size="sm" onClick={reconnect} className="shrink-0 gap-1.5">
            <ArrowClockwise size={12} />
            Reconnect
          </Button>
        )}

        <div className="border-border-divider h-5 w-px shrink-0 border-l" />

        {/* Level filter */}
        <div className="w-32">
          <Select value={level} onChange={handleLevelChange}>
            <option value="">All levels</option>
            <option value="debug">debug</option>
            <option value="info">info</option>
            <option value="warn">warn</option>
            <option value="error">error</option>
            <option value="fatal">fatal</option>
          </Select>
        </div>

        {/* Service filter */}
        <Input
          placeholder="Service…"
          value={serviceInput}
          onChange={handleServiceChange}
          className="h-8 w-36 font-mono text-xs"
        />

        <div className="flex-1" />

        {/* Clear buffer */}
        <Button
          variant="ghost"
          size="sm"
          onClick={clear}
          disabled={logs.length === 0}
          className="text-text-muted hover:text-text-primary shrink-0 gap-1.5"
        >
          <Trash size={13} />
          Clear
        </Button>
      </div>

      {/* Error banner — hiện khi có lỗi kết nối */}
      {error && (
        <div className="bg-status-danger-bg text-status-danger border-status-danger/20 shrink-0 border-b px-6 py-2 font-mono text-xs">
          {error}
        </div>
      )}

      {/* Bảng log stream */}
      <div className="min-h-0 flex-1 overflow-auto">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead className="w-44">Timestamp</TableHead>
              <TableHead className="w-20">Level</TableHead>
              <TableHead className="w-40">Service</TableHead>
              <TableHead>Message</TableHead>
            </TableRow>
          </TableHeader>

          {logs.length === 0 ? (
            <TableBody>
              <TableRow>
                <TableCell colSpan={4} className="py-24 text-center">
                  <div className="flex flex-col items-center gap-3">
                    <Broadcast size={28} weight="duotone" className="text-text-disabled" />
                    <p className="text-text-muted text-sm">
                      {status === "connecting"
                        ? "Connecting to log stream…"
                        : status === "connected"
                          ? "Waiting for logs…"
                          : "Connect to start streaming logs."}
                    </p>
                  </div>
                </TableCell>
              </TableRow>
            </TableBody>
          ) : (
            <TableBody>
              {logs.map((log) => {
                const isExpanded = expandedId === log.id;
                return (
                  <Fragment key={log.id}>
                    <TableRow
                      className={cn("cursor-pointer select-none", isExpanded && "bg-surface")}
                      onClick={() => setExpandedId(isExpanded ? null : log.id)}
                    >
                      <TableCell className="text-text-muted font-mono text-xs whitespace-nowrap">
                        <span className="mr-1 inline-block w-3">
                          {isExpanded ? (
                            <CaretDown size={11} className="text-text-disabled" />
                          ) : (
                            <CaretRight size={11} className="text-text-disabled" />
                          )}
                        </span>
                        {formatTimestamp(log.timestamp)}
                      </TableCell>
                      <TableCell>
                        <Badge variant={levelVariant(log.level)}>{log.level}</Badge>
                      </TableCell>
                      <TableCell className="text-text-subtle font-mono text-xs">
                        {log.service}
                      </TableCell>
                      <TableCell className="text-text-primary max-w-0 truncate text-sm">
                        {log.message}
                      </TableCell>
                    </TableRow>
                    {isExpanded && <ExpandedRow log={log} />}
                  </Fragment>
                );
              })}
            </TableBody>
          )}
        </Table>

        {/* Scroll anchor — bottomRef trỏ tới đây để auto-scroll */}
        <div ref={bottomRef} />
      </div>
    </div>
  );
}
