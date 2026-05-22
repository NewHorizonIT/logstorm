import { NavLink } from "react-router";
import { useState } from "react";

import { cn } from "@/utils/cn";

const navItemBaseClass =
  "sidebar-item-enter group relative flex w-full items-center rounded-xl border px-3 py-2.5 text-left text-sm font-medium transition-all duration-200";

const navItemActiveClass =
  "border-cyan-300/35 bg-cyan-300/12 text-white shadow-[0_0_0_1px_rgba(103,232,249,0.2),0_14px_24px_-18px_rgba(34,211,238,0.7)]";

const navItemIdleClass =
  "border-transparent text-slate-300 hover:-translate-y-0.5 hover:border-white/12 hover:bg-white/4 hover:text-white";

const menuItems = [
  { name: "Dashboard", path: "/", icon: "grid" },
  { name: "Logs", path: "/logs", icon: "logs" },
  { name: "Metrics", path: "/metrics", icon: "metrics" },
  { name: "Alerts", path: "/alerts", icon: "alerts" },
  { name: "API Keys", path: "/api-keys", icon: "key" },
  { name: "Settings", path: "/settings", icon: "settings" },
] as const;

const Icon = ({ name }: { name: string }) => {
  switch (name) {
    case "grid":
      return (
        <svg className="h-5 w-5" viewBox="0 0 24 24" fill="none" stroke="currentColor">
          <rect x="3" y="3" width="8" height="8" rx="1" />
          <rect x="13" y="3" width="8" height="8" rx="1" />
          <rect x="3" y="13" width="8" height="8" rx="1" />
          <rect x="13" y="13" width="8" height="8" rx="1" />
        </svg>
      );
    case "logs":
      return (
        <svg className="h-5 w-5" viewBox="0 0 24 24" fill="none" stroke="currentColor">
          <path d="M8 6h11M8 12h11M8 18h11M3 6h.01M3 12h.01M3 18h.01" />
        </svg>
      );
    case "metrics":
      return (
        <svg className="h-5 w-5" viewBox="0 0 24 24" fill="none" stroke="currentColor">
          <path d="M3 3v18h18" />
          <path d="M7 13l3-3 4 4 5-7" />
        </svg>
      );
    case "alerts":
      return (
        <svg className="h-5 w-5" viewBox="0 0 24 24" fill="none" stroke="currentColor">
          <path d="M12 22c1.1 0 2-.9 2-2H10c0 1.1.9 2 2 2z" />
          <path d="M18 16V11c0-3.3-2.2-6.1-5.2-6.8V4a1 1 0 10-2 0v.2C8.2 4.9 6 7.7 6 11v5l-2 2v1h16v-1l-2-2z" />
        </svg>
      );
    case "key":
      return (
        <svg className="h-5 w-5" viewBox="0 0 24 24" fill="none" stroke="currentColor">
          <path d="M21 2l-6 6" />
          <circle cx="7" cy="17" r="4" />
          <path d="M21 2l-4 4" />
        </svg>
      );
    case "settings":
      return (
        <svg className="h-5 w-5" viewBox="0 0 24 24" fill="none" stroke="currentColor">
          <path d="M12 15.5A3.5 3.5 0 1112 8.5a3.5 3.5 0 010 7z" />
          <path d="M19.4 15a1.65 1.65 0 00.33 1.82l.06.06a2 2 0 11-2.83 2.83l-.06-.06a1.65 1.65 0 00-1.82-.33 1.65 1.65 0 00-1 1.51V21a2 2 0 11-4 0v-.09a1.65 1.65 0 00-1-1.51 1.65 1.65 0 00-1.82.33l-.06.06A2 2 0 114.28 18.9l.06-.06a1.65 1.65 0 00.33-1.82 1.65 1.65 0 00-1.51-1H3a2 2 0 110-4h.09a1.65 1.65 0 001.51-1 1.65 1.65 0 00-.33-1.82L4.21 6.7A2 2 0 116.7 4.21l.06.06a1.65 1.65 0 001.82.33H9a1.65 1.65 0 001-1.51V3a2 2 0 114 0v.09c0 .6.39 1.13 1 1.51h.41a1.65 1.65 0 001.82-.33l.06-.06A2 2 0 1119.79 6.7l-.06.06a1.65 1.65 0 00-.33 1.82V9c.6 0 1.13.39 1.51 1h.09a2 2 0 110 4h-.09c-.38.61-.91 1-1.51 1v.41c0 .6-.39 1.13-1 1.51z" />
        </svg>
      );
    default:
      return null;
  }
};

export const Sidebar = ({ collapsed, setCollapsed }: { collapsed: boolean; setCollapsed: (collapsed: boolean) => void }) => {
  


  return (
    <aside className={cn("sidebar-enter fixed inset-y-0 left-0 z-40 border-r border-white/10 bg-slate-950/72 p-4 backdrop-blur-xl", collapsed ? "w-14" : "w-60")}>
      <div className="mb-4 flex items-start justify-between">
        <div className={cn("rounded-2xl border p-3", collapsed ? "hidden" : "bg-white/3 border-white/10") }>
          <p className="inline-flex items-center rounded-full border border-cyan-300/25 bg-cyan-300/10 px-2.5 py-1 text-[10px] font-semibold uppercase tracking-[0.11em] text-cyan-200">
            Real-time observability
          </p>
          <p className="mt-3 font-display text-lg font-semibold text-white">LogStorm</p>
          <p className="mt-1 text-xs text-slate-400">Observability Control Center</p>
        </div>

        <button
          type="button"
          aria-label={collapsed ? "Expand sidebar" : "Collapse sidebar"}
          className="rounded-md border border-transparent bg-transparent px-2 py-1 text-sm text-slate-300"
          onClick={() => setCollapsed((v) => !v)}
        >
          {collapsed ? "»" : "«"}
        </button>
      </div>

      <nav className="space-y-1.5">
        {menuItems.map((item, index) => (
          <NavLink
            key={item.path}
            to={item.path}
            title={item.name}
            style={{ animationDelay: `${index * 70}ms` }}
            className={({ isActive }) =>
              cn(
                navItemBaseClass,
                isActive ? navItemActiveClass : navItemIdleClass,
                collapsed ? "justify-center px-2" : "px-3"
              )
            }
          >
            <span className="mr-3 flex h-5 w-5 items-center justify-center text-[--color-text-secondary]">
              <Icon name={item.icon} />
            </span>
            <span className={cn(collapsed ? "hidden" : "inline-flex items-center")}>{item.name}</span>
          </NavLink>
        ))}
      </nav>
    </aside>
  );
};
