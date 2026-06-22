/**
 * Maps raw upload errors to user-friendly messages.
 * Keeps the surface stable so the UI never shows a "generic" error to users.
 */
export type UploadErrorCode =
  | "file_too_large"
  | "network_error"
  | "permission_denied"
  | "unsupported_type"
  | "quota_exceeded"
  | "canceled"
  | "unknown";

export interface MappedUploadError {
  code: UploadErrorCode;
  message: string;
  retryable: boolean;
}

const MESSAGES: Record<UploadErrorCode, string> = {
  file_too_large: "File is too large to upload.",
  network_error: "Network connection lost. Please try again.",
  permission_denied: "You do not have permission to upload here.",
  unsupported_type: "This file type is not supported.",
  quota_exceeded: "Storage quota exceeded.",
  canceled: "Upload canceled.",
  unknown: "Something went wrong while uploading.",
};

const RETRYABLE: UploadErrorCode[] = ["network_error", "unknown"];

export function mapUploadError(error: unknown): MappedUploadError {
  const raw =
    typeof error === "string"
      ? error
      : error instanceof Error
        ? error.message
        : "";
  const lower = raw.toLowerCase();

  // Try to read a status code if it's an axios/ApiError-like object
  const status =
    (error as { status?: number; response?: { status?: number } } | null)
      ?.status ??
    (error as { response?: { status?: number } } | null)?.response?.status;

  let code: UploadErrorCode = "unknown";

  if (
    status === 401 ||
    status === 403 ||
    lower.includes("forbidden") ||
    lower.includes("permission")
  ) {
    code = "permission_denied";
  } else if (
    status === 413 ||
    lower.includes("too large") ||
    lower.includes("payload")
  ) {
    code = "file_too_large";
  } else if (
    status === 415 ||
    lower.includes("unsupported") ||
    lower.includes("mime")
  ) {
    code = "unsupported_type";
  } else if (
    status === 507 ||
    lower.includes("quota") ||
    lower.includes("storage full")
  ) {
    code = "quota_exceeded";
  } else if (
    lower.includes("network") ||
    lower.includes("failed to fetch") ||
    lower.includes("timeout") ||
    lower.includes("aborted") ||
    lower.includes("err_network")
  ) {
    code = "network_error";
  } else if (lower.includes("cancel")) {
    code = "canceled";
  }

  return {
    code,
    message: MESSAGES[code],
    retryable: RETRYABLE.includes(code),
  };
}

/** Lightweight telemetry hook — wires into existing logging if available, otherwise no-op. */
export type UploadTelemetryEvent =
  | "upload_started"
  | "upload_success"
  | "upload_failed"
  | "upload_canceled"
  | "retry_clicked";

export function logUploadEvent(
  event: UploadTelemetryEvent,
  payload?: Record<string, unknown>,
) {
  if (import.meta.env.DEV) {
    // eslint-disable-next-line no-console
    console.debug(`[upload] ${event}`, payload ?? {});
  }
  // Hook point for future analytics integration — keeps callsites stable.
}
