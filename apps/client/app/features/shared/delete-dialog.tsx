import { useState } from "react";
import {
	AlertDialog,
	AlertDialogContent,
	AlertDialogHeader,
	AlertDialogTitle,
	AlertDialogDescription,
	AlertDialogFooter,
	AlertDialogCancel,
} from "@casbin/ui";
import { NexusButton } from "@casbin/ui";
import { Loader2, AlertTriangle } from "lucide-react";

interface DeleteDialogProps {
	open: boolean;
	onOpenChange: (open: boolean) => void;
	resourceName: string;
	itemName?: string;
	onConfirm: () => Promise<void> | void;
	loading?: boolean;
}

export function DeleteDialog({
	open,
	onOpenChange,
	resourceName,
	itemName,
	onConfirm,
	loading,
}: DeleteDialogProps) {
	const [deleting, setDeleting] = useState(false);
	const isLoading = deleting || !!loading;

	const handleConfirm = async () => {
		setDeleting(true);
		try {
			await onConfirm();
			onOpenChange(false);
		} catch {
			// error handled by caller
		} finally {
			setDeleting(false);
		}
	};

	return (
		<AlertDialog open={open} onOpenChange={onOpenChange}>
			<AlertDialogContent>
				<AlertDialogHeader>
					<div className="flex items-center gap-3">
						<div className="bg-destructive/10 flex h-10 w-10 items-center justify-center rounded-full">
							<AlertTriangle className="text-destructive h-5 w-5" />
						</div>
						<div>
							<AlertDialogTitle>Delete {resourceName}</AlertDialogTitle>
							<AlertDialogDescription className="mt-1">
								Are you sure you want to delete
								{itemName
									? ` "${itemName}"`
									: ` this ${resourceName.toLowerCase()}`}
								? This action cannot be undone.
							</AlertDialogDescription>
						</div>
					</div>
				</AlertDialogHeader>
				<AlertDialogFooter>
					<AlertDialogCancel disabled={isLoading}>Cancel</AlertDialogCancel>
					<NexusButton
						variant="danger"
						onClick={handleConfirm}
						disabled={isLoading}
					>
						{isLoading && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
						Delete
					</NexusButton>
				</AlertDialogFooter>
			</AlertDialogContent>
		</AlertDialog>
	);
}
