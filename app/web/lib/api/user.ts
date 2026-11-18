import { apiRequest } from "@/lib/api/client";

export interface UserMeResponse {
  id: string;
  email: string;
  full_name: string;
  status: string;
  is_verified: boolean;
  created_at: string;
}

export function getMeApi() {
  return apiRequest<{ data: UserMeResponse }>("/users/me");
}
