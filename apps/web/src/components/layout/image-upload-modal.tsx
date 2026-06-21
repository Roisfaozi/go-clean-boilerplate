"use client";

import { useState } from "react";
import {
  Dialog,
  DialogContent,
  DialogTrigger,
  DialogHeader,
  DialogTitle,
} from "~/components/ui/dialog";
import { Button } from "~/components/ui/button";
import { Input } from "~/components/ui/input";
import { Label } from "~/components/ui/label";
import { Camera, Loader2, UploadCloud } from "lucide-react";

interface ImageUploadModalProps {
  onImageChange: (url: string) => void;
}

export default function ImageUploadModal({
  onImageChange,
}: ImageUploadModalProps) {
  const [open, setOpen] = useState(false);
  const [preview, setPreview] = useState<string | null>(null);
  const [isUploading, setIsUploading] = useState(false);

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (file) {
      const reader = new FileReader();
      reader.onload = (event) => {
        setPreview(event.target?.result as string);
      };
      reader.readAsDataURL(file);
    }
  };

  const handleSave = async () => {
    if (!preview) return;
    setIsUploading(true);

    try {
      // In a real app, you would upload to TUS/S3 here
      // const url = await uploadToCDN(preview);
      // For now, we use the base64 string directly

      // Artificial delay to simulate upload
      await new Promise((resolve) => setTimeout(resolve, 800));

      onImageChange(preview);
      setOpen(false);
      setPreview(null);
    } catch (error) {
      console.error(error);
    } finally {
      setIsUploading(false);
    }
  };

  return (
    <Dialog
      open={open}
      onOpenChange={(val) => {
        setOpen(val);
        if (!val) setPreview(null);
      }}
    >
      <DialogTrigger asChild>
        <Button
          type="button"
          variant="outline"
          size="icon"
          className="bg-background/80 text-muted-foreground hover:bg-background hover:text-foreground absolute right-0 bottom-0 h-8 w-8 rounded-full shadow-sm backdrop-blur-sm"
        >
          <Camera className="h-4 w-4" />
          <span className="sr-only">Change Picture</span>
        </Button>
      </DialogTrigger>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>Upload Profile Picture</DialogTitle>
        </DialogHeader>
        <div className="flex flex-col items-center justify-center gap-4 py-4">
          {preview ? (
            <div className="border-muted relative h-40 w-40 overflow-hidden rounded-full border-4">
              {/* Using standard img to avoid next/image domain config issues for base64/blobs */}
              {/* eslint-disable-next-line @next/next/no-img-element */}
              <img
                src={preview}
                alt="Preview"
                className="h-full w-full object-cover"
              />
            </div>
          ) : (
            <div className="border-muted bg-muted/30 text-muted-foreground flex h-40 w-40 flex-col items-center justify-center rounded-full border-4 border-dashed">
              <UploadCloud className="mb-2 h-8 w-8" />
              <span className="text-xs">No image selected</span>
            </div>
          )}

          <div className="w-full space-y-2">
            <Label htmlFor="picture-upload" className="sr-only">
              Choose file
            </Label>
            <Input
              id="picture-upload"
              type="file"
              accept="image/*"
              onChange={handleFileChange}
              disabled={isUploading}
            />
          </div>

          <div className="flex w-full justify-end gap-2 pt-4">
            <Button
              type="button"
              variant="outline"
              onClick={() => {
                setOpen(false);
                setPreview(null);
              }}
              disabled={isUploading}
            >
              Cancel
            </Button>
            <Button
              type="button"
              onClick={handleSave}
              disabled={!preview || isUploading}
            >
              {isUploading ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  Saving...
                </>
              ) : (
                "Save Image"
              )}
            </Button>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}
