import * as React from "react";

import { cn } from "./lib/utils";

const Input = React.forwardRef<HTMLInputElement, React.ComponentProps<"input">>(
	({ className, type, ...props }, ref) => {
		return (
			<input
				type={type}
				className={cn(
					"border-input bg-background ring-offset-background placeholder:text-muted-foreground focus-visible:ring-ring flex w-full border transition-all file:border-0 file:bg-transparent file:text-sm file:font-medium focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:outline-hidden disabled:cursor-not-allowed disabled:opacity-50",
					"h-input rounded-md px-input-padding-x py-spacing-input-padding-y text-body",
					type === "number" && "font-mono", // Geist Mono for numbers as per spec
					className,
				)}
				ref={ref}
				{...props}
			/>
		);
	},
);
Input.displayName = "Input";

export { Input };
