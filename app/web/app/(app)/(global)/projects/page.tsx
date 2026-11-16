import Link from "next/link";
import { Plus } from "@phosphor-icons/react/ssr";

import { Badge } from "@/components/ui/badge";

type Environment = "production" | "staging" | "development";

const mockProjects = [
  {
    id: "1",
    name: "Payments API",
    slug: "payments-api",
    environment: "production" as Environment,
    createdAt: "Oct 2, 2026",
  },
  {
    id: "2",
    name: "Web Frontend",
    slug: "web-frontend",
    environment: "staging" as Environment,
    createdAt: "Sep 18, 2026",
  },
  {
    id: "3",
    name: "Auth Service",
    slug: "auth-service",
    environment: "production" as Environment,
    createdAt: "Aug 5, 2026",
  },
];

export default function ProjectsPage() {
  return (
    <div className="p-8">
      {/* Header */}
      <div className="mb-8 flex items-start justify-between">
        <div>
          <h1 className="text-text-primary text-2xl font-semibold">Projects</h1>
          <p className="text-text-muted mt-1 text-sm">
            Each project has its own API key and log stream.
          </p>
        </div>
        <Link
          href="/projects/new"
          className="rounded-input bg-accent-brand flex items-center gap-2 px-4 py-2 text-sm font-medium text-white transition-opacity hover:opacity-90"
        >
          <Plus size={16} weight="bold" />
          New Project
        </Link>
      </div>

      {/* Grid */}
      <div className="grid grid-cols-2 gap-4">
        {mockProjects.map((project) => (
          <Link
            key={project.id}
            href={`/projects/${project.slug}/logs`}
            className="group rounded-card border-border-card bg-surface hover:border-border-strong hover:bg-surface-active flex flex-col gap-3 border p-5 transition-colors"
          >
            <div className="flex items-start justify-between">
              <span className="text-text-primary text-base font-semibold">{project.name}</span>
              <Badge variant={project.environment}>{project.environment}</Badge>
            </div>
            <span className="text-text-subtle font-mono text-xs">{project.slug}</span>
            <span className="text-text-muted text-xs">Created {project.createdAt}</span>
          </Link>
        ))}

        {/* Empty state card */}
        <Link
          href="/projects/new"
          className="rounded-card border-border-strong hover:border-accent-brand hover:bg-surface-active flex flex-col items-center justify-center gap-2 border border-dashed bg-transparent p-5 transition-colors"
        >
          <div className="rounded-pill bg-surface flex size-9 items-center justify-center">
            <Plus size={18} className="text-text-muted" />
          </div>
          <span className="text-text-primary text-sm font-medium">Create your first project</span>
          <span className="text-text-muted text-xs">Set up a project to start sending logs.</span>
        </Link>
      </div>
    </div>
  );
}
