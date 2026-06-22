"use client";

import { useRouter } from "next/navigation";
import { Button } from "~/components/ui/button";
import { ArrowLeft } from "lucide-react";

export default function GoBack() {
  const router = useRouter();
  return (
    <Button
      variant="ghost"
      className="mb-4 gap-2 pl-0"
      onClick={() => router.back()}
    >
      <ArrowLeft className="h-4 w-4" />
      Back
    </Button>
  );
}
