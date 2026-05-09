import { ServerCrash, Home, RefreshCw, Zap } from "lucide-react";
import { useNavigate } from "react-router";
import {  NexusButton  } from "@casbin/ui";

export default function Error500V3() {
  const navigate = useNavigate();
  return (
    <div className="min-h-screen bg-sidebar flex items-center justify-center p-8 relative overflow-hidden">
      <div
        className="absolute inset-0 opacity-[0.03]"
        style={{
          backgroundImage: "linear-gradient(hsl(var(--destructive)) 1px, transparent 1px), linear-gradient(90deg, hsl(var(--destructive)) 1px, transparent 1px)",
          backgroundSize: "60px 60px",
        }}
      />
      <div className="absolute top-1/3 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[500px] h-[500px] bg-destructive/5 rounded-full blur-[120px]" />

      <div className="relative z-10 max-w-2xl w-full text-center space-y-10">
        <div className="flex items-center justify-center gap-4">
          <Zap className="h-8 w-8 text-destructive/40" />
          <span className="text-9xl font-black text-destructive/10 leading-none">500</span>
          <Zap className="h-8 w-8 text-destructive/40" />
        </div>
        <div className="space-y-4">
          <h1 className="text-4xl font-bold text-sidebar-foreground">Internal Server Error</h1>
          <p className="text-muted-foreground text-lg max-w-md mx-auto">
            Something broke on our side. Our engineering team has been alerted.
          </p>
        </div>
        <div className="p-6 bg-sidebar-accent/50 border border-border rounded-xl max-w-sm mx-auto">
          <ServerCrash className="h-8 w-8 text-destructive mx-auto mb-3" />
          <p className="text-sm text-muted-foreground">Try refreshing the page. If the problem persists, please contact support.</p>
        </div>
        <div className="flex gap-3 justify-center">
          <NexusButton variant="primary" size="lg" onClick={() => window.location.reload()}>
            <RefreshCw className="h-4 w-4" />
            Retry
          </NexusButton>
          <NexusButton variant="outline" size="lg" onClick={() => navigate("/")}>
            <Home className="h-4 w-4" />
            Dashboard
          </NexusButton>
        </div>
      </div>
    </div>
  );
}
