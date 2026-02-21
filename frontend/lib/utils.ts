import { clsx, type ClassValue } from "clsx";
import { twMerge } from "tailwind-merge";

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

// ---- Backend Token Cache ----
// Caches the HS256 JWT that the Go backend can verify.
// The token is fetched from /api/auth/token (which decodes NextAuth JWE
// and re-signs as HS256). Cached for 50 minutes (token expires in 1h).
let cachedBackendToken: string | null = null;
let tokenFetchedAt: number = 0;
let tokenFetchPromise: Promise<string | null> | null = null;
const TOKEN_CACHE_DURATION_MS = 50 * 60 * 1000; // 50 minutes

async function getBackendToken(): Promise<string | null> {
  const now = Date.now();

  // Return cached token if still valid
  if (cachedBackendToken && now - tokenFetchedAt < TOKEN_CACHE_DURATION_MS) {
    return cachedBackendToken;
  }

  // Deduplicate concurrent requests — only one fetch at a time
  if (tokenFetchPromise) {
    return tokenFetchPromise;
  }

  tokenFetchPromise = (async () => {
    try {
      const res = await fetch("/api/auth/token");
      if (!res.ok) {
        console.error("[utils] /api/auth/token returned", res.status);
        cachedBackendToken = null;
        return null;
      }
      const data = await res.json();
      if (data.token) {
        cachedBackendToken = data.token;
        tokenFetchedAt = Date.now();
        return data.token;
      }
      return null;
    } catch {
      cachedBackendToken = null;
      return null;
    } finally {
      tokenFetchPromise = null;
    }
  })();

  return tokenFetchPromise;
}

/** Clear the cached backend token (call on logout) */
export function clearBackendTokenCache() {
  cachedBackendToken = null;
  tokenFetchedAt = 0;
  tokenFetchPromise = null;
}

// Enhanced fetch function that includes authentication via Bearer token
export async function fetchWithAuth(url: string, options: RequestInit = {}) {
  let origin = "http://localhost:3000";
  if (typeof window !== "undefined") {
    origin = window.location.origin;
  } else if (process.env.NEXT_PUBLIC_APP_URL) {
    origin = process.env.NEXT_PUBLIC_APP_URL;
  }

  const fullUrl = url.startsWith("http") ? url : `${origin}${url}`;

  // Fetch HS256 token for Go backend authentication
  const backendToken = await getBackendToken();

  const headers: Record<string, string> = {
    "Content-Type": "application/json",
    ...(options.headers as Record<string, string>),
  };

  // Attach Bearer token if available
  if (backendToken) {
    headers["Authorization"] = `Bearer ${backendToken}`;
  }

  const response = await fetch(fullUrl, {
    ...options,
    headers,
    credentials: "include",
  });

  // On 401, invalidate cached token and retry ONCE
  if (response.status === 401 && !url.includes("/api/auth/")) {
    // Clear stale token
    clearBackendTokenCache();

    // Retry with fresh token
    const freshToken = await getBackendToken();
    if (freshToken) {
      headers["Authorization"] = `Bearer ${freshToken}`;
      const retryResponse = await fetch(fullUrl, {
        ...options,
        headers,
        credentials: "include",
      });

      if (retryResponse.status === 401) {
        console.warn(`Unauthorized access to ${url}. Status: ${retryResponse.status}`);
      }

      return retryResponse;
    }

    console.warn(`Unauthorized access to ${url}. Status: ${response.status}`);

    // Force redirect to login if we are in the browser and not already on the auth page
    if (typeof window !== "undefined" && !window.location.pathname.startsWith("/auth")) {
      console.error("[utils] Session expired or invalid. Redirecting to login...");
      // Use window.location to force a full page reload and clear any efficient-state issues
      window.location.href = `/auth/signin?callbackUrl=${encodeURIComponent(window.location.href)}`;
      return response; // technically unreachable after redirect starts, but keeps flow valid
    }
  }

  return response;
}
