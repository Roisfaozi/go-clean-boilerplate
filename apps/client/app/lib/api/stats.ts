import { apiClient } from "./client";
import {
  dashboardStatsSchema,
  activityMetricSchema,
  systemInsightSchema,
  type DashboardStats,
  type ActivityMetric,
  type SystemInsight,
} from "./schemas";
import { z } from "zod";

export const statsApi = {
  getDashboardSummary: () =>
    apiClient.get<DashboardStats>("/stats/dashboard", dashboardStatsSchema),

  getActivityMetrics: (params?: { from?: string; to?: string }) =>
    apiClient.get<ActivityMetric[]>(
      "/stats/activity",
      z.array(activityMetricSchema),
      { params },
    ),

  getSystemInsights: () =>
    apiClient.get<SystemInsight>("/stats/insights", systemInsightSchema),
};
