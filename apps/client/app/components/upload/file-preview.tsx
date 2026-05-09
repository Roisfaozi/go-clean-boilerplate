import { cn } from "@casbin/ui";
import {  NexusCard  } from "@casbin/ui";
import { FileText, Image, Film, Music, Archive, File } from "lucide-react";

interface FilePreviewProps {
  file: File;
  url?: string;
  className?: string;
}

function getFileIcon(type: string) {
  if (type.startsWith("image/")) return Image;
  if (type.startsWith("video/")) return Film;
  if (type.startsWith("audio/")) return Music;
  if (type.includes("zip") || type.includes("rar") || type.includes("tar")) return Archive;
  if (type.includes("pdf") || type.includes("doc") || type.includes("text")) return FileText;
  return File;
}

function formatFileSize(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  return `${(bytes / 1024 / 1024).toFixed(1)} MB`;
}

export function FilePreview({ file, url, className }: FilePreviewProps) {
  const Icon = getFileIcon(file.type);
  const isImage = file.type.startsWith("image/");
  const previewUrl = isImage ? URL.createObjectURL(file) : null;

  return (
    <NexusCard className={cn("overflow-hidden", className)}>
      {/* Preview area */}
      <div className="aspect-video bg-muted/50 flex items-center justify-center overflow-hidden">
        {previewUrl ? (
          <img
            src={previewUrl}
            alt={file.name}
            className="w-full h-full object-cover"
            onLoad={() => URL.revokeObjectURL(previewUrl)}
          />
        ) : (
          <Icon className="h-12 w-12 text-muted-foreground/50" />
        )}
      </div>

      {/* Info */}
      <div className="p-3 space-y-1">
        <p className="text-sm font-medium text-foreground truncate">{file.name}</p>
        <div className="flex items-center gap-2 text-[11px] text-muted-foreground">
          <span>{formatFileSize(file.size)}</span>
          <span>·</span>
          <span className="uppercase">{file.name.split(".").pop()}</span>
        </div>
        {url && (
          <a
            href={url}
            target="_blank"
            rel="noopener noreferrer"
            className="text-[11px] text-primary hover:underline truncate block"
          >
            View uploaded file
          </a>
        )}
      </div>
    </NexusCard>
  );
}
