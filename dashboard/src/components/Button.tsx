import type { ButtonHTMLAttributes, ReactNode } from "react";
import { cn } from "@/utils/cn";

type Variant = "primary" | "secondary" | "outline" | "inverted";

type Props = {
  variant?: Variant;
  leftIcon?: ReactNode;
  rightIcon?: ReactNode;
} & ButtonHTMLAttributes<HTMLButtonElement>;

const buttonBaseClass =
  "inline-flex h-11 items-center justify-center gap-2 rounded-xl px-4 text-sm font-semibold text-white transition duration-200 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-cyan-400/40 focus-visible:ring-offset-2 focus-visible:ring-offset-slate-950 disabled:cursor-not-allowed disabled:opacity-45";

const variantClass: Record<Variant, string> = {
  primary:
    "border border-transparent bg-linear-to-r from-cyan-500 to-blue-500 shadow-[0_10px_25px_-12px_rgba(6,182,212,0.9)] enabled:hover:brightness-110 enabled:active:translate-y-px",
  secondary:
    "border border-white/10 bg-slate-800/85 shadow-[inset_0_1px_0_rgba(255,255,255,0.03)] enabled:hover:border-white/20 enabled:hover:bg-slate-700/85 enabled:active:translate-y-px",
  outline:
    "border border-white/16 bg-white/3 text-slate-100 enabled:hover:border-white/25 enabled:hover:bg-white/6 enabled:active:translate-y-px",
  inverted:
    "border border-white bg-white text-slate-950 enabled:hover:bg-white/90 enabled:active:translate-y-px",
};

export const Button = ({
  variant = "primary",
  className,
  leftIcon,
  rightIcon,
  children,
  ...props
}: Props) => {
  return (
    <button
      className={cn(
        buttonBaseClass,
        variantClass[variant],
        className,
      )}
      {...props}
    >
      {leftIcon}
      {children}
      {rightIcon}
    </button>
  );
};
