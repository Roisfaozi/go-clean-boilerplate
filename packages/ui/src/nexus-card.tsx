import { cn } from "./lib/utils";

interface NexusCardProps extends React.HTMLAttributes<HTMLDivElement> {}

export function NexusCard({ className, ...props }: NexusCardProps) {
  return (
    <div
      className={cn(
        "bg-card text-card-foreground border-border p-card-pad rounded-lg border shadow-sm",
        className,
      )}
      {...props}
    />
  );
}

export function NexusCardHeader({
  className,
  ...props
}: React.HTMLAttributes<HTMLDivElement>) {
  return (
    <div className={cn("flex flex-col gap-1.5 pb-4", className)} {...props} />
  );
}

export function NexusCardTitle({
  className,
  ...props
}: React.HTMLAttributes<HTMLHeadingElement>) {
  return (
    <h3
      className={cn(
        "text-h3 leading-none font-semibold tracking-tight",
        className,
      )}
      {...props}
    />
  );
}

export function NexusCardDescription({
  className,
  ...props
}: React.HTMLAttributes<HTMLParagraphElement>) {
  return (
    <p
      className={cn("text-small text-muted-foreground", className)}
      {...props}
    />
  );
}

export function NexusCardContent({
  className,
  ...props
}: React.HTMLAttributes<HTMLDivElement>) {
  return <div className={cn("", className)} {...props} />;
}

export function NexusCardFooter({
  className,
  ...props
}: React.HTMLAttributes<HTMLDivElement>) {
  return <div className={cn("flex items-center pt-4", className)} {...props} />;
}
