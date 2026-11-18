import { create } from "zustand";

interface AuthUser {
  id: string;
  email: string;
  fullName: string;
}

interface AuthState {
  accessToken: string | null;
  user: AuthUser | null;
  setAuth: (token: string, user: AuthUser) => void;
  setToken: (token: string) => void;
  clearAuth: () => void;
}

export const useAuthStore = create<AuthState>((set) => ({
  accessToken: null,
  user: null,
  setAuth: (token, user) => set({ accessToken: token, user }),
  setToken: (token) => set({ accessToken: token }),
  clearAuth: () => set({ accessToken: null, user: null }),
}));
