import api from "@/services/api";

export type AuthUser = {
  id: string;
  email: string;
};

export type LoginRequest = {
  email: string;
  password: string;
};

export type RegisterRequest = {
  email: string;
  password: string;
};

export type LoginResponse = {
  user: AuthUser;
  token: string;
};

export type RegisterResponse = {
  message: string;
};

export type ProfileResponse = {
  user: AuthUser;
};

export const authApi = {
  login: (data: LoginRequest) => api.post<LoginResponse>("/auth/login", data),

  register: (data: RegisterRequest) =>
    api.post<RegisterResponse>("/auth/register", data),

  getProfile: () => api.get<ProfileResponse>("/auth/me"),
};