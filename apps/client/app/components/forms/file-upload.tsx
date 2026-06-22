import { useCallback, useState, useRef } from "react";
import { cn } from "@casbin/ui";
import { Upload, X, File } from "lucide-react";

interface FileUploadProps {
  accept?: string;
  multiple?: boolean;
  maxSize?: number; // bytes
  onChange?: (files: File[]) => void;
  className?: string;
}

export function FileUpload({
  accept,
  multiple,
  maxSize = 10 * 1024 * 1024,
  onChange,
  className,
}: FileUploadProps) {
  const [files, setFiles] = useState<File[]>([]);
  const [dragOver, setDragOver] = useState(false);
  const inputRef = useRef<HTMLInputElement>(null);

  const handleFiles = useCallback(
    (incoming: FileList | null) => {
      if (!incoming) return;
      const valid = Array.from(incoming).filter((f) => f.size <= maxSize);
      const next = multiple ? [...files, ...valid] : valid.slice(0, 1);
      setFiles(next);
      onChange?.(next);
    },
    [files, maxSize, multiple, onChange],
  );

  const removeFile = (index: number) => {
    const next = files.filter((_, i) => i !== index);
    setFiles(next);
    onChange?.(next);
  };

  return (
    <div className={cn("space-y-3", className)}>
      <div
        onDragOver={(e) => {
          e.preventDefault();
          setDragOver(true);
        }}
        onDragLeave={() => setDragOver(false)}
        onDrop={(e) => {
          e.preventDefault();
          setDragOver(false);
          handleFiles(e.dataTransfer.files);
        }}
        onClick={() => inputRef.current?.click()}
        className={cn(
          "duration-normal flex cursor-pointer flex-col items-center justify-center gap-2 rounded-lg border-2 border-dashed p-8 transition-colors",
          dragOver
            ? "border-primary bg-primary/5"
            : "border-border hover:border-primary/50 hover:bg-surface-hover",
        )}
      >
        <Upload className="text-muted-foreground h-8 w-8" />
        <p className="text-body text-muted-foreground">
          Drag & drop or{" "}
          <span className="text-primary font-medium">browse</span>
        </p>
        <p className="text-caption text-muted-foreground">
          Max {(maxSize / 1024 / 1024).toFixed(0)}MB per file
        </p>
        <input
          ref={inputRef}
          type="file"
          accept={accept}
          multiple={multiple}
          onChange={(e) => handleFiles(e.target.files)}
          className="hidden"
        />
      </div>

      {files.length > 0 && (
        <ul className="space-y-2">
          {files.map((file, i) => (
            <li
              key={i}
              className="border-border bg-surface flex items-center gap-3 rounded-md border p-3"
            >
              <File className="text-muted-foreground h-4 w-4 shrink-0" />
              <span className="text-body flex-1 truncate">{file.name}</span>
              <span className="text-caption text-muted-foreground">
                {(file.size / 1024).toFixed(1)}KB
              </span>
              <button
                onClick={() => removeFile(i)}
                className="text-muted-foreground hover:text-danger transition-colors"
              >
                <X className="h-4 w-4" />
              </button>
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}
