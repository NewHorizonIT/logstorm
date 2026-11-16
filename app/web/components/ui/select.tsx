"use client";

import * as React from "react";
import { CaretDown } from "@phosphor-icons/react";

import { cn } from "@/lib/utils";

function Select({ className, children, ...props }: React.ComponentProps<"select">) {
  return (
    <div className="relative w-full">
      <select
        data-slot="select"
        className={cn(
          "border-border-input focus:border-accent-focus disabled:bg-bg-inset rounded-input bg-surface text-text-primary focus:ring-accent-focus/30 w-full appearance-none border px-3 py-2 pr-8 text-xs transition-colors outline-none focus:ring-1 disabled:cursor-not-allowed disabled:opacity-50",
          className,
        )}
        {...props}
      >
        {children}
      </select>
      <CaretDown
        size={12}
        className="text-text-muted pointer-events-none absolute top-1/2 right-3 -translate-y-1/2"
      />
    </div>
  );
}

export { Select };
