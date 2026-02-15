"use client";

import { usePresenceStore } from "~/stores/use-presence-store";
import { useAuthStore } from "~/stores/use-auth-store";
import { Avatar, AvatarFallback, AvatarImage } from "~/components/ui/avatar";
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "~/components/ui/tooltip";
import { cn } from "~/lib/utils";

export function PresenceAvatarStack({ className }: { className?: string }) {
  const { onlineUsers } = usePresenceStore();
  const { user: currentUser } = useAuthStore();

  const otherUsers = onlineUsers.filter(u => u.user_id !== currentUser?.id);

  const displayUsers = otherUsers.slice(0, 5);
  const remainingCount = Math.max(0, otherUsers.length - 5);

  if (otherUsers.length === 0) return null;

  return (
    <div className={cn("flex items-center -space-x-2 overflow-hidden px-2", className)}>
      <TooltipProvider>
        {displayUsers.map((user) => (
          <Tooltip key={user.user_id}>
            <TooltipTrigger asChild>
              <div className="relative inline-block ring-2 ring-background rounded-full transition-transform hover:z-10 hover:scale-110">
                <Avatar className="h-8 w-8">
                  <AvatarImage src={user.avatar_url} alt={user.name} />
                  <AvatarFallback className="bg-primary/10 text-[10px] font-bold">
                    {(user.name || "U")[0].toUpperCase()}
                  </AvatarFallback>
                </Avatar>
                <span className="absolute bottom-0 right-0 block h-2 w-2 rounded-full bg-emerald-500 ring-1 ring-white" />
              </div>
            </TooltipTrigger>
            <TooltipContent>
              <div className="flex flex-col gap-1">
                <p className="text-xs font-semibold">{user.name || "Anonymous"}</p>
                <p className="text-[10px] text-muted-foreground uppercase">{user.role}</p>
              </div>
            </TooltipContent>
          </Tooltip>
        ))}

        {remainingCount > 0 && (
          <Tooltip>
            <TooltipTrigger asChild>
              <div className="flex h-8 w-8 items-center justify-center rounded-full bg-muted text-[10px] font-bold ring-2 ring-background hover:z-10">
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
