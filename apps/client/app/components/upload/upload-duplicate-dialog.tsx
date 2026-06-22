import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@casbin/ui";
import { NexusButton } from "@casbin/ui";
import { useUploadStore } from "@/lib/upload/upload-store";

/**
 * Surfaces a prompt when a user uploads a file with the same name in the same target.
 * Resolution: replace, keep both, or skip — applied per pending entry.
 */
export function UploadDuplicateDialog() {
  const pending = useUploadStore((s) => s.duplicatePrompt.pending);
  const resolve = useUploadStore((s) => s.resolveDuplicate);
  const dismiss = useUploadStore((s) => s.dismissDuplicatePrompt);

  const current = pending[0];
  const open = Boolean(current);

  return (
    <AlertDialog open={open} onOpenChange={(o) => !o && dismiss()}>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>File already exists</AlertDialogTitle>
          <AlertDialogDescription>
            <span className="text-foreground font-medium">
              {current?.file.name}
            </span>{" "}
            already exists in this folder. Choose how to handle this upload.
            {pending.length > 1 && (
              <span className="text-muted-foreground mt-2 block text-xs">
                {pending.length - 1} more conflict
                {pending.length - 1 > 1 ? "s" : ""} after this.
              </span>
            )}
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter className="gap-2 sm:gap-2">
          <AlertDialogCancel asChild>
            <NexusButton
              variant="ghost"
              onClick={() => current && resolve(current.existingId, "skip")}
            >
              Skip
            </NexusButton>
          </AlertDialogCancel>
          <NexusButton
            variant="outline"
            onClick={() => current && resolve(current.existingId, "keep-both")}
          >
            Keep both
          </NexusButton>
          <AlertDialogAction asChild>
            <NexusButton
              onClick={() => current && resolve(current.existingId, "replace")}
            >
              Replace
            </NexusButton>
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
}
