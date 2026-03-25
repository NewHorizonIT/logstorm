
import { Card } from "@/components/Card";

type Metric = {
  label: string;
  value: string;
  delta: string;
};

type TopService = {
  name: string;
  logsPerMin: string;
  errorRate: string;
};

type RecentError = {
  time: string;
  service: string;
  message: string;
};

const metrics: Metric[] = [
  { label: "Logs/sec", value: "12,480", delta: "+7.2%" },
  { label: "Error rate", value: "1.9%", delta: "-0.4%" },
  { label: "Active services", value: "36", delta: "+2" },
  { label: "Kafka lag", value: "143 ms", delta: "-22 ms" },
];

const topServices: TopService[] = [
  { name: "api-gateway", logsPerMin: "122,331", errorRate: "0.9%" },
  { name: "payment-service", logsPerMin: "95,772", errorRate: "2.4%" },
  { name: "auth-service", logsPerMin: "82,194", errorRate: "0.7%" },
  { name: "notification-service", logsPerMin: "51,203", errorRate: "1.2%" },
  { name: "inventory-service", logsPerMin: "39,402", errorRate: "1.6%" },
];

const recentErrors: RecentError[] = [
  {
    time: "2026-03-24 10:43:22",
    service: "payment-service",
    message: "timeout while calling card processor",
  },
  {
    time: "2026-03-24 10:42:05",
    service: "api-gateway",
    message: "upstream unavailable: inventory-service",
  },
  {
    time: "2026-03-24 10:40:37",
    service: "notification-service",
    message: "SMTP provider rate-limited request",
  },
  {
    time: "2026-03-24 10:39:12",
    service: "auth-service",
    message: "token verification failed: invalid signature",
  },
  {
    time: "2026-03-24 10:38:46",
    service: "payment-service",
    message: "retry exhausted for order #A92F31",
  },
];

const Dashboard = () => {
  return (
    <div className="space-y-6">
      <section className="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
        {metrics.map((metric) => (
          <Card key={metric.label} variant="elevated">
            <p className="text-sm text-[--color-text-secondary]">{metric.label}</p>
            <p className="mt-2 text-2xl font-semibold text-[--color-text-primary]">{metric.value}</p>
            <p className="mt-2 text-xs text-emerald-300">{metric.delta} vs last hour</p>
          </Card>
        ))}
      </section>

      <Card variant="elevated">
        <div className="mb-4 flex items-center justify-between gap-3">
          <h2 className="text-lg font-semibold text-[--color-text-primary]">Traffic Overview</h2>
          <span className="text-sm text-[--color-text-secondary]">Chart placeholder</span>
        </div>
        <div className="h-72 rounded-[--radius-md] border border-dashed border-[--color-border] bg-[color-mix(in_srgb,var(--color-card),white_4%)]" />
      </Card>

      <section className="grid gap-6 xl:grid-cols-2">
        <Card>
          <h2 className="mb-4 text-lg font-semibold text-[--color-text-primary]">Top Services</h2>
          <div className="overflow-x-auto">
            <table className="w-full min-w-120 text-left text-sm">
              <thead>
                <tr className="border-b border-[--color-border] text-[--color-text-secondary]">
                  <th className="py-2 font-medium">Service</th>
                  <th className="py-2 font-medium">Logs/min</th>
                  <th className="py-2 font-medium">Error rate</th>
                </tr>
              </thead>
              <tbody>
                {topServices.map((service) => (
                  <tr key={service.name} className="border-b border-[--color-border]/70 last:border-b-0">
                    <td className="py-3 text-[--color-text-primary]">{service.name}</td>
                    <td className="py-3 text-[--color-text-secondary]">{service.logsPerMin}</td>
                    <td className="py-3 text-[--color-text-secondary]">{service.errorRate}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </Card>

        <Card>
          <h2 className="mb-4 text-lg font-semibold text-[--color-text-primary]">Recent Errors</h2>
          <div className="overflow-x-auto">
            <table className="w-full min-w-120 text-left text-sm">
              <thead>
                <tr className="border-b border-[--color-border] text-[--color-text-secondary]">
                  <th className="py-2 font-medium">Time</th>
                  <th className="py-2 font-medium">Service</th>
                  <th className="py-2 font-medium">Message</th>
                </tr>
              </thead>
              <tbody>
                {recentErrors.map((error) => (
                  <tr key={`${error.time}-${error.service}`} className="border-b border-[--color-border]/70 last:border-b-0">
                    <td className="py-3 text-[--color-text-secondary]">{error.time}</td>
                    <td className="py-3 text-[--color-text-primary]">{error.service}</td>
                    <td className="py-3 text-[--color-text-secondary]">{error.message}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </Card>
      </section>
    </div>
  );
};

export default Dashboard
