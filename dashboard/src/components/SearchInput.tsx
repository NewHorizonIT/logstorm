import { memo } from "react";
import { Input } from "@/components/Input";

type Props = {
  value: string;
  onChange: (value: string) => void;
  placeholder?: string;
};

const SearchInputBase = ({ value, onChange, placeholder = "Search logs, services, trace_id..." }: Props) => {
  return (
    <div className="relative">
      <span className="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-sm text-slate-500">⌕</span>
      <Input
        value={value}
        onChange={(event) => onChange(event.target.value)}
        placeholder={placeholder}
        className="pl-9"
      />
    </div>
  );
};

export const SearchInput = memo(SearchInputBase);
