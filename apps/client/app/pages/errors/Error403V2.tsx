import { ShieldOff, ArrowLeft, Home } from "lucide-react";
import { useNavigate } from "react-router";
import {  NexusButton  } from "@casbin/ui";

export default function Error403V2() {
  const navigate = useNavigate();
  return (
    <div className="min-h-screen flex items-center justify-center p-8 relative overflow-hidden">
      <div className="absolute inset-0 bg-gradient-to-br from-danger/20 via-warning/10 to-primary/10 animate-pulse" />
      <div className="absolute top-1/3 -left-20 w-72 h-72 bg-danger/10 rounded-full blur-3xl" />
      <div className="absolute bottom-1/4 -right-20 w-72 h-72 bg-warning/10 rounded-full blur-3xl" />

      <div className="relative z-10 bg-card/80 backdrop-blur-xl border border-border rounded-2xl p-10 max-w-lg w-full text-center shadow-xl space-y-8">
        <div className="h-20 w-20 rounded-2xl bg-danger/10 flex items-center justify-center mx-auto rotate-12">
          <ShieldOff className="h-10 w-10 text-danger -rotate-12" />
        </div>
        <div className="space-y-2">
          <p className="text-sm font-semibold text-danger uppercase tracking-widest">Error 403</p>
          <h1 className="text-4xl font-bold text-foreground">Access Forbidden</h1>
          <p className="text-muted-foreground leading-relaxed">
            Your account doesn't have the required permissions. Contact an admin to request access.
          </p>
        </div>
        <div className="flex flex-col gap-3">
          <NexusButton variant="primary" size="lg" className="w-full" onClick={() => navigate("/")}>
            <Home className="h-4 w-4" />
            Return to Dashboard
          </NexusButton>
          <NexusButton variant="ghost" onClick={() => navigate(-1)}>
            <ArrowLeft className="h-4 w-4" />
            Go Back
          </NexusButton>
        </div>
      </div>
    </div>
  );
}
