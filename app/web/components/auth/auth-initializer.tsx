"use client";

import { useEffect } from "react";

import { getMeApi } from "@/lib/api/user";
import { useAuthStore } from "@/store/auth";

export function AuthInitializer() {
  const setAuth = useAuthStore((s) => s.setAuth);
  const clearAuth = useAuthStore((s) => s.clearAuth);

  useEffect(() => {
    async function init() {
      try {
        const { data } = await getMeApi();
        // Token is guaranteed present: either it was in the store before the call,
        // or the 401 interceptor refreshed and stored it before getMeApi resolved.
        const token = useAuthStore.getState().accessToken!;
        setAuth(token, {
          id: data.id,
          email: data.email,
          fullName: data.full_name,
        });
      } catch {
        clearAuth();
      }
    }

    void init();
  }, []); // eslint-disable-line react-hooks/exhaustive-deps

  return null;
}
