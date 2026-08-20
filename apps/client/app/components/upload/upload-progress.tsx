import { cn } from "@casbin/ui";
import { Progress } from "@casbin/ui";
import { NexusButton } from "@casbin/ui";
import type { UploadItem } from "@/lib/upload/tus-client";
import {
	Pause,
	Play,
	X,
	RotateCcw,
	CheckCircle2,
	AlertCircle,
	Clock,
	Loader2,
	Ban,
} from "lucide-react";

interface UploadProgressProps {
	item: UploadItem;
	onPause?: () => void;
	onResume?: () => void;
	onCancel?: () => void;
	onRetry?: () => void;
}

function formatFileSize(bytes: number): string {
	if (bytes < 1024) return `${bytes} B`;
	if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
	return `${(bytes / 1024 / 1024).toFixed(1)} MB`;
}

function formatSpeed(bps?: number): string {
	if (!bps || bps <= 0) return "";
	if (bps < 1024) return `${Math.round(bps)} B/s`;
	if (bps < 1024 * 1024) return `${(bps / 1024).toFixed(1)} KB/s`;
	return `${(bps / 1024 / 1024).toFixed(1)} MB/s`;
}

function formatEta(seconds?: number): string {
	if (seconds == null || !isFinite(seconds)) return "";
	if (seconds < 60) return `${seconds}s left`;
	const m = Math.floor(seconds / 60);
	const s = seconds % 60;
	if (m < 60) return `${m}m ${s}s left`;
	const h = Math.floor(m / 60);
	return `${h}h ${m % 60}m left`;
}

const statusConfig = {
	queued: { icon: Clock, label: "Queued", color: "text-muted-foreground" },
	preparing: { icon: Loader2, label: "Preparing", color: "text-primary" },
	uploading: { icon: Loader2, label: "Uploading", color: "text-primary" },
	paused: { icon: Pause, label: "Paused", color: "text-warning" },
	success: { icon: CheckCircle2, label: "Success", color: "text-success" },
	error: { icon: AlertCircle, label: "Error", color: "text-destructive" },
	canceled: { icon: Ban, label: "Canceled", color: "text-muted-foreground" },
};

export function UploadProgress({
	item,
	onPause,
	onResume,
	onCancel,
	onRetry,
}: UploadProgressProps) {
	const config = statusConfig[item.status];
	const StatusIcon = config.icon;

	return (
		<div className="bg-muted/30 group hover:bg-muted/50 flex items-center gap-3 rounded-md px-3 py-2.5 transition-colors">
			{/* File icon / status */}
			<StatusIcon
				className={cn(
					"h-4.5 w-4.5 shrink-0",
					config.color,
					(item.status === "preparing" || item.status === "uploading") &&
						"animate-spin",
				)}
			/>

			{/* Info */}
			<div className="min-w-0 flex-1 space-y-1">
				<div className="flex items-center justify-between gap-2">
					<p className="text-foreground truncate text-sm font-medium">
						{item.relativePath || item.fileName}
					</p>
					<span className="text-muted-foreground shrink-0 text-[11px]">
						{formatFileSize(item.fileSize)}
					</span>
				</div>

				{item.targetFolderName && (
					<p className="text-muted-foreground truncate text-[11px]">
						Target: {item.targetFolderName}
					</p>
				)}

				{(item.status === "preparing" ||
					item.status === "uploading" ||
					item.status === "paused") && (
					<div className="flex items-center gap-2">
						<Progress value={item.progress} className="h-1.5 flex-1" />
						<span className="text-muted-foreground w-8 text-right text-[11px] font-medium">
							{item.progress}%
						</span>
					</div>
				)}

				{item.status === "uploading" &&
					(item.speedBps || item.etaSeconds != null) && (
						<p className="text-muted-foreground truncate text-[10px]">
							{[formatSpeed(item.speedBps), formatEta(item.etaSeconds)]
								.filter(Boolean)
								.join(" · ")}
						</p>
					)}

				{item.status === "error" && (item.errorMessage || item.error) && (
					<p className="text-destructive truncate text-[11px]">
						{item.errorMessage || item.error}
						{item.retryCount && item.retryCount > 0
							? ` · attempt ${item.retryCount + 1}`
							: ""}
					</p>
				)}
			</div>

			{/* Actions */}
			<div className="flex items-center gap-0.5 opacity-0 transition-opacity group-hover:opacity-100">
				{item.status === "uploading" && onPause && (
					<NexusButton
						variant="ghost"
						size="icon"
						className="h-7 w-7"
						onClick={onPause}
						aria-label={`Pause ${item.fileName}`}
					>
						<Pause className="h-3.5 w-3.5" />
					</NexusButton>
				)}
				{item.status === "paused" && onResume && (
					<NexusButton
						variant="ghost"
						size="icon"
						className="h-7 w-7"
						onClick={onResume}
						aria-label={`Resume ${item.fileName}`}
					>
						<Play className="h-3.5 w-3.5" />
					</NexusButton>
				)}
				{item.status === "error" && onRetry && (
					<NexusButton
						variant="ghost"
						size="icon"
						className="h-7 w-7"
						onClick={onRetry}
						aria-label={`Retry ${item.fileName}`}
					>
						<RotateCcw className="h-3.5 w-3.5" />
					</NexusButton>
				)}
				{item.status !== "success" &&
					item.status !== "canceled" &&
					onCancel && (
						<NexusButton
							variant="ghost"
							size="icon"
							className="h-7 w-7"
							onClick={onCancel}
							aria-label={`Cancel ${item.fileName}`}
						>
							<X className="text-destructive h-3.5 w-3.5" />
						</NexusButton>
					)}
			</div>
		</div>
	);
}
