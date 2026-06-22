import { ServerCrash, Home, RefreshCw, Zap } from "lucide-react";
import { useNavigate } from "react-router";
import { NexusButton } from "@casbin/ui";

export default function Error500V3() {
  const navigate = useNavigate();
  return (
    <div className="bg-sidebar relative flex min-h-screen items-center justify-center overflow-hidden p-8">
      <div
        className="absolute inset-0 opacity-[0.03]"
        style={{
          backgroundImage:
            "linear-gradient(hsl(var(--destructive)) 1px, transparent 1px), linear-gradient(90deg, hsl(var(--destructive)) 1px, transparent 1px)",
          backgroundSize: "60px 60px",
        }}
      />
      <div className="bg-destructive/5 absolute top-1/3 left-1/2 h-[500px] w-[500px] -translate-x-1/2 -translate-y-1/2 rounded-full blur-[120px]" />

      <div className="relative z-10 w-full max-w-2xl space-y-10 text-center">
        <div className="flex items-center justify-center gap-4">
          <Zap className="text-destructive/40 h-8 w-8" />
          <span className="text-destructive/10 text-9xl leading-none font-black">
            500
          </span>
          <Zap className="text-destructive/40 h-8 w-8" />
        </div>
        <div className="space-y-4">
          <h1 className="text-sidebar-foreground text-4xl font-bold">
            Internal Server Error
          </h1>
          <p className="text-muted-foreground mx-auto max-w-md text-lg">
            Something broke on our side. Our engineering team has been alerted.
          </p>
        </div>
        <div className="bg-sidebar-accent/50 border-border mx-auto max-w-sm rounded-xl border p-6">
          <ServerCrash className="text-destructive mx-auto mb-3 h-8 w-8" />
          <p className="text-muted-foreground text-sm">
            Try refreshing the page. If the problem persists, please contact
            support.
          </p>
        </div>
        <div className="flex justify-center gap-3">
          <NexusButton
            variant="primary"
            size="lg"
            onClick={() => window.location.reload()}
          >
            <RefreshCw className="h-4 w-4" />
            Retry
          </NexusButton>
          <NexusButton
            variant="outline"
            size="lg"
            onClick={() => navigate("/")}
          >
            <Home className="h-4 w-4" />
            Dashboard
          </NexusButton>
        </div>
      </div>
    </div>
  );
}
