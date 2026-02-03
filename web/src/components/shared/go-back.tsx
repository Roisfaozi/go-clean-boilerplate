"use client";

import { useRouter } from "next/navigation";
import { Button } from "~/components/ui/button";
import { ArrowLeft } from "lucide-react";

export default function GoBack() {
  const router = useRouter();
  return (
    <Button variant="ghost" className="gap-2 pl-0 mb-4" onClick={() => router.back()}>
      <ArrowLeft className="h-4 w-4" />
      Back
    </Button>
  );
}
