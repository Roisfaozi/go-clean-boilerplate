import { useAuthStore } from "~/stores/use-auth-store";

type FetchOptions = RequestInit & {
  headers?: Record<string, string>;
};

const BASE_URL = "/api/v1";

class ApiClient {
  public async request<T>(
    endpoint: string,
    options: FetchOptions = {}
  ): Promise<T> {
    const url = `${BASE_URL}${endpoint}`;

    const isFormData = options.body instanceof FormData;

    // Default headers
    const headers: Record<string, string> = {
      ...options.headers,
    };

    if (!isFormData) {
      headers["Content-Type"] = "application/json";
    }

    // Note: We don't need to manually send Authorization header if we use HttpOnly cookies.
    // The browser automatically sends cookies with credentials: "include".
    const config = {
      ...options,
      headers,
      credentials: "include" as RequestCredentials, // Important for cookies
      body: isFormData
        ? options.body
        : options.body && typeof options.body === "object"
          ? JSON.stringify(options.body)
          : options.body,
    };

    try {
      const response = await fetch(url, config);

      // Handle 401 Unauthorized
      if (response.status === 401) {
        // Prevent infinite loop if /auth/refresh itself returns 401
        if (endpoint === "/auth/refresh") {
          this.handleHardLogout();
          return response as any;
        }

        try {
          const refreshResponse = await fetch(`${BASE_URL}/auth/refresh`, {
            method: "POST",
            credentials: "include",
          });

          if (refreshResponse.ok) {
            return await fetch(url, config).then((res) => {
              if (res.status === 401) {
                this.handleHardLogout();
              }
              return this.parseResponse<T>(res);
            });
          } else {
            this.handleHardLogout();
          }
        } catch (refreshError) {
          console.error("Token refresh failed:", refreshError);
          this.handleHardLogout();
        }
      }

      return await this.parseResponse<T>(response);
    } catch (error) {
      console.error("API Request Failed:", error);
      throw error;
    }
  }

  private handleHardLogout() {
    // 1. Clear Zustand Store
    useAuthStore.getState().logout();

    // 2. Redirect to login if on client side
    if (
      typeof window !== "undefined" &&
      !window.location.pathname.includes("/login")
    ) {
      const returnTo = encodeURIComponent(
        window.location.pathname + window.location.search
      );
      window.location.href = `/login?returnTo=${returnTo}`;
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

        // We use dynamic import for toast to avoid issues in RSC if this file is imported there,
        // although ApiClient is primarily for client-side usage.
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
