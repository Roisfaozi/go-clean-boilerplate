import { useUploadStore } from "@/lib/upload/upload-store";
import { UploadProgress } from "./upload-progress";
import { NexusButton } from "@casbin/ui";
import { NexusCard } from "@casbin/ui";
import { Badge } from "@casbin/ui";
import { Trash2, Play } from "lucide-react";

export function UploadQueue() {
	const items = useUploadStore((s) => s.uploadQueue);
	const startAll = useUploadStore((s) => s.startAll);
	const clearCompleted = useUploadStore((s) => s.clearCompleted);
	const pauseUpload = useUploadStore((s) => s.pauseUpload);
	const resumeUpload = useUploadStore((s) => s.resumeUpload);
	const cancelUpload = useUploadStore((s) => s.cancelUpload);
	const retryUpload = useUploadStore((s) => s.retryUpload);

	const uploading = items.filter(
		(i) => i.status === "preparing" || i.status === "uploading",
	).length;
	const queued = items.filter((i) => i.status === "queued").length;
	const completed = items.filter((i) => i.status === "success").length;
	const errors = items.filter((i) => i.status === "error").length;
	const canceled = items.filter((i) => i.status === "canceled").length;

	if (items.length === 0) return null;

	return (
		<NexusCard>
			{/* Header */}
			<div className="border-border flex items-center justify-between border-b px-4 py-3">
				<div className="flex items-center gap-2">
					<h3 className="text-foreground text-sm font-semibold">
						Upload Queue
					</h3>
					<Badge variant="secondary" className="text-[10px]">
						{items.length}
					</Badge>
				</div>
				<div className="flex items-center gap-2 text-[11px]">
					{uploading > 0 && (
						<Badge className="bg-primary/10 text-primary border-primary/20">
							{uploading} uploading
						</Badge>
					)}
					{queued > 0 && <Badge variant="outline">{queued} queued</Badge>}
					{completed > 0 && (
						<Badge className="bg-success/10 text-success border-success/20">
							{completed} done
						</Badge>
					)}
					{errors > 0 && <Badge variant="destructive">{errors} failed</Badge>}
					{canceled > 0 && <Badge variant="outline">{canceled} canceled</Badge>}
				</div>
			</div>

			{/* Items */}
			<div className="max-h-[360px] space-y-1 overflow-y-auto p-2">
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

			{/* Footer actions */}
			<div className="border-border flex items-center justify-between border-t px-4 py-2.5">
				{queued > 0 && (
					<NexusButton size="sm" variant="outline" onClick={startAll}>
						<Play className="mr-1.5 h-3.5 w-3.5" />
						Start All
					</NexusButton>
				)}
				{completed > 0 && (
					<NexusButton
						size="sm"
						variant="ghost"
						onClick={clearCompleted}
						className="ml-auto"
					>
						<Trash2 className="mr-1.5 h-3.5 w-3.5" />
						Clear Completed
					</NexusButton>
				)}
			</div>
		</NexusCard>
	);
}
