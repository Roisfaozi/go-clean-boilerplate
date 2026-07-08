import { useCallback, useState, useRef } from "react";
import { cn } from "@casbin/ui";
import { useUploadStore } from "@/lib/upload/upload-store";
import { Upload, FileUp } from "lucide-react";

interface FileUploaderProps {
	accept?: string;
	maxFiles?: number;
	maxSize?: number; // bytes
	directory?: boolean;
	targetFolderId?: string;
	targetFolderName?: string;
	title?: string;
	description?: string;
	className?: string;
}

export function FileUploader({
	accept,
	maxFiles = 20,
	maxSize = 100 * 1024 * 1024, // 100MB
	directory = false,
	targetFolderId,
	targetFolderName,
	title,
	description,
	className,
}: FileUploaderProps) {
	const addFiles = useUploadStore((s) => s.addFiles);
	const [dragActive, setDragActive] = useState(false);
	const inputRef = useRef<HTMLInputElement>(null);

	const handleFiles = useCallback(
		(fileList: FileList | null) => {
			if (!fileList) return;
			const files = Array.from(fileList)
				.slice(0, maxFiles)
				.filter((f) => f.size <= maxSize);
			if (files.length > 0) {
				addFiles(files, { targetFolderId, targetFolderName });
			}
		},
		[addFiles, maxFiles, maxSize, targetFolderId, targetFolderName],
	);

	const onDragOver = (e: React.DragEvent) => {
		e.preventDefault();
		e.stopPropagation();
		setDragActive(true);
	};

	const onDragLeave = (e: React.DragEvent) => {
		e.preventDefault();
		e.stopPropagation();
		setDragActive(false);
	};

	const onDrop = (e: React.DragEvent) => {
		e.preventDefault();
		e.stopPropagation();
		setDragActive(false);
		handleFiles(e.dataTransfer.files);
	};

	return (
		<div
			onDragOver={onDragOver}
			onDragLeave={onDragLeave}
			onDrop={onDrop}
			onClick={() => inputRef.current?.click()}
			className={cn(
				"relative flex cursor-pointer flex-col items-center justify-center gap-3 rounded-lg border-2 border-dashed p-8 transition-all duration-200",
				dragActive
					? "border-primary bg-primary/5 scale-[1.01]"
					: "border-border hover:border-primary/50 hover:bg-muted/30",
				className,
			)}
		>
			<div
				className={cn(
					"flex h-12 w-12 items-center justify-center rounded-full transition-colors",
					dragActive
						? "bg-primary/10 text-primary"
						: "bg-muted text-muted-foreground",
				)}
			>
				{dragActive ? (
					<FileUp className="h-6 w-6 animate-bounce" />
				) : (
					<Upload className="h-6 w-6" />
				)}
			</div>
			<div className="text-center">
				<p className="text-foreground text-sm font-medium">
					{dragActive
						? "Drop here"
						: (title ??
							(directory ? "Upload folder" : "Drag & drop files here"))}
				</p>
				<p className="text-muted-foreground mt-1 text-xs">
					{description ??
						`or click to browse · Max ${Math.round(maxSize / 1024 / 1024)}MB per file`}
				</p>
			</div>
			<input
				ref={inputRef}
				type="file"
				multiple
				accept={accept}
				{...(directory ? { webkitdirectory: "", directory: "" } : {})}
				onChange={(e) => handleFiles(e.target.files)}
				className="hidden"
			/>
		</div>
	);
}
