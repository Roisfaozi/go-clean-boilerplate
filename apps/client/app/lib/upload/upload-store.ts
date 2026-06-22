import { create } from "zustand";
import type {
  UploadItem,
  UploadStatus,
  TusUploaderOptions,
  DuplicateResolution,
} from "./tus-client";
import { createTusUpload } from "./tus-client";
import { toast } from "@casbin/ui";
import { mapUploadError, logUploadEvent } from "./upload-errors";

interface AddFilesOptions {
  targetFolderId?: string;
  targetFolderName?: string;
  /** If set, automatically resolves duplicates without prompting. */
  duplicateResolution?: DuplicateResolution;
}

interface UploadSummary {
  total: number;
  active: number;
  queued: number;
  done: number;
  failed: number;
  canceled: number;
  paused: number;
  /** Average completion percentage across the queue (0–100). */
  progress: number;
  /** Aggregate upload speed across active items, bytes/sec. */
  speedBps: number;
  /** Estimated seconds remaining for the active set, or undefined when unknown. */
  etaSeconds?: number;
}

interface DuplicatePrompt {
  pending: { file: File; options?: AddFilesOptions; existingId: string }[];
}

interface UploadState {
  items: UploadItem[];
  uploadQueue: UploadItem[];
  endpoint: string;
  maxConcurrent: number;
  managerOpen: boolean;
  autoRemoveSuccessMs: number;
  maxAutoRetries: number;
  duplicatePrompt: DuplicatePrompt;
  setEndpoint: (endpoint: string) => void;
  setMaxConcurrent: (value: number) => void;
  setManagerOpen: (open: boolean) => void;
  setAutoRemoveSuccessMs: (ms: number) => void;
  addFiles: (files: File[], options?: AddFilesOptions) => void;
  startUpload: (id: string) => void;
  startAll: () => void;
  pauseUpload: (id: string) => void;
  resumeUpload: (id: string) => void;
  cancelUpload: (id: string) => void;
  cancelAllUploading: () => void;
  cancelAllQueued: () => void;
  retryUpload: (id: string) => void;
  retryAllFailed: () => void;
  removeItem: (id: string) => void;
  clearCompleted: () => void;
  clearCanceled: () => void;
  resolveDuplicate: (
    existingId: string,
    resolution: DuplicateResolution,
  ) => void;
  dismissDuplicatePrompt: () => void;
  getSummary: () => UploadSummary;
  _updateItem: (id: string, patch: Partial<UploadItem>) => void;
  _onItemFinalized: (id: string) => void;
}

let counter = 0;
const MANAGER_OPEN_KEY = "nexus-upload-manager-open";
const AUTO_REMOVE_DEFAULT_MS = 4000;

function getStoredManagerOpen() {
  if (typeof window === "undefined") return false;
  return window.localStorage.getItem(MANAGER_OPEN_KEY) === "true";
}

function getRelativePath(file: File) {
  return (
    (file as File & { webkitRelativePath?: string }).webkitRelativePath ||
    file.name
  );
}

function isSameTargetPath(
  a: UploadItem,
  file: File,
  options?: AddFilesOptions,
) {
  return (
    a.fileName === file.name &&
    (a.targetFolderId ?? null) === (options?.targetFolderId ?? null)
  );
}

function uniqueName(name: string, existing: Set<string>): string {
  if (!existing.has(name)) return name;
  const dot = name.lastIndexOf(".");
  const base = dot > 0 ? name.slice(0, dot) : name;
  const ext = dot > 0 ? name.slice(dot) : "";
  let i = 1;
  while (existing.has(`${base} (${i})${ext}`)) i++;
  return `${base} (${i})${ext}`;
}

export const useUploadStore = create<UploadState>()((set, get) => {
  // Track per-item progress timing for speed/ETA calculations.
  const progressTracker = new Map<
    string,
    { lastBytes: number; lastTs: number }
  >();

  const getOptions = (): TusUploaderOptions => ({
    endpoint: get().endpoint,
    onProgress: (id, progress) => {
      const item = get().items.find((i) => i.id === id);
      if (!item) return;
      const bytes = Math.round((progress / 100) * item.fileSize);
      const now = Date.now();
      const prev = progressTracker.get(id);
      let speedBps = item.speedBps;
      if (prev) {
        const dt = (now - prev.lastTs) / 1000;
        if (dt > 0.25) {
          const inst = (bytes - prev.lastBytes) / dt;
          // Smooth with EMA so the number doesn't jitter wildly.
          speedBps = speedBps ? speedBps * 0.6 + inst * 0.4 : inst;
          progressTracker.set(id, { lastBytes: bytes, lastTs: now });
        }
      } else {
        progressTracker.set(id, { lastBytes: bytes, lastTs: now });
      }
      const remaining = item.fileSize - bytes;
      const etaSeconds =
        speedBps && speedBps > 0
          ? Math.max(0, Math.round(remaining / speedBps))
          : undefined;
      get()._updateItem(id, {
        progress,
        status: "uploading",
        bytesUploaded: bytes,
        speedBps,
        etaSeconds,
        startedAt: item.startedAt ?? now,
      });
    },
    onSuccess: (id, url) => {
      const item = get().items.find((i) => i.id === id);
      progressTracker.delete(id);
      get()._updateItem(id, {
        progress: 100,
        status: "success",
        url,
        uploadedFileId: url,
        responseData: { url },
        completedAt: Date.now(),
        speedBps: undefined,
        etaSeconds: 0,
      });
      logUploadEvent("upload_success", { id, fileName: item?.fileName });
      get().startAll();
      get()._onItemFinalized(id);
      if (item) {
        window.dispatchEvent(
          new CustomEvent("nexus:upload-complete", {
            detail: { item: { ...item, url } },
          }),
        );
      }
    },
    onError: (id, error) => {
      progressTracker.delete(id);
      const mapped = mapUploadError(error);
      const item = get().items.find((i) => i.id === id);
      const retryCount = item?.retryCount ?? 0;
      const canAutoRetry =
        mapped.retryable &&
        !item?.autoRetried &&
        retryCount < get().maxAutoRetries;

      get()._updateItem(id, {
        status: "error",
        error: error.message,
        errorMessage: mapped.message,
        errorCode: mapped.code,
        speedBps: undefined,
        etaSeconds: undefined,
      });
      logUploadEvent("upload_failed", {
        id,
        code: mapped.code,
        message: error.message,
      });

      if (canAutoRetry) {
        // Brief backoff, then retry once for transient errors.
        setTimeout(() => {
          const current = get().items.find((i) => i.id === id);
          if (!current || current.status !== "error") return;
          get()._updateItem(id, { autoRetried: true });
          get().retryUpload(id);
        }, 1500);
      } else {
        if (item) {
          window.dispatchEvent(
            new CustomEvent("nexus:upload-error", {
              detail: { id, code: mapped.code, item },
            }),
          );
        }
      }
      get().startAll();
    },
  });

  return {
    items: [],
    uploadQueue: [],
    endpoint: "/api/v1/uploads",
    maxConcurrent: 3,
    managerOpen: getStoredManagerOpen(),
    autoRemoveSuccessMs: AUTO_REMOVE_DEFAULT_MS,
    maxAutoRetries: 1,
    duplicatePrompt: { pending: [] },

    setEndpoint: (endpoint) => set({ endpoint }),

    setMaxConcurrent: (value) =>
      set({ maxConcurrent: Math.max(1, Math.min(6, value)) }),

    setManagerOpen: (open) => {
      if (typeof window !== "undefined")
        window.localStorage.setItem(MANAGER_OPEN_KEY, String(open));
      set({ managerOpen: open });
    },

    setAutoRemoveSuccessMs: (ms) =>
      set({ autoRemoveSuccessMs: Math.max(0, ms) }),

    addFiles: (files, options) => {
      // Detect duplicates within the same target folder against existing non-canceled items.
      const existing = get().items.filter(
        (i) => i.status !== "canceled" && i.status !== "error",
      );
      const duplicates: { file: File; existingId: string }[] = [];
      const fresh: File[] = [];
      files.forEach((f) => {
        const dup = existing.find((i) => isSameTargetPath(i, f, options));
        if (dup && !options?.duplicateResolution)
          duplicates.push({ file: f, existingId: dup.id });
        else fresh.push(f);
      });

      // If a global resolution was provided, apply it directly.
      const usedNames = new Set(existing.map((i) => i.fileName));
      const resolved: {
        file: File;
        renamedTo?: string;
        replacesId?: string;
      }[] = fresh.map((f) => ({ file: f }));
      if (options?.duplicateResolution) {
        files.forEach((f) => {
          const dup = existing.find((i) => isSameTargetPath(i, f, options));
          if (!dup) return;
          if (options.duplicateResolution === "skip") return;
          if (options.duplicateResolution === "replace") {
            resolved.push({ file: f, replacesId: dup.id });
          } else if (options.duplicateResolution === "keep-both") {
            const newName = uniqueName(f.name, usedNames);
            usedNames.add(newName);
            const renamed = new File([f], newName, {
              type: f.type,
              lastModified: f.lastModified,
            });
            resolved.push({ file: renamed, renamedTo: newName });
          }
        });
      }

      const newItems: UploadItem[] = resolved.map(({ file, replacesId }) => ({
        id: `upload-${++counter}-${Date.now()}`,
        file,
        fileName: file.name,
        fileSize: file.size,
        progress: 0,
        status: "queued" as UploadStatus,
        retryCount: 0,
        targetFolderId: options?.targetFolderId,
        targetFolderName: options?.targetFolderName,
        relativePath: getRelativePath(file),
        createdAt: Date.now(),
        duplicateOf: replacesId,
      }));

      // For "replace", cancel the previously-queued item so the new one supersedes it.
      newItems.forEach((it) => {
        if (it.duplicateOf) get().cancelUpload(it.duplicateOf);
      });

      if (newItems.length > 0) {
        set((s) => ({
          items: [...s.items, ...newItems],
          uploadQueue: [...s.uploadQueue, ...newItems],
        }));
        get().setManagerOpen(true);
        toast.success(
          `${newItems.length} upload${newItems.length > 1 ? "s" : ""} added to queue`,
        );
        newItems.forEach((it) => {
          logUploadEvent("upload_started", {
            id: it.id,
            fileName: it.fileName,
          });
        });
        get().startAll();
      }

      // Surface duplicates to the UI for explicit resolution.
      if (duplicates.length > 0) {
        set((s) => ({
          duplicatePrompt: {
            pending: [
              ...s.duplicatePrompt.pending,
              ...duplicates.map((d) => ({
                file: d.file,
                existingId: d.existingId,
                options,
              })),
            ],
          },
        }));
      }
    },

    startUpload: (id) => {
      const item = get().items.find((i) => i.id === id);
      if (!item || item.status === "uploading" || item.status === "success")
        return;

      const upload = createTusUpload(item, getOptions());
      get()._updateItem(id, {
        upload,
        status: "preparing",
        error: undefined,
        errorCode: undefined,
        startedAt: Date.now(),
      });
      upload.start();
    },

    startAll: () => {
      const { items, maxConcurrent, startUpload } = get();
      const queued = items.filter((i) => i.status === "queued");
      const active = items.filter(
        (i) => i.status === "preparing" || i.status === "uploading",
      ).length;
      const toStart = queued.slice(0, maxConcurrent - active);
      toStart.forEach((i) => {
        startUpload(i.id);
      });
    },

    pauseUpload: (id) => {
      const item = get().items.find((i) => i.id === id);
      if (item?.upload && item.status === "uploading") {
        item.upload.abort();
        get()._updateItem(id, { status: "paused" });
        get().startAll();
      }
    },

    resumeUpload: (id) => {
      const item = get().items.find((i) => i.id === id);
      if (item?.upload && item.status === "paused") {
        get()._updateItem(id, { status: "uploading" });
        item.upload.start();
      }
    },

    cancelUpload: (id) => {
      const item = get().items.find((i) => i.id === id);
      if (!item) return;
      // Skip late callbacks from a finished upload to avoid stale state writes.
      if (item.status === "success" || item.status === "canceled") return;
      if (item?.upload) {
        item.upload.abort();
      }
      progressTracker.delete(id);
      get()._updateItem(id, {
        status: "canceled",
        error: undefined,
        errorCode: undefined,
        upload: undefined,
        speedBps: undefined,
        etaSeconds: undefined,
      });
      logUploadEvent("upload_canceled", { id });
      window.dispatchEvent(
        new CustomEvent("nexus:upload-canceled", { detail: { id } }),
      );
      get().startAll();
    },

    cancelAllUploading: () => {
      get()
        .items.filter(
          (i) =>
            i.status === "uploading" ||
            i.status === "preparing" ||
            i.status === "paused",
        )
        .forEach((i) => {
          get().cancelUpload(i.id);
        });
    },

    cancelAllQueued: () => {
      get()
        .items.filter((i) => i.status === "queued")
        .forEach((i) => {
          get().cancelUpload(i.id);
        });
    },

    retryUpload: (id) => {
      const item = get().items.find((i) => i.id === id);
      if (!item) return;

      logUploadEvent("retry_clicked", { id });
      const upload = createTusUpload(item, getOptions());
      get()._updateItem(id, {
        upload,
        status: "preparing",
        progress: 0,
        bytesUploaded: 0,
        speedBps: undefined,
        etaSeconds: undefined,
        error: undefined,
        errorCode: undefined,
        retryCount: (item.retryCount ?? 0) + 1,
        startedAt: Date.now(),
      });
      upload.start();
    },

    retryAllFailed: () => {
      get()
        .items.filter((i) => i.status === "error")
        .forEach((i) => {
          get().retryUpload(i.id);
        });
    },

    removeItem: (id) => {
      const item = get().items.find((i) => i.id === id);
      if (item?.upload && item.status === "uploading") item.upload.abort();
      progressTracker.delete(id);
      set((s) => ({
        items: s.items.filter((i) => i.id !== id),
        uploadQueue: s.uploadQueue.filter((i) => i.id !== id),
      }));
    },

    clearCompleted: () => {
      set((s) => {
        const uploadQueue = s.uploadQueue.filter(
          (i) => i.status !== "success" && i.status !== "canceled",
        );
        return { items: uploadQueue, uploadQueue };
      });
    },

    clearCanceled: () => {
      set((s) => {
        const uploadQueue = s.uploadQueue.filter(
          (i) => i.status !== "canceled",
        );
        return { items: uploadQueue, uploadQueue };
      });
    },

    resolveDuplicate: (existingId, resolution) => {
      const pending = get().duplicatePrompt.pending.find(
        (p) => p.existingId === existingId,
      );
      if (!pending) return;
      set((s) => ({
        duplicatePrompt: {
          pending: s.duplicatePrompt.pending.filter(
            (p) => p.existingId !== existingId,
          ),
        },
      }));
      if (resolution === "skip") return;
      get().addFiles([pending.file], {
        ...pending.options,
        duplicateResolution: resolution,
      });
    },

    dismissDuplicatePrompt: () => set({ duplicatePrompt: { pending: [] } }),

    getSummary: () => {
      const items = get().items;
      const total = items.length;
      const active = items.filter(
        (i) => i.status === "uploading" || i.status === "preparing",
      ).length;
      const queued = items.filter((i) => i.status === "queued").length;
      const done = items.filter((i) => i.status === "success").length;
      const failed = items.filter((i) => i.status === "error").length;
      const canceled = items.filter((i) => i.status === "canceled").length;
      const paused = items.filter((i) => i.status === "paused").length;
      const progress = total
        ? Math.round(items.reduce((s, i) => s + i.progress, 0) / total)
        : 0;
      const activeItems = items.filter((i) => i.status === "uploading");
      const speedBps = activeItems.reduce((s, i) => s + (i.speedBps ?? 0), 0);
      const remainingBytes = activeItems.reduce(
        (s, i) => s + Math.max(0, i.fileSize - (i.bytesUploaded ?? 0)),
        0,
      );
      const etaSeconds =
        speedBps > 0 ? Math.round(remainingBytes / speedBps) : undefined;
      return {
        total,
        active,
        queued,
        done,
        failed,
        canceled,
        paused,
        progress,
        speedBps,
        etaSeconds,
      };
    },

    _updateItem: (id, patch) => {
      set((s) => ({
        items: s.items.map((i) => (i.id === id ? { ...i, ...patch } : i)),
        uploadQueue: s.uploadQueue.map((i) =>
          i.id === id ? { ...i, ...patch } : i,
        ),
      }));
    },

    _onItemFinalized: (id) => {
      const ms = get().autoRemoveSuccessMs;
      if (ms <= 0) return;
      // Auto-remove succeeded items after a brief delay; failed/canceled items stick around.
      setTimeout(() => {
        const item = get().items.find((i) => i.id === id);
        if (item?.status === "success") get().removeItem(id);
      }, ms);
    },
  };
});
