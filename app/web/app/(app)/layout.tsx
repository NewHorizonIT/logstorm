export default function AppLayout({ children }: { children: React.ReactNode }) {
  return (
    <div suppressHydrationWarning className="bg-bg-base flex h-screen">
      {children}
    </div>
  );
}
