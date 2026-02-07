"use server";
import { revalidatePath } from "next/cache";
import { cookies } from "next/headers";
import { redirect } from "next/navigation";
import { z } from "zod";
import { projectsApi } from "~/lib/api/projects";

async function getOrgId() {
  const cookieStore = await cookies();
  return cookieStore.get("organization_id")?.value || "";
}

async function getAuthHeaders() {
  const cookieStore = await cookies();
  return { Cookie: cookieStore.toString() };
}

const createProjectSchema = z.object({
  name: z.string().min(1, "Name is required"),
  domain: z.string().min(1, "Domain is required"),
});

const updateProjectSchema = z.object({
  name: z.string().optional(),
  domain: z.string().optional(),
  status: z.string().optional(),
});

interface Payload {
  name: string;
  domain: string;
}

export async function createProject(payload: Payload) {
  const validatedFields = createProjectSchema.safeParse(payload);
  if (!validatedFields.success) {
    throw new Error(
      "Invalid input: " + validatedFields.error.flatten().fieldErrors
    );
  }

  const orgId = await getOrgId();
  if (!orgId) throw new Error("No organization selected");

  const headers = await getAuthHeaders();
  await projectsApi.create(orgId, validatedFields.data, { headers });
  revalidatePath(`/dashboard/projects`);
}

export async function checkIfFreePlanLimitReached() {
  const orgId = await getOrgId();
  if (!orgId) return true;

  const headers = await getAuthHeaders();
  try {
    const response = await projectsApi.getAll(orgId, { headers });
    // response.data is Project[]
    const count = response?.length || 0;
    return count >= 3;
  } catch (error) {
    return false;
  }
}

export async function getProjects() {
  const orgId = await getOrgId();
  if (!orgId) return [];

  const headers = await getAuthHeaders();
  try {
    const response = await projectsApi.getAll(orgId, { headers });
    return response || [];
  } catch (error) {
    console.error("Failed to fetch projects:", error);
    return [];
  }
}

export async function getProjectById(id: string) {
  const orgId = await getOrgId();
  if (!orgId) return null;

  const headers = await getAuthHeaders();
  try {
    const response = await projectsApi.getByID(orgId, id, { headers });
    return response;
  } catch (error) {
    return null;
  }
}

export async function updateProjectById(id: string, payload: Payload) {
  const validatedFields = updateProjectSchema.safeParse(payload);
  if (!validatedFields.success) {
    throw new Error("Invalid input");
  }

  const orgId = await getOrgId();
  if (!orgId) throw new Error("No organization selected");

  const headers = await getAuthHeaders();
  await projectsApi.update(orgId, id, validatedFields.data, { headers });
  revalidatePath(`/dashboard/projects`);
}

export async function deleteProjectById(id: string) {
  const orgId = await getOrgId();
  if (!orgId) throw new Error("No organization selected");

  const headers = await getAuthHeaders();
  await projectsApi.delete(orgId, id, { headers });
  revalidatePath(`/dashboard/projects`);
  redirect("/dashboard/projects");
}
