type FetchOptions = RequestInit & {
  headers?: Record<string, string>;
};

const BASE_URL =
  process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080/api/v1";

class ApiClient {
  private async request<T>(
    endpoint: string,
    options: FetchOptions = {}
  ): Promise<T> {
    const url = `${BASE_URL}${endpoint}`;

    // Default headers
    const headers = {
      "Content-Type": "application/json",
      ...options.headers,
    };

    // Note: We don't need to manually send Authorization header if we use HttpOnly cookies.
    // The browser automatically sends cookies with credentials: "include".
    const config = {
      ...options,
      headers,
      credentials: "include" as RequestCredentials, // Important for cookies
    };

    try {
      const response = await fetch(url, config);

      // Handle 401 Unauthorized globally (e.g. redirect to login)
      if (response.status === 401) {
        // Optional: Trigger global logout logic or redirect
        if (
          typeof window !== "undefined" &&
          !window.location.pathname.includes("/login")
        ) {
          // window.location.href = "/login";
        }
      }

      // Try to parse JSON
      let data;
      const contentType = response.headers.get("content-type");
      if (contentType && contentType.includes("application/json")) {
        data = await response.json();
      } else {
        data = await response.text();
      }

      if (!response.ok) {
        // Use backend error message if available
        const errorMessage =
          data?.error ||
          data?.message ||
          `Error ${response.status}: ${response.statusText}`;
        throw new Error(errorMessage);
      }

      return data as T;
    } catch (error) {
      console.error("API Request Failed:", error);
      throw error;
    }
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

  delete<T>(endpoint: string, options?: FetchOptions) {
    return this.request<T>(endpoint, { ...options, method: "DELETE" });
  }
}

export const api = new ApiClient();
