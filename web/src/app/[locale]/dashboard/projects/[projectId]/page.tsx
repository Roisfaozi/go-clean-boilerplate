import { notFound } from "next/navigation";
import { getProjectById } from "../action";
import TabSections from "./tab-sections";

export default async function SingleProject({
  params,
}: {
  params: Promise<{ projectId: string }>;
}) {
  const { projectId } = await params;
  const project = await getProjectById(projectId);

  if (!project) {
    notFound();
  }

  return <TabSections project={project} />;
}
