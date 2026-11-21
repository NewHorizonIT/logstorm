import { apiRequest } from "@/lib/api/client";

export interface ProjectResponse {
  id: string;
  owner_id: string;
  name: string;
  slug: string;
  description: string;
  environment: string;
  created_at: string;
  updated_at: string;
}

export function getProjectsApi() {
  return apiRequest<{ data: ProjectResponse[] }>("/projects");
}

export function createProjectApi(body: { name: string; slug: string }) {
  return apiRequest<{ data: ProjectResponse }>("/projects", { method: "POST", body });
}

export function getProjectBySlugApi(slug: string) {
  return apiRequest<{ data: ProjectResponse }>(`/projects/slug/${slug}`);
}
