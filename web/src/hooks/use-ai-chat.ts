"use client";

import { useState, useCallback, useRef } from "react";

export interface Message {
  role: "user" | "assistant";
  content: string;
}

export function useAiChat() {
  const [messages, setMessages] = useState<Message[]>([]);
  const [isTyping, setIsTyping] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const abortControllerRef = useRef<AbortController | null>(null);

  const sendMessage = useCallback(
    async (content: string) => {
      if (!content.trim()) return;

      // 1. Add user message
      const userMessage: Message = { role: "user", content };
      setMessages((prev) => [...prev, userMessage]);
      setIsTyping(true);
      setError(null);

      // Create a new assistant message placeholder
      const assistantMessage: Message = { role: "assistant", content: "" };
      setMessages((prev) => [...prev, assistantMessage]);

      // Setup for streaming
      abortControllerRef.current = new AbortController();

      try {
        const response = await fetch("/api/ai/chat", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ messages: [...messages, userMessage] }),
          signal: abortControllerRef.current.signal,
        });

        if (!response.ok) throw new Error("Failed to connect to AI assistant");

        const reader = response.body?.getReader();
        const textDecoder = new TextDecoder();

        if (!reader) throw new Error("No reader available");

        let accumulatedContent = "";

        while (true) {
          const { done, value } = await reader.read();
          if (done) break;

          const chunk = textDecoder.decode(value);
          const lines = chunk.split("\n");

          for (const line of lines) {
            if (line.startsWith("data: ")) {
              try {
                const data = JSON.parse(line.slice(6));
                if (data.text) {
                  accumulatedContent += data.text;
                  // Update the last message (the assistant's message)
                  setMessages((prev) => {
                    const newMessages = [...prev];
                    newMessages[newMessages.length - 1] = {
                      role: "assistant",
                      content: accumulatedContent,
                    };
                    return newMessages;
                  });
                }
              } catch (e) {
                // Ignore parse errors for incomplete chunks
              }
            }
          }
        }
      } catch (err: any) {
        if (err.name === "AbortError") {
          console.log("Chat aborted");
        } else {
          setError(err.message || "An error occurred");
        }
      } finally {
        setIsTyping(false);
        abortControllerRef.current = null;
      }
    },
    [messages]
  );

  const stopGeneration = useCallback(() => {
    if (abortControllerRef.current) {
      abortControllerRef.current.abort();
    }
  }, []);

  const clearChat = useCallback(() => {
    setMessages([]);
    setError(null);
  }, []);

  return {
    messages,
    sendMessage,
    isTyping,
    error,
    stopGeneration,
    clearChat,
  };
}
