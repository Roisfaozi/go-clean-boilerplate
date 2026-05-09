import { Search, ArrowLeft, Home, Compass } from "lucide-react";
import { useNavigate } from "react-router";
import {  NexusButton  } from "@casbin/ui";

export default function Error404V3() {
  const navigate = useNavigate();
  return (
    <div className="min-h-screen bg-sidebar flex items-center justify-center p-8 relative overflow-hidden">
      <div
        className="absolute inset-0 opacity-[0.03]"
        style={{
          backgroundImage: "linear-gradient(hsl(var(--primary)) 1px, transparent 1px), linear-gradient(90deg, hsl(var(--primary)) 1px, transparent 1px)",
          backgroundSize: "60px 60px",
        }}
      />
      <div className="absolute top-1/3 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[500px] h-[500px] bg-primary/5 rounded-full blur-[120px]" />

      <div className="relative z-10 max-w-2xl w-full text-center space-y-10">
        <div className="flex items-center justify-center gap-4">
          <Compass className="h-8 w-8 text-primary/40" />
          <span className="text-9xl font-black text-primary/10 leading-none">404</span>
          <Compass className="h-8 w-8 text-primary/40" />
        </div>
        <div className="space-y-4">
          <h1 className="text-4xl font-bold text-sidebar-foreground">Page Not Found</h1>
          <p className="text-muted-foreground text-lg max-w-md mx-auto">
            The destination you're trying to reach doesn't exist in our system.
          </p>
        </div>
        <div className="p-6 bg-sidebar-accent/50 border border-border rounded-xl max-w-sm mx-auto">
          <Search className="h-8 w-8 text-primary mx-auto mb-3" />
          <p className="text-sm text-muted-foreground">Double-check the URL or use the navigation to find what you need.</p>
        </div>
        <div className="flex gap-3 justify-center">
          <NexusButton variant="primary" size="lg" onClick={() => navigate("/")}>
            <Home className="h-4 w-4" />
            Dashboard
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
