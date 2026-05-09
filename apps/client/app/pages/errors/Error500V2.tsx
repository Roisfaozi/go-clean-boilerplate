import { ServerCrash, ArrowLeft, Home, RefreshCw } from "lucide-react";
import { useNavigate } from "react-router";
import {  NexusButton  } from "@casbin/ui";

export default function Error500V2() {
  const navigate = useNavigate();
  return (
    <div className="min-h-screen flex items-center justify-center p-8 relative overflow-hidden">
      <div className="absolute inset-0 bg-gradient-to-br from-destructive/20 via-danger/10 to-warning/10 animate-pulse" />
      <div className="absolute top-1/4 -left-20 w-72 h-72 bg-destructive/10 rounded-full blur-3xl" />
      <div className="absolute bottom-1/3 -right-20 w-72 h-72 bg-danger/10 rounded-full blur-3xl" />

      <div className="relative z-10 bg-card/80 backdrop-blur-xl border border-border rounded-2xl p-10 max-w-lg w-full text-center shadow-xl space-y-8">
        <div className="h-20 w-20 rounded-2xl bg-destructive/10 flex items-center justify-center mx-auto rotate-12">
          <ServerCrash className="h-10 w-10 text-destructive -rotate-12" />
        </div>
        <div className="space-y-2">
          <p className="text-sm font-semibold text-destructive uppercase tracking-widest">Error 500</p>
          <h1 className="text-4xl font-bold text-foreground">Server Error</h1>
          <p className="text-muted-foreground leading-relaxed">
            Our servers are having a moment. We've been notified and are working to fix this.
          </p>
        </div>
        <div className="flex flex-col gap-3">
          <NexusButton variant="primary" size="lg" className="w-full" onClick={() => window.location.reload()}>
            <RefreshCw className="h-4 w-4" />
            Retry
          </NexusButton>
          <NexusButton variant="ghost" onClick={() => navigate("/")}>
            <Home className="h-4 w-4" />
            Go to Dashboard
          </NexusButton>
        </div>
      </div>
    </div>
  );
}
