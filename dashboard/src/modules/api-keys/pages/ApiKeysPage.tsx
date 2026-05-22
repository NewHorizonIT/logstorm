import { Button } from "@/components/Button";
import { Card } from "@/components/Card";
import { Badge } from "@/components/Badge";
import { useState } from "react";

type ApiKeyRow = {
  id: string;
  name: string;
  key: string;
  permission: string;
  created_at: string;
};

const mockApiKeys: ApiKeyRow[] = [
  {
    id: "key-1",
    name: "Ingestion Service",
    key: "lsk_live_8fA2...93Dd",
    permission: "write",
    created_at: "2026-03-24 09:10:22",
  },
  {
    id: "key-2",
    name: "Dashboard Client",
    key: "lsk_live_3Bc7...41Zm",
    permission: "read",
    created_at: "2026-03-22 14:03:11",
  },
  {
    id: "key-3",
    name: "Alert Worker",
    key: "lsk_live_1Qn5...88Xv",
    permission: "read_write",
    created_at: "2026-03-20 08:44:57",
  },
];

const ApiKeysPage = () => {
  const [keys, setKeys] = useState<ApiKeyRow[]>(mockApiKeys);

  const handleRevoke = (id: string) => {
    const ok = window.confirm("Are you sure? This cannot be undone.");
    if (!ok) return;
    setKeys((prev) => prev.filter((k) => k.id !== id));
  };

  const handleCopy = (key: string) => {
    navigator.clipboard?.writeText(key);
  };

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between gap-3">
        <h1 className="text-lg font-semibold text-[--color-text-primary]">API Keys</h1>
        <Button>Create API Key</Button>
      </div>

      <Card>
        <div className="overflow-x-auto rounded-[--radius-md] border border-[--color-border]">
          <table className="w-full min-w-120 text-left text-sm">
            <thead className="bg-[color-mix(in_srgb,var(--color-card),white_5%)] text-[--color-text-secondary]">
                <tr>
                  <th className="px-4 py-3 font-medium">Name</th>
                  <th className="px-4 py-3 font-medium">Key</th>
                  <th className="px-4 py-3 font-medium">Permission</th>
                  <th className="px-4 py-3 font-medium">Created At</th>
                  <th className="px-4 py-3 font-medium">Actions</th>
                </tr>
            </thead>
            <tbody>
                {keys.map((row) => (
                  <tr key={row.id} className="border-t border-[--color-border]">
                    <td className="px-4 py-3 text-[--color-text-primary]">{row.name}</td>
                    <td className="px-4 py-3 font-mono text-[--color-text-secondary] flex items-center gap-3">
                      <span className="truncate">{row.key}</span>
                      <button
                        type="button"
                        className="rounded-[--radius-md] border border-[--color-border] px-2 py-1 text-xs"
                        onClick={() => handleCopy(row.key)}
                      >
                        Copy
                      </button>
                    </td>
                    <td className="px-4 py-3 text-[--color-text-secondary]">
                      <Badge type={row.permission === "read" ? "info" : row.permission === "write" ? "warn" : "error"}>
                        {row.permission}
                      </Badge>
                    </td>
                    <td className="px-4 py-3 text-[--color-text-secondary]">{row.created_at}</td>
                    <td className="px-4 py-3">
                      <button
                        type="button"
                        className="rounded-[--radius-md] border border-rose-500/30 px-2 py-1 text-xs text-rose-200"
                        onClick={() => handleRevoke(row.id)}
                      >
                        Revoke
                      </button>
                    </td>
                  </tr>
                ))}
            </tbody>
          </table>
        </div>
      </Card>
    </div>
  );
};

export default ApiKeysPage;
