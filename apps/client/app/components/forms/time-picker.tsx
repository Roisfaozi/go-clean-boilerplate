import { cn } from "@casbin/ui";
import { Clock } from "lucide-react";

interface TimePickerProps {
	value?: string; // "HH:mm"
	onChange?: (time: string) => void;
	disabled?: boolean;
	className?: string;
}

export function TimePicker({
	value = "",
	onChange,
	disabled,
	className,
}: TimePickerProps) {
	return (
		<div className={cn("relative", className)}>
			<Clock className="text-muted-foreground pointer-events-none absolute top-1/2 left-3 h-4 w-4 -translate-y-1/2" />
			<input
				type="time"
				value={value}
				onChange={(e) => onChange?.(e.target.value)}
				disabled={disabled}
				className={cn(
					"h-input border-border bg-background text-body ring-offset-background focus-visible:ring-ring duration-normal flex w-full rounded-md border py-[var(--input-padding-y)] pr-3 pl-10 transition-colors focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:outline-none disabled:cursor-not-allowed disabled:opacity-50",
				)}
			/>
		</div>
	);
}
