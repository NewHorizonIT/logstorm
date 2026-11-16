import { cva, type VariantProps } from "class-variance-authority";

import { cn } from "@/lib/utils";

const badgeVariants = cva(
  "inline-flex items-center gap-1 px-2 py-0.5 text-[11px] font-medium transition-colors",
  {
    variants: {
      variant: {
        info: "rounded-badge bg-status-info-bg text-status-info",
        success: "rounded-badge bg-status-success-bg text-status-success",
        warning: "rounded-badge bg-status-warning-bg text-status-warning",
        danger: "rounded-badge bg-status-danger-bg text-status-danger",
        neutral: "rounded-badge bg-status-neutral-bg text-status-neutral",
        production: "rounded-pill bg-status-success-bg text-status-success",
        staging: "rounded-pill bg-status-warning-bg text-status-warning",
        development: "rounded-pill bg-status-info-bg text-status-info",
      },
    },
    defaultVariants: {
      variant: "neutral",
    },
  },
);

interface BadgeProps extends React.ComponentProps<"span">, VariantProps<typeof badgeVariants> {}

function Badge({ className, variant, ...props }: BadgeProps) {
  return (
    <span data-slot="badge" className={cn(badgeVariants({ variant }), className)} {...props} />
  );
}

export { Badge, badgeVariants };
