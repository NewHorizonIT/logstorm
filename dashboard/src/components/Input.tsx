import { memo } from "react";
import type { InputHTMLAttributes, SelectHTMLAttributes, TextareaHTMLAttributes } from "react";
import { cn } from "@/utils/cn";

type Props = InputHTMLAttributes<HTMLInputElement>;

type SelectProps = SelectHTMLAttributes<HTMLSelectElement>;

type TextareaProps = TextareaHTMLAttributes<HTMLTextAreaElement>;

const controlBaseClass =
  "w-full rounded-xl border border-white/12 bg-slate-900/70 text-sm text-white shadow-[inset_0_1px_0_rgba(255,255,255,0.03)] outline-none transition duration-200 placeholder:text-slate-500 hover:border-white/20 focus:border-cyan-400/80 focus:ring-2 focus:ring-cyan-400/20 disabled:cursor-not-allowed disabled:opacity-50";

const inputBaseClass = `${controlBaseClass} h-11 px-3`;

const selectBaseClass =
  `${controlBaseClass} h-11 appearance-none bg-[linear-gradient(45deg,transparent_50%,#94a3b8_50%),linear-gradient(135deg,#94a3b8_50%,transparent_50%)] bg-[position:calc(100%-18px)_50%,calc(100%-12px)_50%] bg-[size:6px_6px,6px_6px] bg-no-repeat px-3 pr-10`;

const textareaBaseClass = `${controlBaseClass} min-h-28 resize-y px-3 py-2.5`;

export const Input = ({ className, ...props }: Props) => {
  return (
    <input
      className={cn(
        inputBaseClass,
        className,
      )}
      {...props}
    />
  );
};

export const Select = ({ className, children, ...props }: SelectProps) => {
  return (
    <select
      className={cn(
        selectBaseClass,
        className,
      )}
      {...props}
    >
      {children}
    </select>
  );
};

export const Textarea = ({ className, ...props }: TextareaProps) => {
  return (
    <textarea
      className={cn(
        textareaBaseClass,
        className,
      )}
      {...props}
    />
  );
};

type SearchInputProps = {
  value: string;
  onChange: (value: string) => void;
  placeholder?: string;
};

const SearchInputBase = ({
  value,
  onChange,
  placeholder = "Search logs, services, trace_id...",
}: SearchInputProps) => {
  return (
    <div className="group relative">
      <span className="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-slate-500 transition-colors group-focus-within:text-cyan-300">
        <svg
          aria-hidden="true"
          viewBox="0 0 20 20"
          fill="none"
          className="h-4 w-4"
        >
          <path
            d="M13.5 13.5L17 17M15.5 9C15.5 12.5899 12.5899 15.5 9 15.5C5.41015 15.5 2.5 12.5899 2.5 9C2.5 5.41015 5.41015 2.5 9 2.5C12.5899 2.5 15.5 5.41015 15.5 9Z"
            stroke="currentColor"
            strokeWidth="1.75"
            strokeLinecap="round"
            strokeLinejoin="round"
          />
        </svg>
      </span>
      <Input
        value={value}
        onChange={(event) => onChange(event.target.value)}
        placeholder={placeholder}
        className="pl-10"
      />
    </div>
  );
};

export const SearchInput = memo(SearchInputBase);
