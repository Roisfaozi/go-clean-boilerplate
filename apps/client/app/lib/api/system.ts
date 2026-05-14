import { apiClient } from "./client";

export interface SystemHealthResponse {
  status: "OK" | "DEGRADED" | "DOWN";
  details: Record<string, "UP" | "DOWN" | "CONNECTION_ERROR">;
}

export const getSystemHealth = async (): Promise<SystemHealthResponse> => {
  return await apiClient.get<SystemHealthResponse>("/health");
};
