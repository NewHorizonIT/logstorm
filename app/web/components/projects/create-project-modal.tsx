"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { CircleNotch } from "@phosphor-icons/react";

import { createProjectApi } from "@/lib/api/project";
import { ApiError } from "@/lib/api/client";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";

function toSlug(name: string) {
  return name
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/^-|-$/g, "");
}

interface Props {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function CreateProjectModal({ open, onOpenChange }: Props) {
  const router = useRouter();
  const [name, setName] = useState("");
  const [slug, setSlug] = useState("");
  const [slugTouched, setSlugTouched] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  function handleNameChange(e: React.ChangeEvent<HTMLInputElement>) {
    const val = e.target.value;
    setName(val);
    if (!slugTouched) setSlug(toSlug(val));
  }

  function handleSlugChange(e: React.ChangeEvent<HTMLInputElement>) {
    setSlugTouched(true);
    setSlug(e.target.value);
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    const trimmedName = name.trim();
    const trimmedSlug = slug.trim();
    if (!trimmedName || !trimmedSlug) return;

    if (!/^[a-z0-9]+(?:-[a-z0-9]+)*$/.test(trimmedSlug)) {
      setError("Slug can only contain lowercase letters, numbers, and hyphens.");
      return;
    }

    setSubmitting(true);
    setError(null);

    try {
      const { data } = await createProjectApi({ name: trimmedName, slug: trimmedSlug });
      onOpenChange(false);
      router.push(`/projects/${data.slug}/logs`);
    } catch (err) {
      if (err instanceof ApiError && err.status === 409) {
        setError("A project with this slug already exists.");
      } else if (err instanceof ApiError && err.status === 422) {
        setError("Slug of project invalid.");
      } else {
        setError("Something went wrong. Please try again.");
      }
      setSubmitting(false);
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>New Project</DialogTitle>
        </DialogHeader>

        <form onSubmit={(e) => void handleSubmit(e)} className="flex flex-col gap-4">
          <div className="flex flex-col gap-1.5">
            <Label htmlFor="project-name">Name</Label>
            <Input
              id="project-name"
              value={name}
              onChange={handleNameChange}
              placeholder="Payments API"
              autoFocus
              className="rounded-input border-border-input bg-bg-base text-text-primary placeholder:text-text-disabled focus-visible:border-accent-focus focus-visible:ring-accent-focus/20 h-9 px-3 text-sm"
            />
          </div>

          <div className="flex flex-col gap-1.5">
            <Label htmlFor="project-slug">Slug</Label>
            <Input
              id="project-slug"
              value={slug}
              onChange={handleSlugChange}
              placeholder="payments-api"
              className="rounded-input border-border-input bg-bg-base text-text-primary placeholder:text-text-disabled focus-visible:border-accent-focus focus-visible:ring-accent-focus/20 h-9 px-3 font-mono text-sm"
            />
            <p className="text-text-muted text-xs">
              Used in URLs. Lowercase letters, numbers, and hyphens only — cannot start or end with
              a hyphen.
            </p>
          </div>

          {error && (
            <p className="rounded-control border-status-danger/30 bg-status-danger-bg text-status-danger border px-3 py-2 text-xs">
              {error}
            </p>
          )}

          <DialogFooter>
            <button
              type="submit"
              disabled={submitting || !name.trim() || !slug.trim()}
              className="rounded-input bg-accent-brand flex h-9 w-full items-center justify-center gap-2 text-sm font-medium text-white transition-opacity hover:opacity-90 disabled:cursor-not-allowed disabled:opacity-60"
            >
              {submitting && <CircleNotch size={15} className="animate-spin" />}
              Create Project
            </button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
