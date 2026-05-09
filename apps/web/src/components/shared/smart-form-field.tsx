"use client";

import * as React from "react";
import { Label } from "~/components/ui/label";
import { Input } from "~/components/ui/input";
import { cn } from "~/lib/utils";
import { Icon } from "./icon";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "~/components/ui/tooltip";

interface SmartFormFieldProps extends React.ComponentProps<typeof Input> {
  label: string;
  error?: string;
  hint?: string;
  isAI?: boolean;
  onAiClick?: () => void;
  isLoadingAi?: boolean;
}

/**
 * Level 2 Molecule: Smart Form Field
 * Combines Label, Input, Error, and AI features.
 * Adapts layout based on global density: Vertical (Comfort) vs Horizontal (Compact).
 */
export const SmartFormField = React.forwardRef<
  HTMLInputElement,
  SmartFormFieldProps
>(
  (
    {
      label,
      error,
      hint,
      isAI,
      onAiClick,
      isLoadingAi,
      className,
      id,
      ...props
    },
    ref
  ) => {
    const generatedId = React.useId();
    const fieldId = id || generatedId;

    return (
      <div
        className={cn(
          "group flex flex-col gap-2",
          // Density adaptation: Horizontal layout in compact mode
          "[data-density=compact]:flex-row [data-density=compact]:items-center [data-density=compact]:justify-between [data-density=compact]:gap-4"
        )}
      >
        <div className="flex flex-col gap-1 [data-density=compact]:w-1/3">
          <Label
            htmlFor={fieldId}
            className={cn(
              "group-focus-within:text-primary text-sm font-medium transition-colors",
              "[data-density=compact]:text-muted-foreground [data-density=compact]:text-xs"
            )}
          >
            {label}
            {props.required && <span className="text-destructive ml-1">*</span>}
          </Label>
          {hint && !error && (
            <p className="text-muted-foreground text-xs [data-density=compact]:hidden">
              {hint}
            </p>
          )}
        </div>

        <div className="relative flex-1">
          <Input
            id={fieldId}
            ref={ref}
            className={cn(
              error && "border-destructive focus-visible:ring-destructive",
              isAI && "pr-10",
              className
            )}
            {...props}
          />

          {isAI && (
            <div className="absolute top-1/2 right-3 -translate-y-1/2">
              <TooltipProvider>
                <Tooltip>
                  <TooltipTrigger asChild>
                    <button
                      type="button"
                      onClick={onAiClick}
                      disabled={isLoadingAi || props.disabled}
                      className={cn(
                        "hover:bg-accent flex items-center justify-center rounded-sm transition-all",
                        isLoadingAi
                          ? "animate-pulse text-violet-500"
                          : "text-muted-foreground hover:text-violet-500"
                      )}
                    >
                      <Icon
                        name={(isLoadingAi ? "Loader2" : "Sparkles") as any}
                        size="sm"
                        className={cn(isLoadingAi && "animate-spin")}
                      />
                    </button>
                  </TooltipTrigger>
                  <TooltipContent>
                    <p>Auto-fill with AI</p>
                  </TooltipContent>
                </Tooltip>
              </TooltipProvider>
            </div>
          )}

          {error && (
            <p className="text-destructive animate-in fade-in slide-in-from-top-1 mt-1 text-xs">
              {error}
            </p>
          )}
        </div>
      </div>
    );
  }
);

SmartFormField.displayName = "SmartFormField";
