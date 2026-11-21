"use client";

import { Fragment, useState } from "react";
import { ArrowClockwise, CaretDown, CaretRight } from "@phosphor-icons/react";

import { type LogRecord } from "@/lib/api/log";
import { useLogs } from "@/lib/hooks/use-logs";
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

function TableSkeleton() {
  return (
    <TableBody>
      {Array.from({ length: 6 }).map((_, i) => (
        <TableRow key={i}>
          <TableCell>
            <div className="bg-surface-active h-3 w-32 animate-pulse rounded" />
          </TableCell>
          <TableCell>
            <div className="bg-surface-active rounded-badge h-5 w-12 animate-pulse" />
          </TableCell>
          <TableCell>
            <div className="bg-surface-active h-3 w-24 animate-pulse rounded" />
          </TableCell>
          <TableCell>
            <div className="bg-surface-active h-3 w-96 animate-pulse rounded" />
          </TableCell>
        </TableRow>
      ))}
    </TableBody>
  );
}

interface ExpandedRowProps {
  log: LogRecord;
}

function ExpandedRow({ log }: ExpandedRowProps) {
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

export function LogsContent({ slug }: Props) {
  const { project, loading: projectLoading, error: projectError } = useProject(slug);

  const [level, setLevel] = useState("");
  const [service, setService] = useState("");
  const [text, setText] = useState("");
  const [timeRange, setTimeRange] = useState("1h");
  const [expandedId, setExpandedId] = useState<string | null>(null);

  const {
    logs,
    total,
    loading: logsLoading,
    error: logsError,
    refresh,
  } = useLogs(project?.id ?? null, { level, service, text, timeRange });

  if (projectLoading) {
    return (
      <div className="flex h-full flex-col">
        <div className="border-border-divider shrink-0 border-b p-6">
          <div className="bg-surface-active h-7 w-16 animate-pulse rounded" />
          <div className="bg-surface-active mt-2 h-4 w-48 animate-pulse rounded" />
        </div>
        <div className="border-border-divider flex shrink-0 items-center gap-2 border-b px-6 py-3">
          {Array.from({ length: 4 }).map((_, i) => (
            <div key={i} className="bg-surface-active h-8 w-32 animate-pulse rounded" />
          ))}
        </div>
        <div className="min-h-0 flex-1 overflow-auto">
          <Table>
            <TableHeader>
              <TableRow>
                {["Timestamp", "Level", "Service", "Message"].map((h) => (
                  <TableHead key={h}>{h}</TableHead>
                ))}
              </TableRow>
            </TableHeader>
            <TableSkeleton />
          </Table>
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
          <h1 className="text-text-primary text-xl font-semibold">Logs</h1>
          <p className="text-text-muted mt-0.5 font-mono text-xs">{project.name}</p>
        </div>
        {!logsLoading && (
          <span className="text-text-muted text-xs">
            {total.toLocaleString()} result{total !== 1 ? "s" : ""}
          </span>
        )}
      </div>

      {/* Filter bar */}
      <div className="border-border-divider flex shrink-0 items-center gap-2 border-b px-6 py-3">
        <div className="relative flex-1">
          <Input
            placeholder="Search messages…"
            value={text}
            onChange={(e) => setText(e.target.value)}
            className="h-8 pl-8 text-xs"
          />
          <span className="text-text-disabled pointer-events-none absolute top-1/2 left-2.5 -translate-y-1/2">
            <svg width="13" height="13" viewBox="0 0 256 256" fill="currentColor">
              <path d="M229.66,218.34l-50.07-50.07a88.11,88.11,0,1,0-11.31,11.31l50.06,50.07a8,8,0,0,0,11.32-11.31ZM40,112a72,72,0,1,1,72,72A72.08,72.08,0,0,1,40,112Z" />
            </svg>
          </span>
        </div>
        <div className="w-32">
          <Select value={level} onChange={(e) => setLevel(e.target.value)}>
            <option value="">All levels</option>
            <option value="debug">debug</option>
            <option value="info">info</option>
            <option value="warn">warn</option>
            <option value="error">error</option>
            <option value="fatal">fatal</option>
          </Select>
        </div>
        <Input
          placeholder="Service…"
          value={service}
          onChange={(e) => setService(e.target.value)}
          className="h-8 w-36 font-mono text-xs"
        />
        <div className="w-36">
          <Select value={timeRange} onChange={(e) => setTimeRange(e.target.value)}>
            <option value="1h">Last 1 hour</option>
            <option value="6h">Last 6 hours</option>
            <option value="24h">Last 24 hours</option>
            <option value="7d">Last 7 days</option>
          </Select>
        </div>
        <Button variant="outline" size="sm" onClick={refresh} className="shrink-0 gap-1.5">
          <ArrowClockwise size={13} />
          Refresh
        </Button>
      </div>

      {/* Table */}
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

          {logsLoading ? (
            <TableSkeleton />
          ) : logsError ? (
            <TableBody>
              <TableRow>
                <TableCell colSpan={4} className="text-status-danger py-10 text-center text-sm">
                  {logsError}
                </TableCell>
              </TableRow>
            </TableBody>
          ) : logs.length === 0 ? (
            <TableBody>
              <TableRow>
                <TableCell colSpan={4} className="text-text-muted py-16 text-center text-sm">
                  No logs found for this time range and filters.
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
      </div>
    </div>
  );
}
