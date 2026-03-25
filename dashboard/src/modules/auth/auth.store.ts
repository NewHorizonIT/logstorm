import { create } from "zustand";

type AuthUser = {
  id: string;
  email: string;
};

type AuthStore = {
  user: AuthUser | null;
  token: string | null;
  login: (user: AuthUser, token: string) => void;
  logout: () => void;
};

const TOKEN_KEY = "token";

const getInitialToken = () => {
  if (typeof window === "undefined") {
    return null;
  }

  return window.localStorage.getItem(TOKEN_KEY);
};

export const useAuthStore = create<AuthStore>((set) => ({
  user: null,
  token: getInitialToken(),

  login: (user, token) => {
    window.localStorage.setItem(TOKEN_KEY, token);
    set({ user, token });
  },

  logout: () => {
    window.localStorage.removeItem(TOKEN_KEY);
    set({ user: null, token: null });
  },
}));
