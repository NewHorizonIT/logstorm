"use client";

import { useEffect } from "react";

import { getMeApi } from "@/lib/api/user";
import { useAuthStore } from "@/store/auth";

export function AuthInitializer() {
  const setAuth = useAuthStore((s) => s.setAuth);
  const clearAuth = useAuthStore((s) => s.clearAuth);
  const setAuthReady = useAuthStore((s) => s.setAuthReady);

  useEffect(() => {
    async function init() {
      try {
        const { data } = await getMeApi();
        const token = useAuthStore.getState().accessToken!;
        setAuth(token, {
          id: data.id,
          email: data.email,
          fullName: data.full_name,
        });
      } catch {
        clearAuth();
      } finally {
        setAuthReady();
      }
    }

    void init();
  }, []); // eslint-disable-line react-hooks/exhaustive-deps

  return null;
}
