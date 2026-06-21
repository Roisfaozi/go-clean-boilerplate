import { useMemo } from "react";
import {
  ChevronDown,
  ChevronUp,
  UploadCloud,
  X,
  RotateCcw,
  Ban,
  Trash2,
} from "lucide-react";
import { Badge } from "@casbin/ui";
import { Progress } from "@casbin/ui";
import { NexusButton } from "@casbin/ui";
import { NexusCard } from "@casbin/ui";
import { useUploadStore } from "@/lib/upload/upload-store";
import { UploadProgress } from "./upload-progress";

function formatBytesPerSecond(bps: number) {
  if (bps <= 0) return "";
  if (bps < 1024) return `${Math.round(bps)} B/s`;
  if (bps < 1024 * 1024) return `${(bps / 1024).toFixed(1)} KB/s`;
  return `${(bps / 1024 / 1024).toFixed(1)} MB/s`;
}

function formatEta(seconds?: number) {
  if (seconds == null || !isFinite(seconds)) return "";
  if (seconds < 60) return `${seconds}s left`;
  const m = Math.floor(seconds / 60);
  const s = seconds % 60;
  if (m < 60) return `${m}m ${s}s left`;
  const h = Math.floor(m / 60);
  return `${h}h ${m % 60}m left`;
}

export function UploadManager() {
  const items = useUploadStore((s) => s.uploadQueue);
  const open = useUploadStore((s) => s.managerOpen);
  const setOpen = useUploadStore((s) => s.setManagerOpen);
  const clearCompleted = useUploadStore((s) => s.clearCompleted);
  const pauseUpload = useUploadStore((s) => s.pauseUpload);
  const resumeUpload = useUploadStore((s) => s.resumeUpload);
  const cancelUpload = useUploadStore((s) => s.cancelUpload);
  const retryUpload = useUploadStore((s) => s.retryUpload);
  const retryAllFailed = useUploadStore((s) => s.retryAllFailed);
  const cancelAllUploading = useUploadStore((s) => s.cancelAllUploading);
  const cancelAllQueued = useUploadStore((s) => s.cancelAllQueued);

  // Summary derived inline from the same `items` slice so it stays reactive without an extra subscription.
  const summary = useMemo(() => {
    const active = items.filter(
      (i) => i.status === "preparing" || i.status === "uploading",
    ).length;
    const queued = items.filter((i) => i.status === "queued").length;
    const done = items.filter((i) => i.status === "success").length;
    const failed = items.filter((i) => i.status === "error").length;
    const canceled = items.filter((i) => i.status === "canceled").length;
    const progress = items.length
      ? Math.round(items.reduce((s, i) => s + i.progress, 0) / items.length)
      : 0;
    const speedBps = items
      .filter((i) => i.status === "uploading")
      .reduce((s, i) => s + (i.speedBps ?? 0), 0);
    const remainingBytes = items
      .filter((i) => i.status === "uploading")
      .reduce(
        (s, i) => s + Math.max(0, i.fileSize - (i.bytesUploaded ?? 0)),
        0,
      );
    const etaSeconds =
      speedBps > 0 ? Math.round(remainingBytes / speedBps) : undefined;
    return {
      active,
      queued,
      done,
      failed,
      canceled,
      progress,
      speedBps,
      etaSeconds,
      total: items.length,
    };
  }, [items]);

  if (items.length === 0) return null;

  const showSummaryLine = `Uploading ${summary.active} of ${summary.total} files`;
  const speedLabel = formatBytesPerSecond(summary.speedBps);
  const etaLabel = formatEta(summary.etaSeconds);

  return (
    <aside
      className="fixed right-4 bottom-4 z-50 w-[calc(100vw-2rem)] max-w-md"
      aria-label="Upload manager"
      aria-live="polite"
    >
      <NexusCard className="overflow-hidden p-0 shadow-lg">
        <div className="border-border flex items-center justify-between gap-3 border-b px-4 py-3">
          <button
            type="button"
            onClick={() => setOpen(!open)}
            className="flex min-w-0 flex-1 items-center gap-3 text-left"
          >
            <div className="bg-primary/10 text-primary flex h-9 w-9 shrink-0 items-center justify-center rounded-md">
              <UploadCloud className="h-4.5 w-4.5" />
            </div>
            <div className="min-w-0 flex-1">
              <div className="flex items-center gap-2">
                <h3 className="text-foreground truncate text-sm font-semibold">
                  Uploads
                </h3>
                <Badge variant="secondary" className="text-[10px]">
                  {items.length}
                </Badge>
              </div>
              <p className="text-muted-foreground truncate text-[11px]">
                {summary.active > 0 ? showSummaryLine : "No active uploads"}
                {summary.done > 0 ? ` · ${summary.done} completed` : ""}
                {summary.failed > 0 ? ` · ${summary.failed} failed` : ""}
                {summary.canceled > 0 ? ` · ${summary.canceled} canceled` : ""}
                {summary.active > 0 && speedLabel ? ` · ${speedLabel}` : ""}
                {summary.active > 0 && etaLabel ? ` · ${etaLabel}` : ""}
              </p>
            </div>
          </button>
          <div className="flex items-center gap-1">
            <NexusButton
              variant="ghost"
              size="icon"
              className="h-8 w-8"
              onClick={() => setOpen(!open)}
              aria-label={open ? "Collapse uploads" : "Expand uploads"}
            >
              {open ? (
                <ChevronDown className="h-4 w-4" />
              ) : (
                <ChevronUp className="h-4 w-4" />
              )}
            </NexusButton>
            <NexusButton
              variant="ghost"
              size="icon"
              className="h-8 w-8"
              onClick={clearCompleted}
              aria-label="Clear completed uploads"
            >
              <X className="h-4 w-4" />
            </NexusButton>
          </div>
        </div>

        {!open && (
          <Progress value={summary.progress} className="h-1 rounded-none" />
        )}

        {open && (
          <>
            {(summary.failed > 0 ||
              summary.active > 0 ||
              summary.queued > 0 ||
              summary.done > 0) && (
              <div className="border-border bg-muted/20 flex flex-wrap items-center gap-1.5 border-b px-3 py-2">
                {summary.failed > 0 && (
                  <NexusButton
                    size="sm"
                    variant="outline"
                    className="h-7 text-[11px]"
                    onClick={retryAllFailed}
                  >
                    <RotateCcw className="mr-1 h-3 w-3" />
                    Retry all failed ({summary.failed})
                  </NexusButton>
                )}
                {summary.active > 0 && (
                  <NexusButton
                    size="sm"
                    variant="ghost"
                    className="h-7 text-[11px]"
                    onClick={cancelAllUploading}
                  >
                    <Ban className="mr-1 h-3 w-3" />
                    Cancel active
                  </NexusButton>
                )}
                {summary.queued > 0 && (
                  <NexusButton
                    size="sm"
                    variant="ghost"
                    className="h-7 text-[11px]"
                    onClick={cancelAllQueued}
                  >
                    <Ban className="mr-1 h-3 w-3" />
                    Cancel queued
                  </NexusButton>
                )}
                {summary.done > 0 && (
                  <NexusButton
                    size="sm"
                    variant="ghost"
                    className="ml-auto h-7 text-[11px]"
                    onClick={clearCompleted}
                  >
                    <Trash2 className="mr-1 h-3 w-3" />
                    Clear
                  </NexusButton>
                )}
              </div>
            )}
            <div className="max-h-[420px] space-y-1 overflow-y-auto p-2">
              {items.map((item) => (
                <UploadProgress
                  key={item.id}
                  item={item}
                  onPause={() => pauseUpload(item.id)}
                  onResume={() => resumeUpload(item.id)}
                  onCancel={() => cancelUpload(item.id)}
                  onRetry={() => retryUpload(item.id)}
                />
              ))}
            </div>
          </>
        )}
      </NexusCard>
    </aside>
  );
}
