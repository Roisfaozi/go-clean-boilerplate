import { cn } from "@casbin/ui";
import { NexusCard } from "@casbin/ui";
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
  if (type.includes("zip") || type.includes("rar") || type.includes("tar"))
    return Archive;
  if (type.includes("pdf") || type.includes("doc") || type.includes("text"))
    return FileText;
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
      <div className="bg-muted/50 flex aspect-video items-center justify-center overflow-hidden">
        {previewUrl ? (
          <img
            src={previewUrl}
            alt={file.name}
            className="h-full w-full object-cover"
            onLoad={() => URL.revokeObjectURL(previewUrl)}
          />
        ) : (
          <Icon className="text-muted-foreground/50 h-12 w-12" />
        )}
      </div>

      {/* Info */}
      <div className="space-y-1 p-3">
        <p className="text-foreground truncate text-sm font-medium">
          {file.name}
        </p>
        <div className="text-muted-foreground flex items-center gap-2 text-[11px]">
          <span>{formatFileSize(file.size)}</span>
          <span>·</span>
          <span className="uppercase">{file.name.split(".").pop()}</span>
        </div>
        {url && (
          <a
            href={url}
            target="_blank"
            rel="noopener noreferrer"
            className="text-primary block truncate text-[11px] hover:underline"
          >
            View uploaded file
          </a>
        )}
      </div>
    </NexusCard>
  );
}
