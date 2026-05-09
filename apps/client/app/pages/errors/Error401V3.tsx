import { Lock, ArrowLeft, LogIn, Shield } from "lucide-react";
import { useNavigate } from "react-router";
import {  NexusButton  } from "@casbin/ui";

export default function Error401V3() {
  const navigate = useNavigate();
  return (
    <div className="min-h-screen bg-sidebar flex items-center justify-center p-8 relative overflow-hidden">
      {/* Grid pattern */}
      <div
        className="absolute inset-0 opacity-[0.03]"
        style={{
          backgroundImage: "linear-gradient(hsl(var(--primary)) 1px, transparent 1px), linear-gradient(90deg, hsl(var(--primary)) 1px, transparent 1px)",
          backgroundSize: "60px 60px",
        }}
      />
      {/* Glow */}
      <div className="absolute top-1/3 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[500px] h-[500px] bg-primary/5 rounded-full blur-[120px]" />

      <div className="relative z-10 max-w-2xl w-full text-center space-y-10">
        <div className="flex items-center justify-center gap-4">
          <Shield className="h-8 w-8 text-primary/40" />
          <span className="text-9xl font-black text-primary/10 leading-none">401</span>
          <Shield className="h-8 w-8 text-primary/40" />
        </div>
        <div className="space-y-4">
          <h1 className="text-4xl font-bold text-sidebar-foreground">Authentication Required</h1>
          <p className="text-muted-foreground text-lg max-w-md mx-auto">
            Access denied. Valid credentials are required to view this resource.
          </p>
        </div>
        <div className="p-6 bg-sidebar-accent/50 border border-border rounded-xl max-w-sm mx-auto">
          <Lock className="h-8 w-8 text-primary mx-auto mb-3" />
          <p className="text-sm text-muted-foreground">This area requires authentication. Please sign in with your credentials.</p>
        </div>
        <div className="flex gap-3 justify-center">
          <NexusButton variant="primary" size="lg" onClick={() => navigate("/login")}>
            <LogIn className="h-4 w-4" />
            Sign In
          </NexusButton>
          <NexusButton variant="outline" size="lg" onClick={() => navigate(-1)}>
            <ArrowLeft className="h-4 w-4" />
            Go Back
          </NexusButton>
        </div>
      </div>
    </div>
  );
}
