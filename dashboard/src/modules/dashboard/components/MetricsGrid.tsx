import { memo } from "react";
import { Card } from "@/components/Card";

export type Metric = {
  label: string;
  value: string;
  delta: string;
};

type Props = {
  metrics: Metric[];
};

const MetricsGridBase = ({ metrics }: Props) => {
  return (
    <section className="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
      {metrics.map((metric) => (
        <Card key={metric.label} variant="elevated">
          <p className="text-sm text-[--color-text-secondary]">{metric.label}</p>
          <p className="mt-2 text-2xl font-semibold text-[--color-text-primary]">{metric.value}</p>
          <p className="mt-2 text-xs text-emerald-300">{metric.delta} vs last hour</p>
        </Card>
      ))}
    </section>
  );
};

export const MetricsGrid = memo(MetricsGridBase);
