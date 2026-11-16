"use client";

import { useState } from "react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import {
  Lightning,
  List,
  Key,
  Gear,
  SquaresFour,
  CaretUpDown,
  SidebarSimple,
} from "@phosphor-icons/react";

import { cn } from "@/lib/utils";
import { ThemeToggle } from "@/components/layout/theme-toggle";

type NavItem = {
  icon: React.ComponentType<{ size?: number; className?: string }>;
  label: string;
  href: string;
};

function projectNavItems(slug: string): NavItem[] {
  return [
    { icon: List, label: "Logs", href: `/projects/${slug}/logs` },
    { icon: Key, label: "API Keys", href: `/projects/${slug}/api-keys` },
    { icon: Gear, label: "Settings", href: `/projects/${slug}/settings` },
  ];
}

function NavLink({ item, collapsed }: { item: NavItem; collapsed: boolean }) {
  const pathname = usePathname();
  const isActive = pathname === item.href || pathname.startsWith(item.href + "/");

  return (
    <Link
      href={item.href}
      title={collapsed ? item.label : undefined}
      className={cn(
        "rounded-control flex items-center gap-3 px-3 py-2.5 text-sm font-medium transition-colors",
        collapsed && "justify-center px-2",
        isActive
          ? "bg-accent-brand/15 text-accent-brand"
          : "text-text-muted hover:bg-surface-active hover:text-text-primary",
      )}
    >
      <item.icon size={17} />
      {!collapsed && <span>{item.label}</span>}
    </Link>
  );
}

interface ProjectSidebarProps {
  projectSlug: string;
}

export function ProjectSidebar({ projectSlug }: ProjectSidebarProps) {
  const [collapsed, setCollapsed] = useState(false);
  const navItems = projectNavItems(projectSlug);

  return (
    <aside
      className={cn(
        "border-border-divider bg-bg-sidebar flex h-screen shrink-0 flex-col border-r transition-all duration-200",
        collapsed ? "w-16" : "w-64",
      )}
    >
      {/* Logo + toggle */}
      <div
        className={cn(
          "flex items-center px-4 py-5",
          collapsed ? "justify-center" : "justify-between",
        )}
      >
        {!collapsed && (
          <div className="flex items-center gap-2">
            <Lightning size={18} weight="fill" className="text-accent-brand" />
            <span className="text-text-primary text-sm font-semibold">LogStorm</span>
          </div>
        )}
        <button
          onClick={() => setCollapsed((v) => !v)}
          className="rounded-control text-text-muted hover:bg-surface-active hover:text-text-primary flex size-7 items-center justify-center transition-colors"
          title={collapsed ? "Expand sidebar" : "Collapse sidebar"}
        >
          <SidebarSimple size={16} />
        </button>
      </div>

      {/* Project switcher */}
      {!collapsed && (
        <div className="px-4 pb-3">
          <button className="rounded-control border-border-input bg-surface hover:bg-surface-active flex w-full items-center justify-between border px-3 py-2 text-xs transition-colors">
            <div className="flex flex-col items-start gap-0.5">
              <span className="text-text-subtle text-[10px]">Project</span>
              <span className="text-text-primary font-mono text-xs">{projectSlug}</span>
            </div>
            <CaretUpDown size={12} className="text-text-muted" />
          </button>
        </div>
      )}

      {/* Project nav */}
      <nav className="flex flex-1 flex-col gap-1 px-3 py-2">
        {navItems.map((item) => (
          <NavLink key={item.href} item={item} collapsed={collapsed} />
        ))}

        {/* Divider + All Projects */}
        <div className="border-border-divider my-2 border-t" />
        <Link
          href="/projects"
          title={collapsed ? "All Projects" : undefined}
          className={cn(
            "rounded-control text-text-muted hover:bg-surface-active hover:text-text-primary flex items-center gap-3 px-3 py-2.5 text-sm font-medium transition-colors",
            collapsed && "justify-center px-2",
          )}
        >
          <SquaresFour size={17} />
          {!collapsed && <span>All Projects</span>}
        </Link>
      </nav>

      {/* User footer */}
      <div className="border-border-divider border-t px-3 py-4">
        <div className={cn("flex items-center gap-2.5", collapsed && "justify-center")}>
          <div className="rounded-pill bg-avatar text-text-primary flex size-7 shrink-0 items-center justify-center text-[11px] font-semibold">
            YN
          </div>
          {!collapsed && (
            <div className="min-w-0 flex-1">
              <p className="text-text-primary truncate text-xs font-medium">[YOUR NAME]</p>
              <p className="text-text-subtle truncate text-[11px]">you@company.com</p>
            </div>
          )}
          {!collapsed && <ThemeToggle />}
        </div>
      </div>
    </aside>
  );
}
