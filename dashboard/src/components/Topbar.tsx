import { SearchInput } from "@/components/Input";
import { Button } from "@/components/Button";

type Props = {
  searchValue?: string;
  onSearchChange?: (value: string) => void;
  showTimeRange?: boolean;
};

export const Topbar = ({
  searchValue = "",
  onSearchChange = () => undefined,
  showTimeRange = true,
}: Props) => {
  return (
    <header className="flex h-16 items-center justify-between gap-4 border-b border-white/10 bg-slate-950/70 px-6 shadow-[inset_0_1px_0_rgba(255,255,255,0.04)] backdrop-blur-md">
      <div className="w-full max-w-xl">
        <SearchInput
          value={searchValue}
          onChange={onSearchChange}
          placeholder="Search logs, services, trace_id..."
        />
      </div>

      <div className="flex items-center gap-3">
        {showTimeRange && (
          <div className="hidden rounded-xl border border-white/12 bg-white/3 px-3 py-2 text-sm text-slate-400 shadow-[inset_0_1px_0_rgba(255,255,255,0.03)] sm:block">
            Last 15 minutes
          </div>
        )}

        <Button
          variant="outline"
          className="h-10 px-3 py-0 text-slate-100 border-white/16 bg-white/3 hover:border-white/25 hover:bg-white/6"
        >
          <span className="inline-block h-8 w-8 rounded-full bg-linear-to-br from-cyan-400/60 to-blue-700/60 border border-white/20 shadow-sm" />
          <span className="hidden md:inline">User</span>
        </Button>
      </div>
    </header>
  );
};
