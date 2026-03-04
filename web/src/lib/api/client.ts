import { useAuthStore } from "~/stores/use-auth-store";
import { useOrganizationStore } from "~/stores/use-organization-store";

type FetchOptions = RequestInit & {
  headers?: Record<string, string>;
};

const isServer = typeof window === "undefined";

const getBaseUrl = () => {
  if (!isServer) return "/api/v1";
  return process.env.NEXT_PUBLIC_API_URL || "http://127.0.0.1:8080/api/v1";
};

let isRefreshing = false;
let refreshPromise: Promise<boolean> | null = null;
let isLoggingOut = false;

class ApiClient {
  public async request<T>(
    endpoint: string,
    options: FetchOptions = {}
  ): Promise<T> {
    const url = `${getBaseUrl()}${endpoint}`;

    const isFormData = options.body instanceof FormData;

    const headers: Record<string, string> = {
      ...options.headers,
    };

    if (!isFormData) {
      headers["Content-Type"] = "application/json";
    }

    // Automatically inject organization context if available
    if (!isServer) {
      const currentOrg = useOrganizationStore.getState().currentOrganization;
      if (currentOrg) {
        if (!headers["X-Organization-ID"]) {
          headers["X-Organization-ID"] = currentOrg.id;
        }
        if (!headers["X-Organization-Slug"]) {
          headers["X-Organization-Slug"] = currentOrg.slug;
        }
      }
    }

    if (isServer) {
      try {
        const { cookies } = await import("next/headers");
        const cookieStore = await cookies();

        // Pass all cookies
        const cookieStrings = cookieStore
          .getAll()
          .map((c) => `${c.name}=${c.value}`);
        if (cookieStrings.length > 0) {
          headers["Cookie"] = cookieStrings.join("; ");
        }

        // Specifically inject Org headers from specific cookies if they exist
        const orgId = cookieStore.get("organization_id")?.value;
        const orgSlug = cookieStore.get("organization_slug")?.value;

        if (orgId && !headers["X-Organization-ID"]) {
          headers["X-Organization-ID"] = orgId;
        }
        if (orgSlug && !headers["X-Organization-Slug"]) {
          headers["X-Organization-Slug"] = orgSlug;
        }
      } catch (error) {
        // Ignore errors if next/headers is not available
      }
    }

    const config = {
      ...options,
      headers,
      credentials: "include" as RequestCredentials,
      body: isFormData
        ? options.body
        : options.body && typeof options.body === "object"
          ? JSON.stringify(options.body)
          : options.body,
    };

    try {
      const response = await fetch(url, config);

      if (response.status === 401) {
        if (endpoint === "/auth/refresh") {
          return response as any;
        }

        const refreshed = await this.tryRefresh();

        if (refreshed) {
          return await fetch(url, config).then((res) => {
            if (res.status === 401) {
              this.handleHardLogout();
            }
            return this.parseResponse<T>(res);
          });
        } else {
          this.handleHardLogout();
          throw new Error("Session expired");
        }
      }

      return await this.parseResponse<T>(response);
    } catch (error) {
      if (error instanceof Error && error.message === "Session expired") {
        throw error;
      }
      console.error("API Request Failed:", error);
      throw error;
    }
  }

  private async tryRefresh(): Promise<boolean> {
    if (isRefreshing && refreshPromise) {
      return refreshPromise;
    }

    isRefreshing = true;
    refreshPromise = (async () => {
      try {
        const refreshResponse = await fetch(`${getBaseUrl()}/auth/refresh`, {
          method: "POST",
          credentials: "include",
        });
        return refreshResponse.ok;
      } catch {
        return false;
      } finally {
        isRefreshing = false;
        refreshPromise = null;
      }
    })();

    return refreshPromise;
  }

  private handleHardLogout() {
    if (isLoggingOut) return;
    isLoggingOut = true;

    useAuthStore.getState().logout();

    if (
      typeof window !== "undefined" &&
      !window.location.pathname.includes("/login")
    ) {
      const returnTo = encodeURIComponent(
        window.location.pathname + window.location.search
      );

      // Clear HttpOnly cookies via server-side API route, then redirect
      fetch("/api/auth/logout", {
        method: "POST",
        credentials: "include",
      }).finally(() => {
        isLoggingOut = false;
        window.location.href = `/login?returnTo=${returnTo}`;
      });
    } else {
      isLoggingOut = false;
    }
  }

  private async parseResponse<T>(response: Response): Promise<T> {
    let data;
    const contentType = response.headers.get("content-type");
    if (contentType && contentType.includes("application/json")) {
      data = await response.json();
    } else {
      data = await response.text();
    }

    if (!response.ok) {
      if (response.status === 502 || data?.code === "BACKEND_OFFLINE") {
        const message =
          data?.message ||
          "Gagal terhubung ke API Server. Pastikan backend sudah menyala.";

        if (typeof window !== "undefined") {
          import("sonner").then(({ toast }) => {
            toast.error("Koneksi Server Gagal", {
              description: message,
              duration: 5000,
            });
          });
        }
        throw new Error("BACKEND_OFFLINE_ERROR");
      }

      const errorMessage =
        data?.error ||
        data?.message ||
        `Error ${response.status}: ${response.statusText}`;
      throw new Error(errorMessage);
    }

    return data as T;
  }

  get<T>(endpoint: string, options?: FetchOptions) {
    return this.request<T>(endpoint, { ...options, method: "GET" });
  }

  post<T>(endpoint: string, body: any, options?: FetchOptions) {
    return this.request<T>(endpoint, {
      ...options,
      method: "POST",
      body: JSON.stringify(body),
    });
  }

  put<T>(endpoint: string, body: any, options?: FetchOptions) {
    return this.request<T>(endpoint, {
      ...options,
      method: "PUT",
      body: JSON.stringify(body),
    });
  }

  patch<T>(endpoint: string, body: any, options?: FetchOptions) {
    return this.request<T>(endpoint, {
      ...options,
      method: "PATCH",
      body: JSON.stringify(body),
    });
  }

  delete<T>(endpoint: string, options?: FetchOptions) {
    return this.request<T>(endpoint, { ...options, method: "DELETE" });
  }
}

export const api = new ApiClient();
