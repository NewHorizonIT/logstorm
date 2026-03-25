import { NavLink } from "react-router";

import { cn } from "@/utils/cn";

const navItemBaseClass =
  "sidebar-item-enter group relative flex w-full items-center rounded-xl border px-3 py-2.5 text-left text-sm font-medium transition-all duration-200";

const navItemActiveClass =
  "border-cyan-300/35 bg-cyan-300/12 text-white shadow-[0_0_0_1px_rgba(103,232,249,0.2),0_14px_24px_-18px_rgba(34,211,238,0.7)]";

const navItemIdleClass =
  "border-transparent text-slate-300 hover:-translate-y-0.5 hover:border-white/12 hover:bg-white/4 hover:text-white";

const menuItems = [
  {
    name: "Dashboard",
    path: "/",
  },
  {
    name: "Logs",
    path: "/logs",
  },
  {
    name: "Metrics",
    path: "/metrics",
  },
  {
    name: "Alerts",
    path: "/alerts",
  },
  {
    name: "API Keys",
    path: "/api-keys",
  },
  {
    name: "Settings",
    path: "/settings",
  }
] as const;

export const Sidebar = () => {
  return (
    <aside className="sidebar-enter fixed inset-y-0 left-0 z-40 w-60 border-r border-white/10 bg-slate-950/72 p-4 backdrop-blur-xl">
      <div className="mb-6 rounded-2xl border border-white/10 bg-white/3 p-4">
        <p className="inline-flex items-center rounded-full border border-cyan-300/25 bg-cyan-300/10 px-2.5 py-1 text-[10px] font-semibold uppercase tracking-[0.11em] text-cyan-200">
          Real-time observability
        </p>
        <p className="mt-3 font-display text-lg font-semibold text-white">LogStorm</p>
        <p className="mt-1 text-xs text-slate-400">Observability Control Center</p>
      </div>

      <nav className="space-y-1.5">
        {menuItems.map((item, index) => (
          <NavLink
            key={item.path}
            to={item.path}
            style={{ animationDelay: `${index * 70}ms` }}
            className={({ isActive }) =>
              cn(
                navItemBaseClass,
                isActive ? navItemActiveClass : navItemIdleClass,
              )
            }
          >
            {item.name}
          </NavLink>
        ))}
      </nav>
    </aside>
  );
};
