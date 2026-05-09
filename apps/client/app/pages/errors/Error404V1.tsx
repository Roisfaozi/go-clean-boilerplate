import { Search, ArrowLeft, Home } from "lucide-react";
import { useNavigate } from "react-router";
import {  NexusButton  } from "@casbin/ui";

export default function Error404V1() {
  const navigate = useNavigate();
  return (
    <div className="min-h-screen flex bg-background">
      <div className="hidden lg:flex w-1/2 bg-muted/50 items-center justify-center relative overflow-hidden">
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
            <Search className="h-12 w-12 text-primary" />
          </div>
          <h2 className="text-3xl font-bold text-foreground">Lost in Space</h2>
          <p className="text-muted-foreground text-lg max-w-md">
            The page you're looking for has been moved, deleted, or never existed.
          </p>
        </div>
      </div>

      <div className="flex-1 flex items-center justify-center p-8">
        <div className="max-w-md w-full text-center space-y-8">
          <div className="space-y-2">
            <p className="text-8xl font-black text-primary/20">404</p>
            <h1 className="text-3xl font-bold text-foreground">Page Not Found</h1>
            <p className="text-muted-foreground">
              We couldn't find the page you were looking for. Check the URL or navigate back.
            </p>
          </div>
          <div className="flex flex-col sm:flex-row gap-3 justify-center">
            <NexusButton variant="primary" onClick={() => navigate("/")}>
              <Home className="h-4 w-4" />
              Go Home
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
