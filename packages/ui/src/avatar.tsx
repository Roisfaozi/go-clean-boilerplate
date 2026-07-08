"use client";

import * as React from "react";
import { Avatar } from "@base-ui/react/avatar";

import { cn } from "./lib/utils";

const AvatarRoot = React.forwardRef<
  React.ComponentRef<typeof Avatar.Root>,
  React.ComponentProps<typeof Avatar.Root>
>(({ className, ...props }, ref) => (
  <Avatar.Root
    ref={ref}
    className={cn(
      "relative flex h-10 w-10 shrink-0 overflow-hidden rounded-full",
      className,
    )}
    {...props}
  />
));
AvatarRoot.displayName = "Avatar";

const AvatarImage = React.forwardRef<
  React.ComponentRef<typeof Avatar.Image>,
  React.ComponentProps<typeof Avatar.Image>
>(({ className, ...props }, ref) => (
  <Avatar.Image
    ref={ref}
    className={cn("aspect-square h-full w-full", className)}
    {...props}
  />
));
AvatarImage.displayName = "AvatarImage";

const AvatarFallback = React.forwardRef<
  React.ComponentRef<typeof Avatar.Fallback>,
  React.ComponentProps<typeof Avatar.Fallback>
>(({ className, ...props }, ref) => (
  <Avatar.Fallback
    ref={ref}
    className={cn(
      "bg-muted flex h-full w-full items-center justify-center rounded-full",
      className,
    )}
    {...props}
  />
));
AvatarFallback.displayName = "AvatarFallback";

export { AvatarRoot as Avatar, AvatarImage, AvatarFallback };
