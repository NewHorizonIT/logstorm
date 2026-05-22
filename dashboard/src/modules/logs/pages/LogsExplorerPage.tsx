import { Fragment, useMemo, useState } from "react";
import { SearchInput, Select } from "@/components/Input";
import { Card } from "@/components/Card";

type LogRow = {
  timestamp: string;
  service: string;
  level: "error" | "warn" | "info" | "debug" | "fatal";
  message: string;
};

const mockLogs: LogRow[] = [
  {
    timestamp: "2026-03-24 11:21:32.113",
    service: "api-gateway",
    level: "info",
    message: "request completed in 42ms",
  },
  {
    timestamp: "2026-03-24 11:21:15.844",
    service: "payment-service",
    level: "warn",
    message: "retrying payment provider call (attempt 2)",
  },
  {
    timestamp: "2026-03-24 11:20:58.291",
    service: "auth-service",
    level: "error",
    message: "token validation failed: invalid signature",
  },
  {
    timestamp: "2026-03-24 11:20:39.006",
    service: "notification-service",
    level: "info",
    message: "queued email event to kafka topic notifications.v1",
  },
  {
    timestamp: "2026-03-24 11:20:02.700",
    service: "inventory-service",
    level: "warn",
    message: "cache miss ratio exceeded threshold",
  },
];

const levelClass: Record<LogRow["level"], string> = {
  fatal: "text-rose-300",
  error: "text-rose-300",
  warn: "text-amber-300",
  info: "text-cyan-300",
  debug: "text-slate-400",
};

const timeRangeOptions = ["Last 15 minutes", "Last 1 hour", "Last 6 hours", "Last 24 hours"];
const serviceOptions = ["All services", "api-gateway", "payment-service", "auth-service", "notification-service"];
const levelOptions = ["All levels", "error", "warn", "info"];

const LogsExplorerPage = () => {
  const [openRowId, setOpenRowId] = useState<string | null>(null);
  const [search, setSearch] = useState("");
  const [timeRange, setTimeRange] = useState(timeRangeOptions[0]);
  const [service, setService] = useState(serviceOptions[0]);
  const [level, setLevel] = useState(levelOptions[0]);

  const toggleRow = (rowId: string) => {
    setOpenRowId((prev) => (prev === rowId ? null : rowId));
  };

  const filteredLogs = useMemo(() => {
    return mockLogs.filter((log) => {
      const traceId = `trace-${log.timestamp.replace(/[^a-zA-Z0-9]/g, "").slice(0, 12)}-${log.service}`;

      if (service !== "All services" && log.service !== service) return false;
      if (level !== "All levels" && log.level !== level) return false;

      const q = search.trim().toLowerCase();
      if (!q) return true;

      return (
        log.message.toLowerCase().includes(q) ||
        log.service.toLowerCase().includes(q) ||
        traceId.toLowerCase().includes(q)
      );
    });
  }, [search, service, level]);

  return (
    <div className="space-y-4">
      <Card>
        <SearchInput value={search} onChange={(v) => setSearch(v)} placeholder="Search logs, trace_id, service..." />
      </Card>

      <div className="grid gap-4 lg:grid-cols-[280px_1fr]">
        <Card className="h-fit">
          <h2 className="mb-4 text-sm font-semibold text-[--color-text-primary]">Filters</h2>

          <div className="space-y-3">
            <div>
              <label className="mb-1 block text-xs text-[--color-text-secondary]">Time range</label>
              <Select value={timeRange} onChange={(e) => setTimeRange(e.target.value)}>
                {timeRangeOptions.map((option) => (
                  <option key={option} value={option}>
                    {option}
                  </option>
                ))}
              </Select>
            </div>

            <div>
              <label className="mb-1 block text-xs text-[--color-text-secondary]">Service</label>
              <Select value={service} onChange={(e) => setService(e.target.value)}>
                {serviceOptions.map((option) => (
                  <option key={option} value={option}>
                    {option}
                  </option>
                ))}
              </Select>
            </div>

            <div>
              <label className="mb-1 block text-xs text-[--color-text-secondary]">Log level</label>
              <Select value={level} onChange={(e) => setLevel(e.target.value)}>
                {levelOptions.map((option) => (
                  <option key={option} value={option}>
                    {option}
                  </option>
                ))}
              </Select>
            </div>
          </div>
        </Card>

        <Card>
          <div className="mb-4 flex items-center justify-between">
            <h2 className="text-sm font-semibold text-[--color-text-primary]">Logs</h2>
            <span className="text-xs text-[--color-text-secondary]">{filteredLogs.length} rows</span>
          </div>

          <div className="overflow-x-auto rounded-[--radius-md] border border-[--color-border]">
            <table className="w-full min-w-120 text-left text-sm">
              <thead className="bg-[color-mix(in_srgb,var(--color-card),white_5%)] text-[--color-text-secondary]">
                <tr>
                  <th className="px-4 py-3 font-medium">Timestamp</th>
                  <th className="px-4 py-3 font-medium">Service</th>
                  <th className="px-4 py-3 font-medium">Level</th>
                  <th className="px-4 py-3 font-medium">Message</th>
                </tr>
              </thead>
              <tbody>
                {filteredLogs.map((log) => {
                  const rowId = `${log.timestamp}-${log.service}`;
                  const isOpen = openRowId === rowId;

                  const logDetails = {
                    timestamp: log.timestamp,
                    service: log.service,
                    level: log.level,
                    message: log.message,
                    trace_id: `trace-${rowId.replace(/[^a-zA-Z0-9]/g, "").slice(0, 12)}`,
                    metadata: {
                      environment: "production",
                      region: "ap-southeast-1",
                    },
                  };

                  return (
                    <Fragment key={rowId}>
                      <tr key={rowId} className="border-t border-[--color-border]">
                        <td className="px-4 py-3 text-[--color-text-secondary]">{log.timestamp}</td>
                        <td className="px-4 py-3 text-[--color-text-primary]">{log.service}</td>
                        <td className="px-4 py-3 uppercase tracking-wide">
                          <span className={levelClass[log.level]}>{log.level}</span>
                        </td>
                        <td className="px-4 py-3 text-[--color-text-secondary]">
                          <div className="flex items-center justify-between gap-3">
                            <span>{log.message}</span>
                            <button
                              type="button"
                              className="shrink-0 rounded-[--radius-md] border border-[--color-border] px-2 py-1 text-xs text-[--color-text-primary]"
                              onClick={() => toggleRow(rowId)}
                            >
                              {isOpen ? "▼ Close" : "▶ Open"}
                            </button>
                          </div>
                        </td>
                      </tr>

                      {isOpen && (
                        <tr className="border-t border-[--color-border] bg-[color-mix(in_srgb,var(--color-card),white_3%)]">
                          <td colSpan={4} className="p-4">
                            <div className="mb-2 flex items-center justify-between gap-4">
                              <div className="text-xs text-[--color-text-secondary]">trace_id: <span className="ml-2 font-mono text-[--color-text-primary]">{logDetails.trace_id}</span></div>
                              <button
                                type="button"
                                className="rounded-[--radius-md] border border-[--color-border] px-2 py-1 text-xs"
                                onClick={() => navigator.clipboard?.writeText(logDetails.trace_id)}
                              >
                                Copy trace_id
                              </button>
                            </div>
                            <pre className="overflow-x-auto rounded-[--radius-md] border border-[--color-border] bg-[--color-bg] p-3 text-xs text-[--color-text-secondary]">
                              {JSON.stringify(logDetails, null, 2)}
                            </pre>
                          </td>
                        </tr>
                      )}
                    </Fragment>
                  );
                })}
              </tbody>
            </table>
          </div>
        </Card>
      </div>
    </div>
  );
};

export default LogsExplorerPage;
