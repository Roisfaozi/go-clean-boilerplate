import * as tus from "tus-js-client";
import type { UploadErrorCode } from "./upload-errors";

export type UploadStatus =
  | "queued"
  | "preparing"
  | "uploading"
  | "paused"
  | "success"
  | "error"
  | "canceled";

export type DuplicateResolution = "replace" | "keep-both" | "skip";

export interface UploadItem {
  id: string;
  file: File;
  fileName: string;
  fileSize: number;
  progress: number;
  status: UploadStatus;
  error?: string;
  errorMessage?: string;
  errorCode?: UploadErrorCode;
  retryCount?: number;
  autoRetried?: boolean;
  upload?: tus.Upload;
  abortController?: AbortController;
  url?: string;
  uploadedFileId?: string;
  responseData?: unknown;
  targetFolderId?: string;
  targetFolderName?: string;
  relativePath?: string;
  createdAt?: number;
  completedAt?: number;
  startedAt?: number;
  bytesUploaded?: number;
  speedBps?: number;
  etaSeconds?: number;
  duplicateOf?: string;
  duplicateResolution?: DuplicateResolution;
}

export interface TusUploaderOptions {
  endpoint: string;
  chunkSize?: number;
  headers?: Record<string, string>;
  onProgress?: (id: string, progress: number) => void;
  onSuccess?: (id: string, url: string) => void;
  onError?: (id: string, error: Error) => void;
}

const DEFAULT_CHUNK_SIZE = 5 * 1024 * 1024; // 5MB

export function createTusUpload(
  item: UploadItem,
  options: TusUploaderOptions,
): tus.Upload {
  const upload = new tus.Upload(item.file, {
    endpoint: options.endpoint,
    chunkSize: options.chunkSize ?? DEFAULT_CHUNK_SIZE,
    retryDelays: [0, 1000, 3000, 5000],
    headers: options.headers,
    metadata: {
      filename: item.file.name,
      filetype: item.file.type,
      filesize: String(item.file.size),
    },
    onProgress: (bytesUploaded, bytesTotal) => {
      const pct = Math.round((bytesUploaded / bytesTotal) * 100);
      options.onProgress?.(item.id, pct);
    },
    onSuccess: () => {
      options.onSuccess?.(item.id, upload.url ?? "");
    },
    onError: (error) => {
      options.onError?.(item.id, error);
    },
  });

  return upload;
}
