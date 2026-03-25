import type { ReactNode } from "react";
import { cn } from "@/utils/cn";

type BadgeType = "error" | "warn" | "info";

type Props = {
  type: BadgeType;
  children: ReactNode;
  className?: string;
};

const typeClass: Record<BadgeType, string> = {
  error: "border-rose-300/30 bg-rose-500/15 text-rose-200",
  warn: "border-amber-300/30 bg-amber-500/15 text-amber-200",
  info: "border-cyan-300/30 bg-cyan-500/15 text-cyan-200",
};

export const Badge = ({ type, children, className }: Props) => {
  return (
    <span
      className={cn(
        "inline-flex items-center rounded-full border px-2 py-0.5 text-xs font-semibold uppercase tracking-wide",
        typeClass[type],
        className,
      )}
    >
      {children}
    </span>
  );
};
