import { Lightning } from "@phosphor-icons/react/ssr";

export default function AuthLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="bg-bg-base flex min-h-screen flex-col items-center justify-center px-4">
      <div className="mb-8 flex items-center gap-2">
        <Lightning size={22} weight="fill" className="text-accent-brand" />
        <span className="text-text-primary text-base font-semibold">LogStorm</span>
      </div>
      {children}
    </div>
  );
}
