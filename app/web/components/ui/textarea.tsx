import * as React from "react";

import { cn } from "@/lib/utils";

function Textarea({ className, ...props }: React.ComponentProps<"textarea">) {
  return (
    <textarea
      data-slot="textarea"
      className={cn(
        "border-border-input placeholder:text-text-disabled focus-visible:border-accent-focus disabled:bg-bg-inset rounded-input text-text-primary focus-visible:ring-accent-focus/30 min-h-[80px] w-full border bg-transparent px-3 py-2 text-xs transition-colors outline-none focus-visible:ring-1 disabled:cursor-not-allowed disabled:opacity-50",
        className,
      )}
      {...props}
    />
  );
}

export { Textarea };
