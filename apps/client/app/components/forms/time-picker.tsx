import { cn } from "@casbin/ui";
import { Clock } from "lucide-react";

interface TimePickerProps {
  value?: string; // "HH:mm"
  onChange?: (time: string) => void;
  disabled?: boolean;
  className?: string;
}

export function TimePicker({ value = "", onChange, disabled, className }: TimePickerProps) {
  return (
    <div className={cn("relative", className)}>
      <Clock className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground pointer-events-none" />
      <input
        type="time"
        value={value}
        onChange={(e) => onChange?.(e.target.value)}
        disabled={disabled}
        className={cn(
          "flex h-input w-full rounded-md border border-border bg-background pl-10 pr-3 py-[var(--input-padding-y)] text-body ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50 transition-colors duration-normal",
        )}
      />
    </div>
  );
}
