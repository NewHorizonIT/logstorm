import { Button } from "@/components/Button";
import { Card } from "@/components/Card";
import { Badge } from "@/components/Badge";

type AlertRule = {
  id: string;
  name: string;
  condition: string;
  status: "error" | "warn" | "info";
};

const mockAlerts: AlertRule[] = [
  {
    id: "alert-1",
    name: "High Error Rate",
    condition: "error_rate > 3% for 5m",
    status: "error",
  },
  {
    id: "alert-2",
    name: "Kafka Lag Spike",
    condition: "kafka_lag > 500ms for 10m",
    status: "warn",
  },
  {
    id: "alert-3",
    name: "Ingestion Recovery",
    condition: "logs_ingested_per_sec back to normal",
    status: "info",
  },
];

const AlertsPage = () => {
  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between gap-3">
        <h1 className="text-lg font-semibold text-[--color-text-primary]">Alert Rules</h1>
        <Button>Create Alert</Button>
      </div>

      <Card>
        <ul className="space-y-2">
          {mockAlerts.map((alert) => (
            <li
              key={alert.id}
              className="flex items-center justify-between gap-3 rounded-[--radius-md] border border-[--color-border] bg-[color-mix(in_srgb,var(--color-card),white_3%)] px-3 py-2"
            >
              <div>
                <p className="text-sm font-medium text-[--color-text-primary]">{alert.name}</p>
                <p className="text-xs text-[--color-text-secondary]">{alert.condition}</p>
              </div>
              <Badge type={alert.status}>{alert.status}</Badge>
            </li>
          ))}
        </ul>
      </Card>
    </div>
  );
};

export default AlertsPage;
