import { Project } from "~/types";
import { api } from "./client";

export type { Project };

export interface CreateProjectRequest {
  name: string;
  domain: string;
}

export interface UpdateProjectRequest {
  name?: string;
  domain?: string;
  status?: string;
}

export interface ProjectListResponse {
  data: Project[];
}

interface RequestOptions {
  headers?: Record<string, string>;
}

export const projectsApi = {
  getAll: (orgId: string, options?: RequestOptions) =>
    api
      .get<ProjectListResponse>("/projects", {
        headers: { "X-Organization-ID": orgId, ...options?.headers },
      })
      .then((res) => res.data),

  getByID: (orgId: string, id: string, options?: RequestOptions) =>
    api
      .get<{ data: Project }>(`/projects/${id}`, {
        headers: { "X-Organization-ID": orgId, ...options?.headers },
      })
      .then((res) => res.data),

  create: (
    orgId: string,
    req: CreateProjectRequest,
    options?: RequestOptions
  ) =>
    api
      .post<{ data: Project }>("/projects", req, {
        headers: { "X-Organization-ID": orgId, ...options?.headers },
      })
      .then((res) => res.data),

  update: (
    orgId: string,
    id: string,
    req: UpdateProjectRequest,
    options?: RequestOptions
  ) =>
    api
      .put<{ data: Project }>(`/projects/${id}`, req, {
        headers: { "X-Organization-ID": orgId, ...options?.headers },
      })
      .then((res) => res.data),

  delete: (orgId: string, id: string, options?: RequestOptions) =>
    api.delete(`/projects/${id}`, {
      headers: { "X-Organization-ID": orgId, ...options?.headers },
    }),
};
