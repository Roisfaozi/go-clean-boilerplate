import {
	AlertDialog,
	AlertDialogAction,
	AlertDialogCancel,
	AlertDialogContent,
	AlertDialogFooter,
	AlertDialogHeader,
	AlertDialogTitle,
	AlertDialogTrigger,
} from "@casbin/ui";
import { Button } from "@casbin/ui";

interface CancelConfirmModalProps {
	reset: () => void;
	isDisabled: boolean;
}

export default function CancelConfirmModal({
	reset,
	isDisabled,
}: CancelConfirmModalProps) {
	return (
		<AlertDialog>
			<AlertDialogTrigger asChild>
				<Button variant="secondary" type="reset" disabled={isDisabled}>
					Reset
				</Button>
			</AlertDialogTrigger>
			<AlertDialogContent>
				<AlertDialogHeader>
					<AlertDialogTitle>
						Are you sure to discard the changes?
					</AlertDialogTitle>
				</AlertDialogHeader>
				<AlertDialogFooter>
					<AlertDialogCancel onClick={reset}>Yes</AlertDialogCancel>
					<AlertDialogAction>No</AlertDialogAction>
				</AlertDialogFooter>
			</AlertDialogContent>
		</AlertDialog>
	);
}
