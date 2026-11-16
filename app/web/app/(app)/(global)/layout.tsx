import { GlobalSidebar } from "@/components/layout/global-sidebar";

export default function GlobalLayout({ children }: { children: React.ReactNode }) {
  return (
    <>
      <GlobalSidebar />
      <main className="flex flex-1 flex-col overflow-auto">{children}</main>
    </>
  );
}
