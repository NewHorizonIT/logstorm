"use client";

import { useState } from "react";
import { Check, CircleNotch, CopySimple } from "@phosphor-icons/react";

import { ApiError } from "@/lib/api/client";
import { createApiKeyApi } from "@/lib/api/api-key";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";

interface Props {
  projectId: string;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onCreated: () => void;
}

export function CreateApiKeyModal({ projectId, open, onOpenChange, onCreated }: Props) {
  const [name, setName] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [createdKey, setCreatedKey] = useState<string | null>(null);
  const [copied, setCopied] = useState(false);

  function handleOpenChange(isOpen: boolean) {
    if (!isOpen) {
      setName("");
      setError(null);
      setCreatedKey(null);
      setCopied(false);
      setSubmitting(false);
    }
    onOpenChange(isOpen);
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    const trimmed = name.trim();
    if (!trimmed) return;
    setSubmitting(true);
    setError(null);
    try {
      const { data } = await createApiKeyApi({ project_id: projectId, name: trimmed });
      setCreatedKey(data.key);
      onCreated();
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Something went wrong. Please try again.");
      setSubmitting(false);
    }
  }

  async function handleCopy() {
    if (!createdKey) return;
    await navigator.clipboard.writeText(createdKey);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Create API Key</DialogTitle>
        </DialogHeader>

        {createdKey ? (
          <div className="flex flex-col gap-4">
            <p className="text-status-warning text-xs">
              Copy this key now — you won&apos;t be able to see it again.
            </p>
            <div className="flex items-center gap-2">
              <code className="bg-surface-active text-text-primary flex-1 overflow-x-auto rounded-none p-2 font-mono text-[11px] break-all">
                {createdKey}
              </code>
              <Button variant="outline" size="icon-sm" onClick={() => void handleCopy()}>
                {copied ? <Check className="text-status-success" /> : <CopySimple />}
              </Button>
            </div>
            <DialogFooter>
              <Button variant="outline" onClick={() => handleOpenChange(false)}>
                Done
              </Button>
            </DialogFooter>
          </div>
        ) : (
          <form onSubmit={(e) => void handleSubmit(e)} className="flex flex-col gap-4">
            <div className="flex flex-col gap-1.5">
              <Label htmlFor="api-key-name">Name</Label>
              <Input
                id="api-key-name"
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="Production SDK"
                autoFocus
              />
            </div>

            {error && (
              <p className="border-status-danger/30 bg-status-danger-bg text-status-danger border px-3 py-2 text-xs">
                {error}
              </p>
            )}

            <DialogFooter>
              <Button type="submit" disabled={submitting || !name.trim()}>
                {submitting && <CircleNotch size={14} className="animate-spin" />}
                Create
              </Button>
            </DialogFooter>
          </form>
        )}
      </DialogContent>
    </Dialog>
  );
}
