import { apiClient } from "@/lib/api/client";

export interface UploadRequest {
	file: File;
	targetFolderId?: string;
	relativePath?: string;
	signal?: AbortSignal;
	onProgress?: (progress: number) => void;
}

export interface UploadResult {
	url: string;
}

const UPLOAD_ENDPOINT = "/uploads";

export const uploadService = {
	async uploadFile({
		file,
		targetFolderId,
		relativePath,
		signal,
		onProgress,
	}: UploadRequest): Promise<UploadResult> {
		const form = new FormData();
		form.append("file", file);
		if (targetFolderId) form.append("folderId", targetFolderId);
		if (relativePath) form.append("relativePath", relativePath);

		return apiClient.post<UploadResult>(UPLOAD_ENDPOINT, form, undefined, {
			signal,
			headers: { "Content-Type": "multipart/form-data" },
			onUploadProgress: (event) => {
				if (!event.total) return;
				onProgress?.(Math.round((event.loaded / event.total) * 100));
			},
		});
	},
};
