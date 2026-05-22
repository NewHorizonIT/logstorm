import { Sidebar } from "@/components/Sidebar";
import { useState } from "react";
import { Outlet } from "react-router";

const MainLayout = () => {
  const [collapsed, setCollapsed] = useState(false);
  return (
    <div className="relative flex min-h-screen bg-[radial-gradient(circle_at_15%_20%,rgba(59,130,246,0.12),transparent_45%),radial-gradient(circle_at_85%_0%,rgba(209,105,0,0.12),transparent_30%),var(--color-bg)] text-[--color-text-primary]">
      <Sidebar collapsed={collapsed} setCollapsed={setCollapsed} />

      <div className="ml-60 flex min-h-screen flex-1 flex-col">
        <header className="sticky top-0 z-30 flex h-16 items-center border-b border-[--color-border] bg-[--color-card]/85 px-6 backdrop-blur-md">
          <div>
            <h1 className="font-display text-base font-semibold tracking-wide">LogStorm Dashboard</h1>
            <p className="text-xs text-[--color-text-secondary]">Real-time logs, metrics and system visibility</p>
          </div>
        </header>

        <main className="flex-1 overflow-y-auto p-6">
          <div className="min-h-[calc(100vh-7rem)] rounded-[--radius-lg] border border-[--color-border] bg-[--color-card]/45 p-5 shadow-[inset_0_1px_0_rgba(255,255,255,0.04)] backdrop-blur-sm md:p-6">
            <Outlet />
          </div>
        </main>
      </div>
    </div>
  );
};

export default MainLayout;