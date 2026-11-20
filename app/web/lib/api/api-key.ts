import { api, apiRequest, wrapKyError } from "@/lib/api/client";

export interface ApiKeyResponse {
  id: string;
  project_id: string;
  name: string;
  key_prefix: string;
  created_at: string;
  revoked_at: string | null;
}

export interface CreateApiKeyResponse extends ApiKeyResponse {
  key: string;
}

export function getApiKeysApi(projectId: string) {
  return apiRequest<{ data: ApiKeyResponse[] }>(`/api-keys?project_id=${projectId}`);
}

export function createApiKeyApi(body: { project_id: string; name: string }) {
  return apiRequest<{ data: CreateApiKeyResponse }>("/api-keys", { method: "POST", body });
}

export function revokeApiKeyApi(id: string): Promise<void> {
  return wrapKyError(() => api(`/api-keys/${id}`, { method: "DELETE" }).then(() => undefined));
}
