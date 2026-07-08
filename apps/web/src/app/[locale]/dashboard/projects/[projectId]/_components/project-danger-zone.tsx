"use client";

import { useProjectDetail } from "./project-detail-context";
import {
	AlertDialog,
	AlertDialogAction,
	AlertDialogCancel,
	AlertDialogContent,
	AlertDialogDescription,
	AlertDialogFooter,
	AlertDialogHeader,
	AlertDialogTitle,
	AlertDialogTrigger,
} from "@casbin/ui";
import { Button } from "@casbin/ui";
import {
	Card,
	CardContent,
	CardDescription,
	CardHeader,
	CardTitle,
} from "@casbin/ui";
import { Icon } from "~/components/shared/icon";

export function ProjectDangerZone() {
	const { deleteProject, isLoading } = useProjectDetail();

	return (
		<Card className="border-destructive/20 bg-destructive/5">
			<CardHeader>
				<CardTitle className="text-destructive">Danger Zone</CardTitle>
				<CardDescription>
					Permanently delete this project and all its associated data.
				</CardDescription>
			</CardHeader>
			<CardContent className="flex items-center justify-between">
				<p className="text-muted-foreground text-sm">
					Once deleted, there is no way to recover this project.
				</p>

				<AlertDialog>
					<AlertDialogTrigger
						render={
							<Button variant="destructive" disabled={isLoading}>
								Delete Project
							</Button>
						}
					/>
					<AlertDialogContent>
						<AlertDialogHeader>
							<AlertDialogTitle>Are you absolutely sure?</AlertDialogTitle>
							<AlertDialogDescription>
								This action cannot be undone. This will permanently delete the
								project and remove all its data from our servers.
							</AlertDialogDescription>
						</AlertDialogHeader>
						<AlertDialogFooter>
							<AlertDialogCancel>Cancel</AlertDialogCancel>
							<AlertDialogAction
								render={
									<Button variant="destructive" onClick={deleteProject}>
										{isLoading && (
											<Icon
												name="Loader"
												className="mr-2 h-4 w-4 animate-spin"
											/>
										)}
										Yes, Delete Project
									</Button>
								}
							/>
						</AlertDialogFooter>
					</AlertDialogContent>
				</AlertDialog>
			</CardContent>
		</Card>
	);
}
