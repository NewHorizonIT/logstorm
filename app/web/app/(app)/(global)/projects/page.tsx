"use client";

import { useState } from "react";
import Link from "next/link";
import { Plus } from "@phosphor-icons/react";

import { Badge } from "@/components/ui/badge";
import { CreateProjectModal } from "@/components/projects/create-project-modal";
import { useProjects } from "@/lib/hooks/use-projects";

type EnvVariant = "production" | "staging" | "development";

const ENV_VARIANTS = new Set(["production", "staging", "development"]);

function envVariant(env: string): EnvVariant | undefined {
  return ENV_VARIANTS.has(env) ? (env as EnvVariant) : undefined;
}

function formatDate(iso: string) {
  return new Date(iso).toLocaleDateString("en-US", {
    month: "short",
    day: "numeric",
    year: "numeric",
  });
}

function ProjectsSkeleton() {
  return (
    <div className="grid grid-cols-2 gap-4">
      {Array.from({ length: 4 }).map((_, i) => (
        <div
          key={i}
          className="rounded-card border-border-card bg-surface flex animate-pulse flex-col gap-3 border p-5"
        >
          <div className="flex items-start justify-between">
            <div className="bg-surface-active h-5 w-32 rounded" />
            <div className="bg-surface-active rounded-pill h-5 w-20" />
          </div>
          <div className="bg-surface-active h-3 w-24 rounded" />
          <div className="bg-surface-active h-3 w-20 rounded" />
        </div>
      ))}
    </div>
  );
}

export default function ProjectsPage() {
  const { projects, loading, error } = useProjects();
  const [createOpen, setCreateOpen] = useState(false);

  return (
    <div className="p-8">
      <div className="mb-8">
        <h1 className="text-text-primary text-2xl font-semibold">Projects</h1>
        <p className="text-text-muted mt-1 text-sm">
          Each project has its own API key and log stream.
        </p>
      </div>

      {loading ? (
        <ProjectsSkeleton />
      ) : error ? (
        <p className="text-status-danger text-sm">{error}</p>
      ) : projects.length === 0 ? (
        <button
          onClick={() => setCreateOpen(true)}
          className="rounded-card border-border-strong hover:border-accent-brand hover:bg-surface-active flex w-full flex-col items-center justify-center gap-2 border border-dashed bg-transparent p-10 transition-colors"
        >
          <div className="rounded-pill bg-surface flex size-9 items-center justify-center">
            <Plus size={18} className="text-text-muted" />
          </div>
          <span className="text-text-primary text-sm font-medium">Create your first project</span>
          <span className="text-text-muted text-xs">Set up a project to start sending logs.</span>
        </button>
      ) : (
        <div className="grid grid-cols-2 gap-4">
          {projects.map((project) => {
            const variant = envVariant(project.environment);
            return (
              <Link
                key={project.id}
                href={`/projects/${project.slug}/logs`}
                className="group rounded-card border-border-card bg-surface hover:border-border-strong hover:bg-surface-active flex flex-col gap-3 border p-5 transition-colors"
              >
                <div className="flex items-start justify-between">
                  <span className="text-text-primary text-base font-semibold">{project.name}</span>
                  {variant && <Badge variant={variant}>{project.environment}</Badge>}
                </div>
                <span className="text-text-subtle font-mono text-xs">{project.slug}</span>
                <span className="text-text-muted text-xs">
                  Created {formatDate(project.created_at)}
                </span>
              </Link>
            );
          })}

          <button
            onClick={() => setCreateOpen(true)}
            className="rounded-card border-border-strong hover:border-accent-brand hover:bg-surface-active flex flex-col items-center justify-center gap-2 border border-dashed bg-transparent p-5 transition-colors"
          >
            <div className="rounded-pill bg-surface flex size-9 items-center justify-center">
              <Plus size={18} className="text-text-muted" />
            </div>
            <span className="text-text-primary text-sm font-medium">New project</span>
            <span className="text-text-muted text-xs">Set up a project to start sending logs.</span>
          </button>
        </div>
      )}

      <CreateProjectModal open={createOpen} onOpenChange={setCreateOpen} />
    </div>
  );
}
