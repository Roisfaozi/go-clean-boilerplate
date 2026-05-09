import { Lock, ArrowLeft, LogIn } from "lucide-react";
import { useNavigate } from "react-router";
import {  NexusButton  } from "@casbin/ui";

export default function Error401V2() {
  const navigate = useNavigate();
  return (
    <div className="min-h-screen flex items-center justify-center p-8 relative overflow-hidden">
      {/* Animated gradient background */}
      <div className="absolute inset-0 bg-gradient-to-br from-primary/20 via-accent/10 to-secondary/20 animate-pulse" />
      <div className="absolute top-1/4 -left-20 w-72 h-72 bg-primary/10 rounded-full blur-3xl" />
      <div className="absolute bottom-1/4 -right-20 w-72 h-72 bg-accent/10 rounded-full blur-3xl" />

      <div className="relative z-10 bg-card/80 backdrop-blur-xl border border-border rounded-2xl p-10 max-w-lg w-full text-center shadow-xl space-y-8">
        <div className="h-20 w-20 rounded-2xl bg-warning/10 flex items-center justify-center mx-auto rotate-12">
          <Lock className="h-10 w-10 text-warning -rotate-12" />
        </div>
        <div className="space-y-2">
          <p className="text-sm font-semibold text-warning uppercase tracking-widest">Error 401</p>
          <h1 className="text-4xl font-bold text-foreground">Not Authenticated</h1>
          <p className="text-muted-foreground leading-relaxed">
            Your session has expired or you haven't signed in yet. Please authenticate to continue.
          </p>
        </div>
        <div className="flex flex-col gap-3">
          <NexusButton variant="primary" size="lg" className="w-full" onClick={() => navigate("/login")}>
            <LogIn className="h-4 w-4" />
            Sign In to Continue
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
