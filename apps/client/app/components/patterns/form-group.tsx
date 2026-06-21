import { cn } from "@casbin/ui";

interface FormGroupProps {
  label: string;
  description?: string;
  error?: string;
  required?: boolean;
  children: React.ReactNode;
  className?: string;
}

export function FormGroup({
  label,
  description,
  error,
  required,
  children,
  className,
}: FormGroupProps) {
  return (
    <div className={cn("space-y-1.5", className)}>
      <label className="text-small text-foreground font-medium">
        {label}
        {required && <span className="text-danger ml-0.5">*</span>}
      </label>
      {description && (
        <p className="text-caption text-muted-foreground">{description}</p>
      )}
      {children}
      {error && <p className="text-caption text-danger">{error}</p>}
    </div>
  );
}
