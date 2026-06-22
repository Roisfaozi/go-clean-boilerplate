import { Lock, ArrowLeft, LogIn, Shield } from "lucide-react";
import { useNavigate } from "react-router";
import { NexusButton } from "@casbin/ui";

export default function Error401V3() {
  const navigate = useNavigate();
  return (
    <div className="bg-sidebar relative flex min-h-screen items-center justify-center overflow-hidden p-8">
      {/* Grid pattern */}
      <div
        className="absolute inset-0 opacity-[0.03]"
        style={{
          backgroundImage:
            "linear-gradient(hsl(var(--primary)) 1px, transparent 1px), linear-gradient(90deg, hsl(var(--primary)) 1px, transparent 1px)",
          backgroundSize: "60px 60px",
        }}
      />
      {/* Glow */}
      <div className="bg-primary/5 absolute top-1/3 left-1/2 h-[500px] w-[500px] -translate-x-1/2 -translate-y-1/2 rounded-full blur-[120px]" />

      <div className="relative z-10 w-full max-w-2xl space-y-10 text-center">
        <div className="flex items-center justify-center gap-4">
          <Shield className="text-primary/40 h-8 w-8" />
          <span className="text-primary/10 text-9xl leading-none font-black">
            401
          </span>
          <Shield className="text-primary/40 h-8 w-8" />
        </div>
        <div className="space-y-4">
          <h1 className="text-sidebar-foreground text-4xl font-bold">
            Authentication Required
          </h1>
          <p className="text-muted-foreground mx-auto max-w-md text-lg">
            Access denied. Valid credentials are required to view this resource.
          </p>
        </div>
        <div className="bg-sidebar-accent/50 border-border mx-auto max-w-sm rounded-xl border p-6">
          <Lock className="text-primary mx-auto mb-3 h-8 w-8" />
          <p className="text-muted-foreground text-sm">
            This area requires authentication. Please sign in with your
            credentials.
          </p>
        </div>
        <div className="flex justify-center gap-3">
          <NexusButton
            variant="primary"
            size="lg"
            onClick={() => navigate("/login")}
          >
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
