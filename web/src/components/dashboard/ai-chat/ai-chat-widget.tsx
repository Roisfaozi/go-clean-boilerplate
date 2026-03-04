"use client";

import * as React from "react";
import { useState, useRef, useEffect } from "react";
import { useAiChat } from "~/hooks/use-ai-chat";
import { Button } from "~/components/ui/button";
import { Icon } from "~/components/shared/icon";
import { ScrollArea } from "~/components/ui/scroll-area";
import { cn } from "~/lib/utils";
import { motion, AnimatePresence } from "framer-motion";
import { Input } from "~/components/ui/input";

export function AiChatWidget() {
  const [isOpen, setIsOpen] = useState(false);
  const [isDocked, setIsDocked] = useState(false);
  const { messages, sendMessage, isTyping } = useAiChat();
  const [inputValue, setInputValue] = useState("");
  const scrollRef = useRef<HTMLDivElement>(null);

  // Auto-scroll to bottom
  useEffect(() => {
    if (scrollRef.current) {
      scrollRef.current.scrollIntoView({ behavior: "smooth" });
    }
  }, [messages]);

  const handleSend = async () => {
    if (!inputValue.trim() || isTyping) return;
    const content = inputValue;
    setInputValue("");
    await sendMessage(content);
  };

  return (
    <>
      {/* FAB Button when closed */}
      <AnimatePresence>
        {!isOpen && (
          <motion.div
            initial={{ scale: 0, opacity: 0 }}
            animate={{ scale: 1, opacity: 1 }}
            exit={{ scale: 0, opacity: 0 }}
            className="fixed right-6 bottom-6 z-50"
          >
            <Button
              size="icon"
              className="h-14 w-14 rounded-full shadow-2xl transition-transform hover:scale-110"
              onClick={() => setIsOpen(true)}
            >
              <Icon name="Bot" className="h-6 w-6" />
            </Button>
          </motion.div>
        )}
      </AnimatePresence>

      {/* Chat Window */}
      <AnimatePresence>
        {isOpen && (
          <motion.div
            initial={{ opacity: 0, y: 20, scale: 0.95 }}
            animate={{ opacity: 1, y: 0, scale: 1 }}
            exit={{ opacity: 0, y: 20, scale: 0.95 }}
            className={cn(
              "bg-card fixed z-50 flex flex-col overflow-hidden border shadow-2xl transition-all duration-300",
              isDocked
                ? "top-0 right-0 h-screen w-80 rounded-none border-l"
                : "right-6 bottom-6 h-[500px] w-96 rounded-2xl"
            )}
          >
            {/* Header */}
            <div className="bg-muted/30 flex items-center justify-between border-b px-4 py-3">
              <div className="flex items-center gap-2">
                <div className="bg-primary/10 flex h-8 w-8 items-center justify-center rounded-lg">
                  <Icon name="Bot" className="text-primary h-5 w-5" />
                </div>
                <div>
                  <h3 className="text-sm font-semibold">NexusAI</h3>
                  <div className="flex items-center gap-1">
                    <div
                      className={cn(
                        "h-1.5 w-1.5 rounded-full",
                        isTyping
                          ? "animate-pulse bg-emerald-500"
                          : "bg-muted-foreground/30"
                      )}
                    />
                    <span className="text-muted-foreground text-[10px] font-bold tracking-wider uppercase">
                      {isTyping ? "Thinking..." : "Idle"}
                    </span>
                  </div>
                </div>
              </div>
              <div className="flex items-center gap-1">
                <Button
                  variant="ghost"
                  size="icon"
                  className="h-8 w-8"
                  onClick={() => setIsDocked(!isDocked)}
                >
                  <Icon
                    name={isDocked ? "PanelRightOpen" : "PanelRightClose"}
                    className="h-4 w-4"
                  />
                </Button>
                <Button
                  variant="ghost"
                  size="icon"
                  className="h-8 w-8"
                  onClick={() => setIsOpen(false)}
                >
                  <Icon name="X" className="h-4 w-4" />
                </Button>
              </div>
            </div>

            {/* Chat Area */}
            <ScrollArea className="flex-1 p-4">
              {messages.length === 0 ? (
                <div className="flex h-full flex-col items-center justify-center py-12 text-center opacity-50">
                  <Icon name="MessagesSquare" className="mb-2 h-8 w-8" />
                  <p className="text-sm">
                    How can I help you with your dashboard today?
                  </p>
                </div>
              ) : (
                <div className="space-y-4">
                  {messages.map((msg, i) => (
                    <div
                      key={i}
                      className={cn(
                        "flex",
                        msg.role === "user" ? "justify-end" : "justify-start"
                      )}
                    >
                      <div
                        className={cn(
                          "max-w-[85%] rounded-2xl px-4 py-2.5 text-sm shadow-sm",
                          msg.role === "user"
                            ? "bg-primary text-primary-foreground rounded-tr-none"
                            : "bg-muted rounded-tl-none border"
                        )}
                      >
                        {msg.content || (
                          <div className="flex gap-1 py-1">
                            <div className="h-1.5 w-1.5 animate-bounce rounded-full bg-current" />
                            <div className="h-1.5 w-1.5 animate-bounce rounded-full bg-current [animation-delay:0.2s]" />
                            <div className="h-1.5 w-1.5 animate-bounce rounded-full bg-current [animation-delay:0.4s]" />
                          </div>
                        )}
                      </div>
                    </div>
                  ))}
                  <div ref={scrollRef} />
                </div>
              )}
            </ScrollArea>

            {/* Input Area */}
            <div className="bg-muted/10 border-t p-4">
              <form
                onSubmit={(e) => {
                  e.preventDefault();
                  handleSend();
                }}
                className="relative flex items-center"
              >
                <Input
                  value={inputValue}
                  onChange={(e) => setInputValue(e.target.value)}
                  placeholder="Ask anything..."
                  className="rounded-xl pr-10"
                  disabled={isTyping}
                />
                <Button
                  type="submit"
                  size="icon"
                  variant="ghost"
                  className="hover:bg-primary/10 hover:text-primary absolute right-1 h-8 w-8"
                  disabled={!inputValue.trim() || isTyping}
                >
                  <Icon name="SendHorizontal" className="h-4 w-4" />
                </Button>
              </form>
              <div className="no-scrollbar mt-2 flex gap-2 overflow-x-auto pb-1">
                <QuickChip
                  label="Analyze Page"
                  icon="BarChart"
                  onClick={() => setInputValue("Analyze the data on this page")}
                />
                <QuickChip
                  label="Audit Help"
                  icon="FileText"
                  onClick={() =>
                    setInputValue("Explain how to read audit logs")
                  }
                />
              </div>
            </div>
          </motion.div>
        )}
      </AnimatePresence>
    </>
  );
}

function QuickChip({
  label,
  icon,
  onClick,
}: {
  label: string;
  icon: string;
  onClick: () => void;
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      className="bg-background text-muted-foreground hover:bg-muted hover:text-foreground flex items-center gap-1.5 rounded-full border px-2.5 py-1 text-[10px] font-bold tracking-wider whitespace-nowrap uppercase transition-colors"
    >
      <Icon name={icon as any} className="h-3 w-3" />
      {label}
    </button>
  );
}
