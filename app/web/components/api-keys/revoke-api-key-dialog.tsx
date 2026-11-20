"use client";

import { useState } from "react";
import { CircleNotch } from "@phosphor-icons/react";

import { ApiError } from "@/lib/api/client";
import { revokeApiKeyApi } from "@/lib/api/api-key";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";

interface Props {
  keyId: string;
  keyName: string;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onRevoked: () => void;
}

export function RevokeApiKeyDialog({ keyId, keyName, open, onOpenChange, onRevoked }: Props) {
  const [revoking, setRevoking] = useState(false);
  const [error, setError] = useState<string | null>(null);

  function handleOpenChange(isOpen: boolean) {
    if (!isOpen) setError(null);
    onOpenChange(isOpen);
  }

  async function handleRevoke() {
    setRevoking(true);
    setError(null);
    try {
      await revokeApiKeyApi(keyId);
      onRevoked();
      onOpenChange(false);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Something went wrong. Please try again.");
      setRevoking(false);
    }
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Revoke API Key</DialogTitle>
          <DialogDescription>
            Revoke <strong>{keyName}</strong>? Services using this key will immediately lose access.
            This cannot be undone.
          </DialogDescription>
        </DialogHeader>

        {error && (
          <p className="border-status-danger/30 bg-status-danger-bg text-status-danger border px-3 py-2 text-xs">
            {error}
          </p>
        )}

        <DialogFooter>
          <Button variant="outline" onClick={() => handleOpenChange(false)} disabled={revoking}>
            Cancel
          </Button>
          <Button variant="destructive" onClick={() => void handleRevoke()} disabled={revoking}>
            {revoking && <CircleNotch size={14} className="animate-spin" />}
            Revoke
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
