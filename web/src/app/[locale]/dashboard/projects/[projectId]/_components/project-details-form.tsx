"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";
import * as z from "zod";
import CopyButton from "~/components/shared/copy-button";
import { Icon } from "~/components/shared/icon";
import { Button } from "~/components/ui/button";
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "~/components/ui/form";
import { Input } from "~/components/ui/input";
import { useProjectDetail } from "./project-detail-context";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "~/components/ui/card";

const projectSchema = z.object({
  name: z.string().min(1, { message: "Please enter a project name." }),
  domain: z.string().min(1, { message: "Please enter a project domain." }),
});

type ProjectFormValues = z.infer<typeof projectSchema>;

export function ProjectDetailsForm() {
  const { project, updateProject, isLoading } = useProjectDetail();

  const form = useForm<ProjectFormValues>({
    resolver: zodResolver(projectSchema),
    defaultValues: {
      name: project.name,
      domain: project.domain,
    },
  });

  async function onSubmit(values: ProjectFormValues) {
    try {
      await updateProject(values);
      form.reset(values);
    } catch (error) {
      console.error(error);
    }
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Project Details</CardTitle>
        <CardDescription>
          View and manage the core settings of your project.
        </CardDescription>
      </CardHeader>
      <CardContent>
        <Form {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-6">
            <FormItem>
              <FormLabel>Project ID</FormLabel>
              <FormControl>
                <div className="relative">
                  <Input
                    value={project.id}
                    readOnly
                    disabled
                    className="bg-muted font-mono"
                  />
                  <CopyButton
                    content={project.id}
                    className="absolute top-1/2 right-2 -translate-y-1/2"
                  />
                </div>
              </FormControl>
              <FormMessage />
            </FormItem>

            <FormField
              control={form.control}
              name="name"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Name</FormLabel>
                  <FormControl>
                    <Input placeholder="Acme Dashboard" {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
            <FormField
              control={form.control}
              name="domain"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Domain</FormLabel>
                  <FormControl>
                    <Input placeholder="acme.com" {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
            <Button
              disabled={isLoading || !form.formState.isDirty}
              type="submit"
            >
              {isLoading && (
                <Icon name="Loader" className={"mr-2 h-4 w-4 animate-spin"} />
              )}
              Save Changes
            </Button>
          </form>
        </Form>
      </CardContent>
    </Card>
  );
}
