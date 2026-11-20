"use client";

import { useState } from "react";
import { Plus } from "@phosphor-icons/react";

import { type ApiKeyResponse } from "@/lib/api/api-key";
import { useProject } from "@/lib/hooks/use-project";
import { useApiKeys } from "@/lib/hooks/use-api-keys";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { CreateApiKeyModal } from "@/components/api-keys/create-api-key-modal";
import { RevokeApiKeyDialog } from "@/components/api-keys/revoke-api-key-dialog";

function formatDate(iso: string) {
  return new Date(iso).toLocaleDateString("en-US", {
    month: "short",
    day: "numeric",
    year: "numeric",
  });
}

function TableSkeleton() {
  return (
    <TableBody>
      {Array.from({ length: 3 }).map((_, i) => (
        <TableRow key={i}>
          {Array.from({ length: 5 }).map((__, j) => (
            <TableCell key={j}>
              <div className="bg-surface-active h-3 w-24 animate-pulse rounded" />
            </TableCell>
          ))}
        </TableRow>
      ))}
    </TableBody>
  );
}

interface RevokeTarget {
  id: string;
  name: string;
}

interface Props {
  slug: string;
}

export function ApiKeysContent({ slug }: Props) {
  const { project, loading: projectLoading, error: projectError } = useProject(slug);
  const { keys, loading: keysLoading, error: keysError, refresh } = useApiKeys(project?.id ?? null);
  const [createOpen, setCreateOpen] = useState(false);
  const [revokeTarget, setRevokeTarget] = useState<RevokeTarget | null>(null);

  if (projectLoading) {
    return (
      <div className="p-8">
        <div className="mb-8 flex items-center justify-between">
          <div>
            <div className="bg-surface-active h-7 w-32 animate-pulse rounded" />
            <div className="bg-surface-active mt-2 h-4 w-48 animate-pulse rounded" />
          </div>
          <div className="bg-surface-active h-8 w-32 animate-pulse rounded" />
        </div>
        <Table>
          <TableHeader>
            <TableRow>
              {["Name", "Prefix", "Created", "Status", ""].map((h) => (
                <TableHead key={h}>{h}</TableHead>
              ))}
            </TableRow>
          </TableHeader>
          <TableSkeleton />
        </Table>
      </div>
    );
  }

  if (projectError || !project) {
    return (
      <div className="p-8">
        <p className="text-status-danger text-sm">{projectError ?? "Project not found."}</p>
      </div>
    );
  }

  return (
    <div className="p-8">
      <div className="mb-8 flex items-center justify-between">
        <div>
          <h1 className="text-text-primary text-2xl font-semibold">API Keys</h1>
          <p className="text-text-muted mt-1 text-sm">
            Keys used to authenticate log ingestion for{" "}
            <span className="font-mono">{project.name}</span>.
          </p>
        </div>
        <Button onClick={() => setCreateOpen(true)} size="sm">
          <Plus />
          Create API Key
        </Button>
      </div>

      {keysError ? (
        <p className="text-status-danger text-sm">{keysError}</p>
      ) : (
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Name</TableHead>
              <TableHead>Prefix</TableHead>
              <TableHead>Created</TableHead>
              <TableHead>Status</TableHead>
              <TableHead />
            </TableRow>
          </TableHeader>
          {keysLoading ? (
            <TableSkeleton />
          ) : keys.length === 0 ? (
            <TableBody>
              <TableRow>
                <TableCell colSpan={5} className="text-text-muted py-10 text-center">
                  No API keys yet. Create one to start sending logs.
                </TableCell>
              </TableRow>
            </TableBody>
          ) : (
            <TableBody>
              {keys.map((key) => (
                <ApiKeyRow
                  key={key.id}
                  apiKey={key}
                  onRevoke={(id, name) => setRevokeTarget({ id, name })}
                />
              ))}
            </TableBody>
          )}
        </Table>
      )}

      <CreateApiKeyModal
        projectId={project.id}
        open={createOpen}
        onOpenChange={setCreateOpen}
        onCreated={refresh}
      />

      {revokeTarget && (
        <RevokeApiKeyDialog
          keyId={revokeTarget.id}
          keyName={revokeTarget.name}
          open={true}
          onOpenChange={(isOpen) => {
            if (!isOpen) setRevokeTarget(null);
          }}
          onRevoked={refresh}
        />
      )}
    </div>
  );
}

interface ApiKeyRowProps {
  apiKey: ApiKeyResponse;
  onRevoke: (id: string, name: string) => void;
}

function ApiKeyRow({ apiKey, onRevoke }: ApiKeyRowProps) {
  const isRevoked = apiKey.revoked_at !== null;
  return (
    <TableRow className={isRevoked ? "opacity-60" : undefined}>
      <TableCell className="font-medium">{apiKey.name}</TableCell>
      <TableCell>
        <code className="text-text-subtle font-mono">{apiKey.key_prefix}…</code>
      </TableCell>
      <TableCell className="text-text-muted">{formatDate(apiKey.created_at)}</TableCell>
      <TableCell>
        {isRevoked ? (
          <Badge variant="neutral">Revoked</Badge>
        ) : (
          <Badge variant="success">Active</Badge>
        )}
      </TableCell>
      <TableCell className="text-right">
        {!isRevoked && (
          <Button variant="destructive" size="xs" onClick={() => onRevoke(apiKey.id, apiKey.name)}>
            Revoke
          </Button>
        )}
      </TableCell>
    </TableRow>
  );
}
