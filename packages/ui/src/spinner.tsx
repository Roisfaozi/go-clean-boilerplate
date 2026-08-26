import { Loader2 } from "lucide-react";
import { cn } from "./lib/utils";

interface SpinnerProps {
	className?: string;
	size?: "sm" | "md" | "lg";
}

const sizeMap = { sm: "h-4 w-4", md: "h-6 w-6", lg: "h-8 w-8" };

export function Spinner({ className, size = "md" }: SpinnerProps) {
	return (
		<Loader2
			className={cn(
				"text-muted-foreground animate-spin",
				sizeMap[size],
				className,
			)}
		/>
	);
}
