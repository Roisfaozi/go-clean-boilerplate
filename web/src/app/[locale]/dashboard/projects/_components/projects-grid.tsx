"use client";

import { useProjects } from "./projects-context";
import {
  Card,
  CardContent,
  CardFooter,
  CardHeader,
  CardTitle,
} from "~/components/ui/card";
import { Icon } from "~/components/shared/icon";
import Link from "next/link";
import { Badge } from "~/components/ui/badge";
import { CreateProjectDialog } from "./create-project-dialog";
import { memo } from "react";
import { Project } from "~/lib/api/projects";

export function ProjectsGrid() {
  const { projects, isLoading } = useProjects();

  if (isLoading && projects.length === 0) {
    return (
      <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
        {Array.from({ length: 3 }).map((_, i) => (
          <Card key={i} className="animate-pulse">
            <div className="bg-muted h-48" />
          </Card>
        ))}
      </div>
    );
  }

  return (
    <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
      <CreateProjectDialog />
      {projects.map((project) => (
        <MemoizedProjectCard key={project.id} project={project} />
      ))}
    </div>
  );
}

const MemoizedProjectCard = memo(function ProjectCard({
  project,
}: {
  project: Project;
}) {
  return (
    <Card className="group hover:border-primary/50 relative flex flex-col transition-all">
      <CardHeader>
        <div className="flex items-center justify-between">
          <div className="bg-primary/10 text-primary rounded-md p-2">
            <Icon name="LayoutGrid" className="h-5 w-5" />
          </div>
          <Badge
            variant={project.status === "active" ? "success" : "secondary"}
          >
            {project.status}
          </Badge>
        </div>
        <CardTitle className="line-wrap mt-4">{project.name}</CardTitle>
      </CardHeader>
      <CardContent className="flex-1">
        <p className="text-muted-foreground font-mono text-xs">
          {project.domain}
        </p>
      </CardContent>
      <CardFooter className="bg-muted/20 border-t p-4">
        <Link
          href={`/dashboard/projects/${project.id}`}
          className="text-primary flex w-full items-center justify-between text-sm font-medium transition-colors hover:underline"
        >
          View Details
          <Icon name="ArrowRight" className="h-4 w-4" />
        </Link>
      </CardFooter>
    </Card>
  );
});
