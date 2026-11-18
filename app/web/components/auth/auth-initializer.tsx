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
        // Interceptor may have refreshed the token — read the latest from store
        const token = useAuthStore.getState().accessToken;
        if (token) {
          setAuth(token, {
            id: data.id,
            email: data.email,
            fullName: data.full_name,
          });
        }
      } catch {
        clearAuth();
      }
    }

    void init();
  }, []); // eslint-disable-line react-hooks/exhaustive-deps

  return null;
}
