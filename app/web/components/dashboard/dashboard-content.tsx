"use client";

import { useState } from "react";
import { ArrowClockwise } from "@phosphor-icons/react";

import { type LevelBreakdown, type RecentError, type VolumePoint } from "@/lib/api/dashboard";
import { useDashboard } from "@/lib/hooks/use-dashboard";
import { useProject } from "@/lib/hooks/use-project";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Select } from "@/components/ui/select";
import { cn } from "@/lib/utils";

type TimeRange = "1h" | "6h" | "24h" | "7d";

// ─── Helpers ────────────────────────────────────────────────────────────────

function formatNumber(n: number): string {
  if (n >= 1_000_000) return `${(n / 1_000_000).toFixed(1)}M`;
  if (n >= 1_000) return `${(n / 1_000).toFixed(1)}K`;
  return n.toLocaleString();
}

function formatVolumeLabel(iso: string): string {
  const d = new Date(iso);
  if (d.getHours() === 0 && d.getMinutes() === 0) {
    return d.toLocaleDateString("en-US", { month: "short", day: "numeric" });
  }
  return (
    d.toLocaleDateString("en-US", { month: "short", day: "numeric" }) +
    " " +
    d.toLocaleTimeString("en-US", { hour: "2-digit", minute: "2-digit", hour12: false })
  );
}

function relativeTime(iso: string): string {
  const diff = Date.now() - new Date(iso).getTime();
  const mins = Math.floor(diff / 60_000);
  if (mins < 1) return "just now";
  if (mins < 60) return `${mins}m ago`;
  const hours = Math.floor(mins / 60);
  if (hours < 24) return `${hours}h ago`;
  return `${Math.floor(hours / 24)}d ago`;
}

function levelBadgeVariant(level: string): "info" | "warning" | "danger" | "neutral" {
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

const LEVEL_BAR_CLASS: Record<string, string> = {
  debug: "bg-zinc-400/70",
  info: "bg-accent-brand/70",
  warn: "bg-yellow-400/80",
  error: "bg-red-500/80",
  fatal: "bg-red-700/80",
};

// ─── Stat Card ───────────────────────────────────────────────────────────────

interface StatCardProps {
  label: string;
  value: number;
  loading: boolean;
  valueClass?: string;
}

function StatCard({ label, value, loading, valueClass }: StatCardProps) {
  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-text-muted text-xs font-medium tracking-wider uppercase">
          {label}
        </CardTitle>
      </CardHeader>
      <CardContent>
        {loading ? (
          <div className="bg-surface-active h-8 w-24 animate-pulse rounded" />
        ) : (
          <p className={cn("text-text-primary text-3xl font-semibold tabular-nums", valueClass)}>
            {formatNumber(value)}
          </p>
        )}
      </CardContent>
    </Card>
  );
}

// ─── Volume Chart ────────────────────────────────────────────────────────────

function VolumeChart({ volume, loading }: { volume: VolumePoint[]; loading: boolean }) {
  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-text-muted text-xs font-medium tracking-wider uppercase">
          Log Volume
        </CardTitle>
      </CardHeader>
      <CardContent>
        {loading ? (
          <div className="bg-surface-active h-28 animate-pulse rounded" />
        ) : volume.length === 0 ? (
          <div className="flex h-28 items-center justify-center">
            <p className="text-text-muted text-sm">No data for this time range.</p>
          </div>
        ) : (
          <>
            <div className="flex h-28 items-end gap-px">
              {volume.map((point, i) => {
                const max = Math.max(...volume.map((p) => p.count));
                const heightPct = max > 0 ? Math.max((point.count / max) * 100, 2) : 2;
                return (
                  <div
                    key={i}
                    title={`${formatVolumeLabel(point.time)}: ${point.count.toLocaleString()} logs`}
                    className="bg-accent-brand/50 hover:bg-accent-brand min-w-px flex-1 cursor-default rounded-sm transition-colors"
                    style={{ height: `${heightPct}%` }}
                  />
                );
              })}
            </div>
            <div className="mt-1.5 flex justify-between">
              <span className="text-text-muted text-[10px]">
                {formatVolumeLabel(volume[0].time)}
              </span>
              <span className="text-text-muted text-[10px]">
                {formatVolumeLabel(volume[volume.length - 1].time)}
              </span>
            </div>
          </>
        )}
      </CardContent>
    </Card>
  );
}

// ─── Level Breakdown ─────────────────────────────────────────────────────────

function LevelBreakdownCard({
  breakdown,
  total,
  loading,
}: {
  breakdown: LevelBreakdown[];
  total: number;
  loading: boolean;
}) {
  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-text-muted text-xs font-medium tracking-wider uppercase">
          Level Breakdown
        </CardTitle>
      </CardHeader>
      <CardContent>
        {loading ? (
          <div className="space-y-3">
            {[80, 50, 30, 20].map((w, i) => (
              <div key={i} className="flex items-center gap-3">
                <div className="bg-surface-active h-3 w-10 animate-pulse rounded" />
                <div
                  className="bg-surface-active h-2 animate-pulse rounded"
                  style={{ width: `${w}%` }}
                />
              </div>
            ))}
          </div>
        ) : breakdown.length === 0 ? (
          <p className="text-text-muted py-4 text-sm">No data.</p>
        ) : (
          <div className="space-y-2.5">
            {breakdown.map((item) => {
              const pct = total > 0 ? Math.round((item.count / total) * 100) : 0;
              const barClass = LEVEL_BAR_CLASS[item.level] ?? "bg-zinc-400/70";
              return (
                <div key={item.level} className="flex items-center gap-3">
                  <span className="text-text-muted w-10 font-mono text-xs">{item.level}</span>
                  <div className="bg-surface-active relative h-1.5 flex-1 overflow-hidden rounded-full">
                    <div
                      className={cn("h-full rounded-full", barClass)}
                      style={{ width: `${pct}%` }}
                    />
                  </div>
                  <span className="text-text-muted w-16 text-right font-mono text-xs">
                    {formatNumber(item.count)}
                  </span>
                </div>
              );
            })}
          </div>
        )}
      </CardContent>
    </Card>
  );
}

// ─── Recent Errors ────────────────────────────────────────────────────────────

function RecentErrorsCard({ errors, loading }: { errors: RecentError[]; loading: boolean }) {
  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-text-muted text-xs font-medium tracking-wider uppercase">
          Recent Errors
        </CardTitle>
      </CardHeader>
      <CardContent>
        {loading ? (
          <div className="space-y-3">
            {[1, 2, 3].map((i) => (
              <div key={i} className="space-y-1">
                <div className="flex items-center gap-2">
                  <div className="bg-surface-active h-4 w-10 animate-pulse rounded" />
                  <div className="bg-surface-active h-3 w-20 animate-pulse rounded" />
                </div>
                <div className="bg-surface-active h-3 w-full animate-pulse rounded" />
              </div>
            ))}
          </div>
        ) : errors.length === 0 ? (
          <div className="flex h-20 items-center justify-center">
            <p className="text-text-muted text-sm">No recent errors. All good!</p>
          </div>
        ) : (
          <div className="space-y-3">
            {errors.map((err, i) => (
              <div key={i} className="border-border-divider border-b pb-3 last:border-0 last:pb-0">
                <div className="mb-0.5 flex items-center gap-2">
                  <Badge variant={levelBadgeVariant(err.level)} className="text-[10px]">
                    {err.level}
                  </Badge>
                  <span className="text-text-subtle font-mono text-xs">{err.service}</span>
                  <span className="text-text-disabled ml-auto text-[10px]">
                    {relativeTime(err.timestamp)}
                  </span>
                </div>
                <p className="text-text-muted truncate text-xs">{err.message}</p>
              </div>
            ))}
          </div>
        )}
      </CardContent>
    </Card>
  );
}

// ─── Loading Skeleton ─────────────────────────────────────────────────────────

function PageSkeleton() {
  return (
    <div className="space-y-6 p-6">
      <div className="flex items-center justify-between">
        <div>
          <div className="bg-surface-active h-6 w-28 animate-pulse rounded" />
          <div className="bg-surface-active mt-1 h-3 w-40 animate-pulse rounded" />
        </div>
        <div className="flex gap-2">
          <div className="bg-surface-active h-8 w-36 animate-pulse rounded" />
          <div className="bg-surface-active h-8 w-24 animate-pulse rounded" />
        </div>
      </div>
      <div className="grid grid-cols-3 gap-4">
        {[1, 2, 3].map((i) => (
          <div key={i} className="bg-surface-active h-24 animate-pulse rounded" />
        ))}
      </div>
      <div className="bg-surface-active h-44 animate-pulse rounded" />
      <div className="grid grid-cols-2 gap-4">
        <div className="bg-surface-active h-48 animate-pulse rounded" />
        <div className="bg-surface-active h-48 animate-pulse rounded" />
      </div>
    </div>
  );
}

// ─── Root Component ───────────────────────────────────────────────────────────

interface Props {
  slug: string;
}

export function DashboardContent({ slug }: Props) {
  const { project, loading: projectLoading, error: projectError } = useProject(slug);
  const [timeRange, setTimeRange] = useState<TimeRange>("24h");
  const {
    stats,
    loading: statsLoading,
    error: statsError,
    refresh,
  } = useDashboard(project?.id ?? null, timeRange);

  if (projectLoading) return <PageSkeleton />;

  if (projectError || !project) {
    return (
      <div className="p-8">
        <p className="text-status-danger text-sm">{projectError ?? "Project not found."}</p>
      </div>
    );
  }

  return (
    <div className="space-y-6 p-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-text-primary text-xl font-semibold">Dashboard</h1>
          <p className="text-text-muted mt-0.5 font-mono text-xs">{project.name}</p>
        </div>
        <div className="flex items-center gap-2">
          <Select
            value={timeRange}
            onChange={(e) => setTimeRange(e.target.value as TimeRange)}
            className="h-8 w-36 text-xs"
          >
            <option value="1h">Last 1 hour</option>
            <option value="6h">Last 6 hours</option>
            <option value="24h">Last 24 hours</option>
            <option value="7d">Last 7 days</option>
          </Select>
          <Button variant="outline" size="sm" onClick={refresh} className="gap-1.5">
            <ArrowClockwise size={13} />
            Refresh
          </Button>
        </div>
      </div>

      {statsError && <p className="text-status-danger text-sm">{statsError}</p>}

      {/* Stat cards */}
      <div className="grid grid-cols-3 gap-4">
        <StatCard label="Total Logs" value={stats?.total_logs ?? 0} loading={statsLoading} />
        <StatCard
          label="Errors"
          value={stats?.error_count ?? 0}
          loading={statsLoading}
          valueClass={stats && stats.error_count > 0 ? "text-status-danger" : undefined}
        />
        <StatCard
          label="Warnings"
          value={stats?.warn_count ?? 0}
          loading={statsLoading}
          valueClass={stats && stats.warn_count > 0 ? "text-status-warning" : undefined}
        />
      </div>

      {/* Volume chart */}
      <VolumeChart volume={stats?.volume ?? []} loading={statsLoading} />

      {/* Level breakdown + Recent errors */}
      <div className="grid grid-cols-2 gap-4">
        <LevelBreakdownCard
          breakdown={stats?.level_breakdown ?? []}
          total={stats?.total_logs ?? 0}
          loading={statsLoading}
        />
        <RecentErrorsCard errors={stats?.recent_errors ?? []} loading={statsLoading} />
      </div>
    </div>
  );
}
