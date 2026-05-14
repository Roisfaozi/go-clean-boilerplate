"use client";

import { icons, LucideProps } from "lucide-react";
import { cn } from "~/lib/utils";

interface IconProps extends LucideProps {
  name: keyof typeof icons;
  size?: "sm" | "md" | "lg" | "xl";
}

/**
 * Atomic Icon component that automatically adapts to the global "Fluid Density" system.
 * Sizes and stroke widths are controlled via CSS variables defined in globals.css.
 */
export const Icon = ({ name, size = "md", className, ...props }: IconProps) => {
  const LucideIcon = icons[name];

  if (!LucideIcon) {
    return null;
  }

  // Size mapping using dynamic CSS variables
  // icon-size changes from 20px (comfort) to 16px (compact)
  const sizeClasses = {
    sm: "size-[calc(var(--icon-size)*0.8)]", // ~16px / 13px
    md: "size-[var(--icon-size)]", // 20px / 16px
    lg: "size-[calc(var(--icon-size)*1.2)]", // 24px / 20px
    xl: "size-[calc(var(--icon-size)*1.6)]", // 32px / 24px
  };

  return (
    <LucideIcon
      className={cn(
        "shrink-0 transition-all",
        // Stroke width adapts: 2px (comfort) vs 1.5px (compact)
        "[stroke-width:var(--icon-stroke,2px)] [data-density=compact]:[stroke-width:1.5px]",
        sizeClasses[size],
        className
      )}
      {...props}
    />
  );
};
