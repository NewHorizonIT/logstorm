import { create } from "zustand";
import { devtools } from "zustand/middleware";

interface AuthUser {
  id: string;
  email: string;
  fullName: string;
}

interface AuthState {
  accessToken: string | null;
  user: AuthUser | null;
  isAuthReady: boolean;
  setAuth: (token: string, user: AuthUser) => void;
  setToken: (token: string) => void;
  clearAuth: () => void;
  setAuthReady: () => void;
}

export const useAuthStore = create<AuthState>()(
  devtools(
    (set) => ({
      accessToken: null,
      user: null,
      isAuthReady: false,
      setAuth: (token, user) => set({ accessToken: token, user }, false, "setAuth"),
      setToken: (token) => set({ accessToken: token }, false, "setToken"),
      clearAuth: () =>
        set({ accessToken: null, user: null, isAuthReady: false }, false, "clearAuth"),
      setAuthReady: () => set({ isAuthReady: true }, false, "setAuthReady"),
    }),
    {
      name: "AuthStore",
      enabled: process.env.NODE_ENV !== "production",
    },
  ),
);
