import { Card } from "@/components/Card";
import { Input } from "@/components/Input";

const savedQueries = [
  "avg(rate(http_requests_total[5m])) by (service)",
  "sum(rate(log_errors_total[1m])) by (service)",
  "histogram_quantile(0.95, sum(rate(request_duration_bucket[5m])) by (le, service))",
];

const MetricsPage = () => {
  return (
    <div className="space-y-4">
      <Card>
        <h2 className="mb-3 text-sm font-semibold text-[--color-text-primary]">Query</h2>
        <Input placeholder="Enter metrics query..." />
      </Card>

      <Card>
        <div className="mb-4 flex items-center justify-between gap-3">
          <h2 className="text-sm font-semibold text-[--color-text-primary]">Chart</h2>
          <span className="text-xs text-[--color-text-secondary]">Placeholder</span>
        </div>
        <div className="h-80 rounded-[--radius-md] border border-dashed border-[--color-border] bg-[color-mix(in_srgb,var(--color-card),white_3%)]" />
      </Card>

      <Card>
        <h2 className="mb-3 text-sm font-semibold text-[--color-text-primary]">Saved Queries</h2>
        <ul className="space-y-2">
          {savedQueries.map((query) => (
            <li
              key={query}
              className="rounded-[--radius-md] border border-[--color-border] bg-[color-mix(in_srgb,var(--color-card),white_3%)] px-3 py-2 text-sm text-[--color-text-secondary]"
            >
              {query}
            </li>
          ))}
        </ul>
      </Card>
    </div>
  );
};

export default MetricsPage;
