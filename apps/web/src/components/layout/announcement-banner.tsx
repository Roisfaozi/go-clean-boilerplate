"use client";

import { X } from "lucide-react";
import Link from "next/link";
import { useState } from "react";

export function AnnouncementBanner() {
  const [isVisible, setIsVisible] = useState(true);

  const handleDismiss = () => {
    setIsVisible(false);
  };

  if (!isVisible) return null;

  return (
    <div className="bg-primary fixed top-0 right-0 left-0 z-50 border-b">
      <div className="container mx-auto px-4 py-2">
        <div className="flex items-center justify-between">
          <div className="text-primary-foreground flex-1 text-center text-sm font-medium">
            <span className="mr-2">🚀</span>
            The Author has released a timesaver tool for vibe coders,
            developers, founders, AI users -{" "}
            <Link
              href="https://voicetypr.com"
              target="_blank"
              rel="noopener noreferrer"
              className="font-semibold underline-offset-4 transition-all hover:underline"
            >
              check it out now!
            </Link>
          </div>
          <button
            onClick={handleDismiss}
            className="hover:bg-primary-foreground/20 ml-4 rounded-full p-1.5 transition-colors"
            aria-label="Dismiss announcement"
          >
            <X className="text-primary-foreground h-4 w-4" />
          </button>
        </div>
      </div>
    </div>
  );
}
