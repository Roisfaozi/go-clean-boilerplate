import { Search, ArrowLeft, Home } from "lucide-react";
import { useNavigate } from "react-router";
import {  NexusButton  } from "@casbin/ui";

export default function Error404V2() {
  const navigate = useNavigate();
  return (
    <div className="min-h-screen flex items-center justify-center p-8 relative overflow-hidden">
      <div className="absolute inset-0 bg-gradient-to-br from-primary/20 via-secondary/10 to-accent/20 animate-pulse" />
      <div className="absolute top-1/4 -left-20 w-72 h-72 bg-primary/10 rounded-full blur-3xl" />
      <div className="absolute bottom-1/3 -right-20 w-72 h-72 bg-secondary/10 rounded-full blur-3xl" />

      <div className="relative z-10 bg-card/80 backdrop-blur-xl border border-border rounded-2xl p-10 max-w-lg w-full text-center shadow-xl space-y-8">
        <div className="h-20 w-20 rounded-2xl bg-primary/10 flex items-center justify-center mx-auto rotate-12">
          <Search className="h-10 w-10 text-primary -rotate-12" />
        </div>
        <div className="space-y-2">
          <p className="text-sm font-semibold text-primary uppercase tracking-widest">Error 404</p>
          <h1 className="text-4xl font-bold text-foreground">Page Not Found</h1>
          <p className="text-muted-foreground leading-relaxed">
            Looks like this page went on vacation. Let's get you back on track.
          </p>
        </div>
        <div className="flex flex-col gap-3">
          <NexusButton variant="primary" size="lg" className="w-full" onClick={() => navigate("/")}>
            <Home className="h-4 w-4" />
            Back to Dashboard
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
