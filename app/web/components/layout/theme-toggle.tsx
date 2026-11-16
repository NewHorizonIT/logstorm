"use client";

import { useTheme } from "next-themes";
import { Sun, Moon, Monitor } from "@phosphor-icons/react";
import { useEffect, useState } from "react";

import { cn } from "@/lib/utils";

const cycle = ["light", "dark", "system"] as const;
type Theme = (typeof cycle)[number];

const icons: Record<Theme, React.ComponentType<{ size?: number }>> = {
  light: Sun,
  dark: Moon,
  system: Monitor,
};

export function ThemeToggle() {
  const { theme, setTheme } = useTheme();

  const current = (cycle.includes(theme as Theme) ? theme : "system") as Theme;
  const Icon = icons[current];

  function handleClick() {
    const next = cycle[(cycle.indexOf(current) + 1) % cycle.length];
    setTheme(next);
  }

  return (
    <button
      onClick={handleClick}
      suppressHydrationWarning
      className={cn(
        "rounded-control flex size-7 items-center justify-center transition-colors",
        "text-text-muted hover:bg-surface-active hover:text-text-primary",
      )}
      title={`Theme: ${current}`}
    >
      <Icon size={15} />
    </button>
  );
}
