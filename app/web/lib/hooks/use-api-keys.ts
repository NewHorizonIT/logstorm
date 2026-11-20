import { useCallback, useEffect, useState } from "react";

import { getApiKeysApi, type ApiKeyResponse } from "@/lib/api/api-key";
import { useAuthStore } from "@/store/auth";

export function useApiKeys(projectId: string | null) {
  const isAuthReady = useAuthStore((s) => s.isAuthReady);
  const accessToken = useAuthStore((s) => s.accessToken);
  const [state, setState] = useState<{
    keys: ApiKeyResponse[];
    isLoading: boolean;
    isError: string | null;
  }>({ keys: [], isLoading: true, isError: null });
  const [tick, setTick] = useState(0);

  useEffect(() => {
    if (!isAuthReady) return;
    if (!accessToken) {
      queueMicrotask(() => setState({ keys: [], isLoading: false, isError: null }));
      return;
    }
    if (!projectId) return;

    getApiKeysApi(projectId)
      .then(({ data }) => setState({ keys: data, isLoading: false, isError: null }))
      .catch((err: Error) => setState({ keys: [], isLoading: false, isError: err.message }));
  }, [isAuthReady, accessToken, projectId, tick]);

  const refresh = useCallback(() => setTick((t) => t + 1), []);

  return { keys: state.keys, loading: state.isLoading, error: state.isError, refresh };
}
