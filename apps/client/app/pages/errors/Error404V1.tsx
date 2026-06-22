import { Search, ArrowLeft, Home } from "lucide-react";
import { useNavigate } from "react-router";
import { NexusButton } from "@casbin/ui";

export default function Error404V1() {
  const navigate = useNavigate();
  return (
    <div className="bg-background flex min-h-screen">
      <div className="bg-muted/50 relative hidden w-1/2 items-center justify-center overflow-hidden lg:flex">
        <div className="absolute inset-0">
          {[...Array(5)].map((_, i) => (
            <div
              key={i}
              className="border-primary/10 absolute rounded-full border"
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
        <div className="relative z-10 space-y-6 px-12 text-center">
          <div className="bg-primary/10 mx-auto flex h-24 w-24 items-center justify-center rounded-full">
            <Search className="text-primary h-12 w-12" />
          </div>
          <h2 className="text-foreground text-3xl font-bold">Lost in Space</h2>
          <p className="text-muted-foreground max-w-md text-lg">
            The page you're looking for has been moved, deleted, or never
            existed.
          </p>
        </div>
      </div>

      <div className="flex flex-1 items-center justify-center p-8">
        <div className="w-full max-w-md space-y-8 text-center">
          <div className="space-y-2">
            <p className="text-primary/20 text-8xl font-black">404</p>
            <h1 className="text-foreground text-3xl font-bold">
              Page Not Found
            </h1>
            <p className="text-muted-foreground">
              We couldn't find the page you were looking for. Check the URL or
              navigate back.
            </p>
          </div>
          <div className="flex flex-col justify-center gap-3 sm:flex-row">
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
