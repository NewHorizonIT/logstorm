import { useEffect, useState } from "react";

import { getProjectsApi, type ProjectResponse } from "@/lib/api/project";
import { useAuthStore } from "@/store/auth";

export function useProjects() {
  const isAuthReady = useAuthStore((s) => s.isAuthReady);
  const accessToken = useAuthStore((s) => s.accessToken);
  const [projects, setProjects] = useState<ProjectResponse[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!isAuthReady) return;
    if (!accessToken) {
      queueMicrotask(() => setLoading(false));
      return;
    }

    getProjectsApi()
      .then(({ data }) => {
        setProjects(data);
        setLoading(false);
      })
      .catch((err: Error) => {
        setError(err.message);
        setLoading(false);
      });
  }, [isAuthReady, accessToken]);

  return { projects, loading, error };
}
