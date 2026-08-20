import { NexusButton } from "./nexus-button";
import {
	Dialog,
	DialogContent,
	DialogHeader,
	DialogTitle,
	DialogDescription,
	DialogFooter,
} from "./dialog";
import { AlertTriangle } from "lucide-react";

/* ── ConfirmationDialog ── */
interface ConfirmationDialogProps {
	open: boolean;
	onOpenChange: (open: boolean) => void;
	title: string;
	description?: string;
	confirmLabel?: string;
	cancelLabel?: string;
	variant?: "danger" | "primary";
	loading?: boolean;
	onConfirm: () => void;
}

export function ConfirmationDialog({
	open,
	onOpenChange,
	title,
	description,
	confirmLabel = "Confirm",
	cancelLabel = "Cancel",
	variant = "primary",
	loading,
	onConfirm,
}: ConfirmationDialogProps) {
	return (
		<Dialog open={open} onOpenChange={onOpenChange}>
			<DialogContent className="sm:max-w-md">
				<DialogHeader>
					{variant === "danger" && (
						<div className="bg-danger/10 mx-auto mb-4 flex h-12 w-12 items-center justify-center rounded-full">
							<AlertTriangle className="text-danger h-6 w-6" />
						</div>
					)}
					<DialogTitle className="text-center">{title}</DialogTitle>
					{description && (
						<DialogDescription className="text-center">
							{description}
						</DialogDescription>
					)}
				</DialogHeader>
				<DialogFooter className="gap-2 sm:justify-center">
					<NexusButton variant="outline" onClick={() => onOpenChange(false)}>
						{cancelLabel}
					</NexusButton>
					<NexusButton
						variant={variant === "danger" ? "danger" : "primary"}
						loading={loading}
						onClick={onConfirm}
					>
						{confirmLabel}
					</NexusButton>
				</DialogFooter>
			</DialogContent>
		</Dialog>
	);
}

/* ── CommandSearch (overlay search bar) ── */
import {
	CommandDialog,
	CommandEmpty,
	CommandGroup,
	CommandInput,
	CommandItem,
	CommandList,
} from "./command";

export interface CommandSearchItem {
	id: string;
	label: string;
	icon?: React.ReactNode;
	category?: string;
	onSelect: () => void;
}

interface CommandSearchProps {
	items: CommandSearchItem[];
	open?: boolean;
	onOpenChange?: (open: boolean) => void;
	placeholder?: string;
}

export function CommandSearch({
	items,
	open,
	onOpenChange,
	placeholder = "Search…",
}: CommandSearchProps) {
	const groups = items.reduce<Record<string, CommandSearchItem[]>>(
		(acc, item) => {
			const cat = item.category || "Results";
			if (!acc[cat]) acc[cat] = [];
			acc[cat].push(item);
			return acc;
		},
		{},
	);

	return (
		<CommandDialog open={open} onOpenChange={onOpenChange}>
			<CommandInput placeholder={placeholder} />
			<CommandList>
				<CommandEmpty>No results found.</CommandEmpty>
				{Object.entries(groups).map(([cat, groupItems]) => (
					<CommandGroup key={cat} heading={cat}>
						{groupItems.map((item) => (
							<CommandItem
								key={item.id}
								onSelect={() => {
									item.onSelect();
									onOpenChange?.(false);
								}}
							>
								{item.icon && <span className="mr-2">{item.icon}</span>}
								{item.label}
							</CommandItem>
						))}
					</CommandGroup>
				))}
			</CommandList>
		</CommandDialog>
	);
}
