import { memo } from "react";
import { Card } from "@/components/Card";
import { Table } from "@/components/Table";

export type TopService = {
  name: string;
  logsPerMin: string;
  errorRate: string;
};

export type RecentError = {
  time: string;
  service: string;
  message: string;
};

type Props = {
  topServices: TopService[];
  recentErrors: RecentError[];
};

const DashboardTablesBase = ({ topServices, recentErrors }: Props) => {
  return (
    <section className="grid gap-6 xl:grid-cols-2">
      <Card>
        <h2 className="mb-4 text-lg font-semibold text-[--color-text-primary]">Top Services</h2>
        <Table
          columns={[
            { key: "name", header: "Service", className: "px-4 py-3 text-[--color-text-primary]" },
            { key: "logsPerMin", header: "Logs/min" },
            { key: "errorRate", header: "Error rate" },
          ]}
          rows={topServices}
          getRowKey={(row) => row.name}
        />
      </Card>

      <Card>
        <h2 className="mb-4 text-lg font-semibold text-[--color-text-primary]">Recent Errors</h2>
        <Table
          columns={[
            { key: "time", header: "Time" },
            { key: "service", header: "Service", className: "px-4 py-3 text-[--color-text-primary]" },
            { key: "message", header: "Message" },
          ]}
          rows={recentErrors}
          getRowKey={(row) => `${row.time}-${row.service}`}
        />
      </Card>
    </section>
  );
};

export const DashboardTables = memo(DashboardTablesBase);
