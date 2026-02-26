import { api } from "./client";
import { Project } from "~/types";

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
  getAll: (options?: RequestOptions) =>
    api.get<ProjectListResponse>("/projects", options).then((res) => res.data),

  getByID: (id: string, options?: RequestOptions) =>
    api
      .get<{ data: Project }>(`/projects/${id}`, options)
      .then((res) => res.data),

  create: (req: CreateProjectRequest, options?: RequestOptions) =>
    api
      .post<{ data: Project }>("/projects", req, options)
      .then((res) => res.data),

  update: (id: string, req: UpdateProjectRequest, options?: RequestOptions) =>
    api
      .put<{ data: Project }>(`/projects/${id}`, req, options)
      .then((res) => res.data),

  delete: (id: string, options?: RequestOptions) =>
    api.delete(`/projects/${id}`, options),
};
