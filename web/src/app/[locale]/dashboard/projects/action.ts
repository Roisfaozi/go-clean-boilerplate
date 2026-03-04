"use server";

import { revalidatePath } from "next/cache";
import { cookies } from "next/headers";
import { redirect } from "next/navigation";
import { z } from "zod";
import { projectsApi } from "~/lib/api/projects";
import { authActionClient } from "~/lib/client/safe-action";

async function getOrgId() {
  const cookieStore = await cookies();
  return (await cookieStore).get("organization_id")?.value || "";
}

const projectSchema = z.object({
  name: z.string().min(1, "Name is required"),
  domain: z.string().min(1, "Domain is required"),
});

export const createProjectAction = authActionClient
  .schema(projectSchema)
  .metadata({ actionName: "createProject" })
  .action(async ({ parsedInput }) => {
    const orgId = await getOrgId();
    if (!orgId) throw new Error("No organization selected");

    const result = await projectsApi.create(parsedInput);
    revalidatePath(`/dashboard/projects`);
    return result;
  });

export const updateProjectAction = authActionClient
  .schema(projectSchema.extend({ id: z.string() }))
  .metadata({ actionName: "updateProject" })
  .action(async ({ parsedInput }) => {
    const { id, ...payload } = parsedInput;
    const orgId = await getOrgId();
    if (!orgId) throw new Error("No organization selected");

    const result = await projectsApi.update(id, payload);
    revalidatePath(`/dashboard/projects`);
    return result;
  });

export const deleteProjectAction = authActionClient
  .schema(z.object({ id: z.string() }))
  .metadata({ actionName: "deleteProject" })
  .action(async ({ parsedInput }) => {
    const orgId = await getOrgId();
    if (!orgId) throw new Error("No organization selected");

    await projectsApi.delete(parsedInput.id);
    revalidatePath(`/dashboard/projects`);
    redirect("/dashboard/projects");
  });

// Non-action helpers for Server Components
export async function checkIfFreePlanLimitReached() {
  const orgId = await getOrgId();
  if (!orgId) return true;

  try {
    const projects = await projectsApi.getAll();
    return (projects?.length || 0) >= 3;
  } catch (error) {
    return false;
  }
}

export async function getProjects() {
  const orgId = await getOrgId();
  if (!orgId) return [];

  try {
    const projects = await projectsApi.getAll();
    return projects || [];
  } catch (error) {
    console.error("Failed to fetch projects:", error);
    return [];
  }
}

export async function getProjectById(id: string) {
  const orgId = await getOrgId();
  if (!orgId) return null;

  try {
    const project = await projectsApi.getByID(id);
    return project;
  } catch (error) {
    return null;
  }
}
