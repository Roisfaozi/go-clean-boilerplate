export * from "@casbin/api-types";

export interface payload {
  name: string;
  email: string;
  picture?: string;
}

export interface Session {
  user: import("@casbin/api-types").User;
  accessToken: string;
  expiresAt: string;
}

export interface AuthState {
  user: import("@casbin/api-types").User | null;
  isAuthenticated: boolean;
  isLoading: boolean;
}
