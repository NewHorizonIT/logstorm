import { memo } from "react";

type Column<T> = {
  key: keyof T | string;
  header: string;
  className?: string;
  cell?: (row: T) => React.ReactNode;
};

type Props<T> = {
  columns: Column<T>[];
  rows: T[];
  getRowKey: (row: T) => string;
  emptyText?: string;
};

const TableBase = <T extends Record<string, unknown>>({
  columns,
  rows,
  getRowKey,
  emptyText = "No data",
}: Props<T>) => {
  return (
    <div className="overflow-x-auto rounded-xl border border-white/10 bg-slate-950/35 shadow-[inset_0_1px_0_rgba(255,255,255,0.03)]">
      <table className="w-full min-w-120 text-left text-sm">
        <thead className="bg-slate-900/80 text-slate-400">
          <tr>
            {columns.map((column) => (
              <th key={String(column.key)} className="px-4 py-3 font-semibold tracking-wide">
                {column.header}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {rows.length === 0 && (
            <tr className="border-t border-white/10">
              <td className="px-4 py-6 text-center text-sm text-slate-500" colSpan={columns.length}>
                {emptyText}
              </td>
            </tr>
          )}

          {rows.map((row) => (
            <tr
              key={getRowKey(row)}
              className="border-t border-white/10 transition-colors duration-150 hover:bg-white/3"
            >
              {columns.map((column) => (
                <td key={String(column.key)} className={column.className ?? "px-4 py-3 text-slate-300"}>
                  {column.cell ? column.cell(row) : String(row[column.key as keyof T] ?? "")}
                </td>
              ))}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
};

export const Table = memo(TableBase) as typeof TableBase;
