import { cva, type VariantProps } from "class-variance-authority";
import { cn } from "./lib/utils";

const nexusBadgeVariants = cva(
	"inline-flex items-center rounded-full px-2.5 py-0.5 text-caption font-medium transition-colors",
	{
		variants: {
			variant: {
				success: "bg-success/10 text-success",
				warning: "bg-warning/10 text-warning",
				danger: "bg-danger/10 text-danger",
				info: "bg-info/10 text-info",
				neutral: "bg-muted text-muted-foreground",
			},
		},
		defaultVariants: { variant: "neutral" },
	},
);

interface NexusBadgeProps
	extends React.HTMLAttributes<HTMLSpanElement>,
		VariantProps<typeof nexusBadgeVariants> {}

export function NexusBadge({ className, variant, ...props }: NexusBadgeProps) {
	return (
		<span
			className={cn(nexusBadgeVariants({ variant }), className)}
			{...props}
		/>
	);
}
