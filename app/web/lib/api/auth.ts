import { apiRequest } from "@/lib/api/client";

export interface TokenResponse {
  access_token: string;
  token_type: string;
  expires_in: number;
}

export interface UserResponse {
  id: string;
  email: string;
  full_name: string;
  status: string;
  is_verified: boolean;
  created_at: string;
}

export function loginApi(input: { email: string; password: string }) {
  return apiRequest<TokenResponse>("/auth/login", {
    method: "POST",
    body: input,
  });
}

export function registerApi(input: { full_name: string; email: string; password: string }) {
  return apiRequest<UserResponse>("/auth/register", {
    method: "POST",
    body: input,
  });
}

export function refreshApi() {
  return apiRequest<TokenResponse>("/auth/refresh", {
    method: "POST",
    credentials: "include",
  });
}
