import { Lock, ArrowLeft, LogIn } from "lucide-react";
import { useNavigate } from "react-router";
import {  NexusButton  } from "@casbin/ui";

export default function Error401V1() {
  const navigate = useNavigate();
  return (
    <div className="min-h-screen flex bg-background">
      {/* Left branding panel */}
      <div className="hidden lg:flex w-1/2 bg-primary/5 items-center justify-center relative overflow-hidden">
        <div className="absolute inset-0">
          {[...Array(5)].map((_, i) => (
            <div
              key={i}
              className="absolute rounded-full border border-primary/10"
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
          <div className="h-24 w-24 rounded-full bg-primary/10 flex items-center justify-center mx-auto">
            <Lock className="h-12 w-12 text-primary" />
          </div>
          <h2 className="text-3xl font-bold text-foreground">Authentication Required</h2>
          <p className="text-muted-foreground text-lg max-w-md">
            You need to sign in to access this resource. Your session may have expired.
          </p>
        </div>
      </div>

      {/* Right content */}
      <div className="flex-1 flex items-center justify-center p-8">
        <div className="max-w-md w-full text-center space-y-8">
          <div className="space-y-2">
            <p className="text-8xl font-black text-primary/20">401</p>
            <h1 className="text-3xl font-bold text-foreground">Unauthorized</h1>
            <p className="text-muted-foreground">
              Please sign in to continue. If you believe this is an error, contact support.
            </p>
          </div>
          <div className="flex flex-col sm:flex-row gap-3 justify-center">
            <NexusButton variant="primary" onClick={() => navigate("/login")}>
              <LogIn className="h-4 w-4" />
              Sign In
            </NexusButton>
            <NexusButton variant="outline" onClick={() => navigate(-1)}>
              <ArrowLeft className="h-4 w-4" />
              Go Back
            </NexusButton>
          </div>
        </div>
      </div>
    </div>
  );
}
