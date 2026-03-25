import type { HTMLAttributes } from "react";
import { cn } from "@/utils/cn";

type CardVariant = "default" | "elevated";

type Props = HTMLAttributes<HTMLDivElement> & {
  variant?: CardVariant;
};

const cardBaseClass =
  "rounded-2xl border p-5 text-white backdrop-blur-sm transition-shadow duration-200";

const cardVariantClass: Record<CardVariant, string> = {
  default: "border-white/10 bg-slate-900/50 shadow-[inset_0_1px_0_rgba(255,255,255,0.03)]",
  elevated:
    "border-white/12 bg-slate-900/65 shadow-[inset_0_1px_0_rgba(255,255,255,0.05),0_24px_40px_-26px_rgba(2,6,23,0.75)]",
};

export const Card = ({ className, children, variant = "default", ...props }: Props) => {
  return (
    <section
      className={cn(
        cardBaseClass,
        cardVariantClass[variant],
        className,
      )}
      {...props}
    >
      {children}
    </section>
  );
};
