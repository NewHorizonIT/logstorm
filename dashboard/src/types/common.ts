export type LogLevel = "error" | "warn" | "info";

export type TableColumn<T extends Record<string, unknown>> = {
  key: keyof T | string;
  header: string;
  className?: string;
  cell?: (row: T) => React.ReactNode;
};
