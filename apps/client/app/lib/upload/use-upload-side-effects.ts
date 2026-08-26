import { useEffect, useRef } from "react";
import { toast } from "@casbin/ui";
import { useUploadStore } from "./upload-store";

/**
 * Registers global, app-level upload behaviors:
 * - beforeunload warning while uploads are active
 * - batch completion toasts (success / partial failure)
 * Mounted once in the app layout.
 */
export function useUploadSideEffects() {
	const items = useUploadStore((s) => s.items);
	const lastBatchRef = useRef<{
		active: number;
		failed: number;
		done: number;
	} | null>(null);

	// Before-unload protection while uploads are in flight.
	useEffect(() => {
		const handler = (e: BeforeUnloadEvent) => {
			const { items: current } = useUploadStore.getState();
			const inFlight = current.some(
				(i) =>
					i.status === "uploading" ||
					i.status === "preparing" ||
					i.status === "queued",
			);
			if (!inFlight) return;
			e.preventDefault();
			e.returnValue = "Upload masih berjalan. Yakin ingin keluar?";
			return e.returnValue;
		};
		window.addEventListener("beforeunload", handler);
		return () => window.removeEventListener("beforeunload", handler);
	}, []);

	// Global paste-to-upload (Ctrl/Cmd + V).
	// Captures files (and image blobs from screenshot tools) from the clipboard
	// and feeds them into the existing upload queue. Skips when the user is
	// typing in an editable field so it doesn't hijack normal text paste.
	useEffect(() => {
		const isEditableTarget = (target: EventTarget | null) => {
			const el = target as HTMLElement | null;
			if (!el) return false;
			const tag = el.tagName;
			if (tag === "INPUT" || tag === "TEXTAREA" || tag === "SELECT")
				return true;
			if (el.isContentEditable) return true;
			// Walk up: catches contentEditable wrappers (rich text editors, code editors).
			if (
				el.closest?.(
					'[contenteditable="true"], input, textarea, select, [role="textbox"], [role="combobox"], [role="searchbox"]',
				)
			) {
				return true;
			}
			return false;
		};

		const handler = (e: ClipboardEvent) => {
			if (isEditableTarget(e.target)) return;
			// Don't hijack when an overlay that owns its own paste UX is open
			// (command palette, modal dialog, etc.).
			if (
				document.querySelector(
					'[data-state="open"][role="dialog"], [cmdk-root]',
				)
			)
				return;

			const cd = e.clipboardData;
			if (!cd) return;

			const files: File[] = [];

			// Direct files (e.g. copied from file explorer).
			if (cd.files && cd.files.length > 0) {
				for (let i = 0; i < cd.files.length; i++) files.push(cd.files[i]);
			}

			// Image blobs from clipboard items (screenshots, "Copy image").
			if (files.length === 0 && cd.items) {
				for (let i = 0; i < cd.items.length; i++) {
					const item = cd.items[i];
					if (item.kind === "file") {
						const f = item.getAsFile();
						if (f) {
							const ext = f.type.split("/")[1] || "png";
							const named =
								f.name && f.name !== "image.png"
									? f
									: new File([f], `pasted-${Date.now()}.${ext}`, {
											type: f.type,
										});
							files.push(named);
						}
					}
				}
			}

			if (files.length === 0) return;
			e.preventDefault();
			useUploadStore.getState().addFiles(files);
			toast.success(
				files.length === 1
					? `Pasted "${files[0].name}" to uploads`
					: `Pasted ${files.length} files to uploads`,
			);
		};

		// Listen on document so we receive paste events from any focus context
		// (including document.body when nothing is focused) on both Win/Linux
		// (Ctrl+V) and macOS (Cmd+V) — the browser fires `paste` for both.
		document.addEventListener("paste", handler);
		return () => document.removeEventListener("paste", handler);
	}, []);

	// Batch completion notifications: fire when the queue drains.
	useEffect(() => {
		const active = items.filter(
			(i) =>
				i.status === "uploading" ||
				i.status === "preparing" ||
				i.status === "queued",
		).length;
		const failed = items.filter((i) => i.status === "error").length;
		const done = items.filter((i) => i.status === "success").length;
		const prev = lastBatchRef.current;

		// Detect: was active, now zero active, has at least one finished item.
		if (prev && prev.active > 0 && active === 0 && (done > 0 || failed > 0)) {
			if (failed === 0) {
				toast.success(
					`${done} file${done > 1 ? "s" : ""} uploaded successfully`,
				);
			} else if (done === 0) {
				toast.error(`${failed} upload${failed > 1 ? "s" : ""} failed`);
			} else {
				toast.warning(
					`Upload completed with ${failed} error${failed > 1 ? "s" : ""}`,
				);
			}
		}
		lastBatchRef.current = { active, failed, done };
	}, [items]);
}
