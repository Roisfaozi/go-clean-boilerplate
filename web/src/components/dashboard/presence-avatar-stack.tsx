"use client";

import { usePresenceStore } from "~/stores/use-presence-store";
import { useAuthStore } from "~/stores/use-auth-store";
import { Avatar, AvatarFallback, AvatarImage } from "~/components/ui/avatar";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "~/components/ui/tooltip";
import { cn } from "~/lib/utils";

export function PresenceAvatarStack({ className }: { className?: string }) {
  const { onlineUsers } = usePresenceStore();
  const { user: currentUser } = useAuthStore();

  const otherUsers = onlineUsers.filter((u) => u.user_id !== currentUser?.id);

  const displayUsers = otherUsers.slice(0, 5);
  const remainingCount = Math.max(0, otherUsers.length - 5);

  if (otherUsers.length === 0) return null;

  return (
    <div
      className={cn(
        "flex items-center -space-x-2 overflow-hidden px-2",
        className
      )}
    >
      <TooltipProvider>
        {displayUsers.map((user) => (
          <Tooltip key={user.user_id}>
            <TooltipTrigger asChild>
              <div className="ring-background relative inline-block rounded-full ring-2 transition-transform hover:z-10 hover:scale-110">
                <Avatar className="h-8 w-8">
                  <AvatarImage src={user.avatar_url} alt={user.name} />
                  <AvatarFallback className="bg-primary/10 text-[10px] font-bold">
                    {(user.name || "U")[0].toUpperCase()}
                  </AvatarFallback>
                </Avatar>
                <span className="absolute right-0 bottom-0 block h-2 w-2 rounded-full bg-emerald-500 ring-1 ring-white" />
              </div>
            </TooltipTrigger>
            <TooltipContent>
              <div className="flex flex-col gap-1">
                <p className="text-xs font-semibold">
                  {user.name || "Anonymous"}
                </p>
                <p className="text-muted-foreground text-[10px] uppercase">
                  {user.role}
                </p>
              </div>
            </TooltipContent>
          </Tooltip>
        ))}

        {remainingCount > 0 && (
          <Tooltip>
            <TooltipTrigger asChild>
              <div className="bg-muted ring-background flex h-8 w-8 items-center justify-center rounded-full text-[10px] font-bold ring-2 hover:z-10">
                +{remainingCount}
              </div>
            </TooltipTrigger>
            <TooltipContent>
              <p className="text-xs">{remainingCount} more online</p>
            </TooltipContent>
          </Tooltip>
        )}
      </TooltipProvider>
    </div>
  );
}
