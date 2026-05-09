import { ServerCrash, ArrowLeft, Home, RefreshCw } from "lucide-react";
import { useNavigate } from "react-router";
import {  NexusButton  } from "@casbin/ui";

export default function Error500V1() {
  const navigate = useNavigate();
  return (
    <div className="min-h-screen flex bg-background">
      <div className="hidden lg:flex w-1/2 bg-destructive/5 items-center justify-center relative overflow-hidden">
        <div className="absolute inset-0">
          {[...Array(5)].map((_, i) => (
            <div
              key={i}
              className="absolute rounded-full border border-destructive/10"
              style={{
                width: `${200 + i * 120}px`,
                height: `${200 + i * 120}px`,
                top: "50%",
                left: "50%",
                transform: "translate(-50%, -50%)",
              }}
            />
          ))}
        </div>
        <div className="relative z-10 text-center space-y-6 px-12">
          <div className="h-24 w-24 rounded-full bg-destructive/10 flex items-center justify-center mx-auto">
            <ServerCrash className="h-12 w-12 text-destructive" />
          </div>
          <h2 className="text-3xl font-bold text-foreground">Server Error</h2>
          <p className="text-muted-foreground text-lg max-w-md">
            Something went wrong on our end. Our team has been notified and is working on a fix.
          </p>
        </div>
      </div>

      <div className="flex-1 flex items-center justify-center p-8">
        <div className="max-w-md w-full text-center space-y-8">
          <div className="space-y-2">
            <p className="text-8xl font-black text-destructive/20">500</p>
            <h1 className="text-3xl font-bold text-foreground">Internal Server Error</h1>
            <p className="text-muted-foreground">
              We're experiencing technical difficulties. Please try again in a few moments.
            </p>
          </div>
          <div className="flex flex-col sm:flex-row gap-3 justify-center">
            <NexusButton variant="primary" onClick={() => window.location.reload()}>
              <RefreshCw className="h-4 w-4" />
              Try Again
            </NexusButton>
            <NexusButton variant="outline" onClick={() => navigate("/")}>
              <Home className="h-4 w-4" />
              Go Home
            </NexusButton>
          </div>
        </div>
      </div>
    </div>
  );
}
