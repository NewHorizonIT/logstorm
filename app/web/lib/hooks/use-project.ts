import { useEffect, useState } from "react";

import { getProjectBySlugApi, type ProjectResponse } from "@/lib/api/project";
import { useAuthStore } from "@/store/auth";

export function useProject(slug: string) {
  const isAuthReady = useAuthStore((s) => s.isAuthReady);
  const accessToken = useAuthStore((s) => s.accessToken);
  const [project, setProject] = useState<{
    project: ProjectResponse | null;
    isLoading: boolean;
    isError: string | null;
  }>({
    project: null,
    isLoading: true,
    isError: null,
  });

  useEffect(() => {
    if (!isAuthReady) return;
    if (!accessToken) {
      queueMicrotask(() => setProject({ project: null, isLoading: false, isError: null }));
      return;
    }

    getProjectBySlugApi(slug)
      .then(({ data }) => {
        setProject({ project: data, isLoading: false, isError: null });
      })
      .catch((err: Error) => {
        setProject({ project: null, isLoading: false, isError: err.message });
      });
  }, [isAuthReady, accessToken, slug]);

  return { project: project.project, loading: project.isLoading, error: project.isError };
}
