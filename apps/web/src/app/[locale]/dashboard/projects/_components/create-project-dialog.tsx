"use client";

import {
	Button,
	Card,
	Dialog,
	DialogContent,
	DialogDescription,
	DialogFooter,
	DialogHeader,
	DialogTitle,
	DialogTrigger,
	Form,
	FormControl,
	FormField,
	FormItem,
	FormLabel,
	FormMessage,
	Input,
} from "@casbin/ui";
import { zodResolver } from "@hookform/resolvers/zod";
import { useAction } from "next-safe-action/hooks";
import { useState } from "react";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import * as z from "zod";
import { Icon } from "~/components/shared/icon";
import { createProjectAction } from "../action";
import { useProjects } from "./projects-context";

const projectSchema = z.object({
	name: z.string().min(1, { message: "Please enter a project name." }),
	domain: z.string().min(1, { message: "Please enter a project domain." }),
});

type ProjectFormValues = z.infer<typeof projectSchema>;

export function CreateProjectDialog() {
	const [isOpen, setIsOpen] = useState(false);
	const { fetchProjects } = useProjects();

	const { execute, isPending } = useAction(createProjectAction, {
		onSuccess: () => {
			toast.success("Project created successfully");
			form.reset();
			setIsOpen(false);
			fetchProjects();
		},
		onError: ({ error }) => {
			toast.error(error.serverError || "Failed to create project");
		},
	});

	const form = useForm<ProjectFormValues>({
		resolver: zodResolver(projectSchema),
		defaultValues: {
			name: "",
			domain: "",
		},
	});

	async function onSubmit(values: ProjectFormValues) {
		execute(values);
	}

	return (
		<Dialog open={isOpen} onOpenChange={setIsOpen}>
			<DialogTrigger
				nativeButton={false}
				render={
					<Card
						role="button"
						className="hover:bg-accent flex flex-col items-center justify-center gap-y-2.5 p-8 text-center transition-colors"
					>
						<div className="bg-primary/10 text-primary rounded-full p-3">
							<Icon name="Plus" className="h-8 w-8" />
						</div>
						<p className="text-xl font-semibold">Create a project</p>
						<p className="text-muted-foreground text-sm">
							Launch a new environment
						</p>
					</Card>
				}
			/>
			<DialogContent className="sm:max-w-[425px]">
				<DialogHeader>
					<DialogTitle>Create Project</DialogTitle>
					<DialogDescription>
						Add a new project to your organization to start managing it.
					</DialogDescription>
				</DialogHeader>
				<Form {...form}>
					<form
						onSubmit={form.handleSubmit(onSubmit)}
						className="space-y-4 py-4"
					>
						<FormField
							control={form.control}
							name="name"
							render={({ field }) => (
								<FormItem>
									<FormLabel>Project Name</FormLabel>
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
						<DialogFooter className="pt-4">
							<Button disabled={isPending} type="submit" className="w-full">
								{isPending && (
									<Icon name="Loader" className="mr-2 h-4 w-4 animate-spin" />
								)}
								Create Project
							</Button>
						</DialogFooter>
					</form>
				</Form>
			</DialogContent>
		</Dialog>
	);
}
