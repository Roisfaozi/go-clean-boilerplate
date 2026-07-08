import { cn } from "@casbin/ui";

/* ── FormSection ── */
interface FormSectionProps {
	title: string;
	description?: string;
	children: React.ReactNode;
	className?: string;
}

export function FormSection({
	title,
	description,
	children,
	className,
}: FormSectionProps) {
	return (
		<div className={cn("space-y-4", className)}>
			<div className="space-y-1">
				<h3 className="text-h3 text-foreground">{title}</h3>
				{description && (
					<p className="text-small text-muted-foreground">{description}</p>
				)}
			</div>
			<div className="space-y-4">{children}</div>
		</div>
	);
}

/* ── FieldGroup ── */
interface FieldGroupProps {
	layout?: "horizontal" | "vertical" | "inline";
	children: React.ReactNode;
	className?: string;
}

export function FieldGroup({
	layout = "vertical",
	children,
	className,
}: FieldGroupProps) {
	return (
		<div
			className={cn(
				layout === "horizontal" && "grid grid-cols-1 gap-4 md:grid-cols-2",
				layout === "vertical" && "space-y-4",
				layout === "inline" && "flex flex-wrap items-end gap-3",
				className,
			)}
		>
			{children}
		</div>
	);
}
