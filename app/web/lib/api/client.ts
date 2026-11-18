import ky, { HTTPError } from "ky";

import { useAuthStore } from "@/store/auth";

const API_BASE = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";
const API_PREFIX = `${API_BASE}/api/v1/`;

// No interceptors — used for auth endpoints to avoid refresh recursion
export const authApi = ky.create({
  prefix: API_PREFIX,
  credentials: "include",
});

// Shared in-flight refresh promise — prevents concurrent refresh storms
let refreshing: Promise<string> | null = null;

async function doRefresh(): Promise<string> {
  if (!refreshing) {
    refreshing = authApi
      .post("auth/refresh")
      .json<{ data: { access_token: string } }>()
      .then(({ data }) => data.access_token)
      .finally(() => {
        refreshing = null;
      });
  }
  return refreshing;
}

export const api = ky.create({
  prefix: API_PREFIX,
  retry: { limit: 0 },
  hooks: {
    beforeRequest: [
      ({ request }) => {
        const token = useAuthStore.getState().accessToken;
        if (token) request.headers.set("Authorization", `Bearer ${token}`);
      },
    ],
    afterResponse: [
      async ({ request, response, retryCount }) => {
        // Only attempt refresh once (retryCount === 0 prevents infinite loop)
        if (response.status !== 401 || retryCount > 0) return;

        try {
          const token = await doRefresh();
          useAuthStore.getState().setToken(token);

          const headers = new Headers(request.headers);
          headers.set("Authorization", `Bearer ${token}`);
          return ky.retry({ request: new Request(request, { headers }), code: "TOKEN_REFRESHED" });
        } catch {
          useAuthStore.getState().clearAuth();
        }
      },
    ],
  },
});

export { HTTPError };

export async function wrapKyError<T>(fn: () => Promise<T>): Promise<T> {
  try {
    return await fn();
  } catch (err) {
    if (err instanceof HTTPError) {
      const text = await err.response.text().catch(() => (err as Error).message);
      throw new ApiError(err.response.status, text);
    }
    throw err;
  }
}

export class ApiError extends Error {
  constructor(
    public status: number,
    message: string,
  ) {
    super(message);
    this.name = "ApiError";
  }
}

type RequestOptions = {
  method?: string;
  body?: unknown;
};

export async function apiRequest<T>(path: string, options: RequestOptions = {}): Promise<T> {
  const { body, method = "GET" } = options;
  return wrapKyError(() =>
    api(path, {
      method,
      ...(body !== undefined ? { json: body } : {}),
    }).json<T>(),
  );
}
