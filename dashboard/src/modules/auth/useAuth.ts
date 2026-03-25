// src/modules/auth/hooks/useAuth.ts

import { useAuthStore } from "./authStore";
import { authApi } from "./services";

export const useAuth = () => {
  const { user, token, setAuth, logout } = useAuthStore();

  const login = async (email: string, password: string) => {
    const res = await authApi.login({ email, password });
    const { user, token } = res.data;

    setAuth(user, token);
  };

  const register = async (email: string, password: string) => {
    await authApi.register({ email, password });
  };

  return {
    user,
    token,
    login,
    register,
    logout,
  };
};