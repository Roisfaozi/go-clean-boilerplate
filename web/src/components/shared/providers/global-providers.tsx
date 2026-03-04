"use client";

import { ReactNode } from "react";
import ThemeProvider from "./theme-provider";
import { AuthProvider } from "./auth-provider";
import { DensityProvider } from "./density-provider";
import { WebSocketProvider } from "./websocket-provider";
import { Toaster } from "~/components/ui/sonner";
import { Toaster as LegacyToaster } from "~/components/ui/toaster";
import { AiChatWidget } from "~/components/dashboard/ai-chat/ai-chat-widget";

/**
 * GlobalProviders - Unified wrapper for all core app providers.
 * Follows the composition pattern to keep layout.tsx clean.
 */
export function GlobalProviders({ children }: { children: ReactNode }) {
  return (
    <ThemeProvider attribute="class" defaultTheme="system" enableSystem>
      <DensityProvider>
        <AuthProvider>
          <WebSocketProvider>
            {children}
            <AiChatWidget />
            <Toaster position="top-right" richColors closeButton />
            <LegacyToaster />
          </WebSocketProvider>
        </AuthProvider>
      </DensityProvider>
    </ThemeProvider>
  );
}
