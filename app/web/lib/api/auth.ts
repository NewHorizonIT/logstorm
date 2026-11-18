import { authApi, wrapKyError } from "@/lib/api/client";

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
  return wrapKyError(() =>
    authApi.post("auth/login", { json: input }).json<{ data: TokenResponse }>(),
  );
}

export function registerApi(input: { full_name: string; email: string; password: string }) {
  return wrapKyError(() =>
    authApi.post("auth/register", { json: input }).json<{ data: UserResponse }>(),
  );
}
