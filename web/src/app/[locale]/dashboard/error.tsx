"use client";

import { useEffect } from "react";
import { Button } from "~/components/ui/button";

export default function Error({
  error,
  reset,
}: {
  error: Error;
  reset: () => void;
}) {
  useEffect(() => {
    // Log the error to an error reporting service
    console.log(error);
  }, [error]);

  return (
    <div className="flex h-[calc(100vh-160px)] w-full flex-col items-center justify-center gap-y-4">
      <h2 className="text-destructive text-5xl font-bold">
        Oops, Something Went Wrong!
      </h2>
      <Button onClick={() => reset()}>Try Again</Button>
    </div>
  );
}
