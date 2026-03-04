"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import Link from "next/link";
import { useSearchParams } from "next/navigation";
import { useState } from "react";
import { useForm } from "react-hook-form";
import { z } from "zod";
import { loginAction } from "~/app/actions/auth";
import { Button } from "~/components/ui/button";
import { toast } from "~/hooks/use-toast";
import { accessApi } from "~/lib/api/access";
import { loginSchema } from "~/lib/api/auth";
import { cn } from "~/lib/utils";
import { useAuthStore } from "~/stores/use-auth-store";
import { usePermissionStore } from "~/stores/use-permission-store";
import Icons from "../shared/icons";
import { Input } from "../ui/input";
import { Label } from "../ui/label";

type FormData = z.infer<typeof loginSchema>;

export default function AuthForm() {
  const [isLoading, setIsLoading] = useState(false);
  const [isGithubLoading, setIsGithubLoading] = useState(false);
  const setPermissions = usePermissionStore((state) => state.setPermissions);
  const setUser = useAuthStore((state) => state.setUser);
  const searchParams = useSearchParams();
  const returnTo = searchParams.get("returnTo") || "/dashboard";

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<FormData>({
    resolver: zodResolver(loginSchema),
  });

  async function onSubmit(data: FormData) {
    setIsLoading(true);

    try {
      const result = await loginAction(data);

      if (!result?.data?.success) {
        throw new Error(result?.data?.message || "Login failed");
      }

      const { user } = result.data;

      // Store user in auth store for UI
      if (user) {
        setUser(user);
      }

      // Fetch permissions for the role
      try {
        const permsResp = await accessApi.getPermissionsForRole(user.role);
        if (permsResp.data) {
          setPermissions(permsResp.data);
        }
      } catch (err) {
        console.error("Failed to fetch permissions", err);
      }

      toast({
        title: "Successfully logged in!",
        description: `Welcome back, ${user.name || "User"}`,
      });

      // Redirect to target page or dashboard
      window.location.href = decodeURIComponent(returnTo);
    } catch (error) {
      const errorMessage =
        error instanceof Error ? error.message : "Invalid username or password";
      toast({
        title: "Login failed",
        description: errorMessage,
        variant: "destructive",
      });
    } finally {
      setIsLoading(false);
    }
  }

  return (
    <div className={cn("grid gap-6")}>
      <form onSubmit={handleSubmit(onSubmit)}>
        <div className="grid gap-4">
          <div className="grid gap-2">
            <Label htmlFor="username">Username</Label>
            <div className="relative">
              <Input
                id="username"
                placeholder="username"
                type="text"
                autoCapitalize="none"
                autoComplete="username"
                autoCorrect="off"
                disabled={isLoading || isGithubLoading}
                className="pr-10"
                {...register("username")}
              />
              <div className="text-muted-foreground/50 hover:text-primary absolute top-1/2 right-3 -translate-y-1/2 cursor-help transition-colors">
                <Icons.user className="h-4 w-4" />
              </div>
            </div>
            {errors?.username && (
              <p className="text-destructive px-1 text-xs">
                {errors.username.message}
              </p>
            )}
          </div>

          <div className="grid gap-2">
            <div className="flex items-center justify-between">
              <Label htmlFor="password">Password</Label>
              <Link
                href="/forgot-password"
                className="text-primary text-xs font-medium underline-offset-4 hover:underline"
              >
                Forgot password?
              </Link>
            </div>
            <Input
              id="password"
              placeholder="••••••••"
              type="password"
              autoComplete="current-password"
              disabled={isLoading || isGithubLoading}
              {...register("password")}
            />
            {errors?.password && (
              <p className="text-destructive px-1 text-xs">
                {errors.password.message}
              </p>
            )}
          </div>

          <Button
            type="submit"
            className="mt-2 w-full"
            disabled={isLoading || isGithubLoading}
          >
            {isLoading ? (
              <Icons.spinner className="mr-2 h-4 w-4 animate-spin" />
            ) : null}
            Sign In
          </Button>
        </div>
      </form>

      <div className="relative">
        <div className="absolute inset-0 flex items-center">
          <span className="w-full border-t" />
        </div>
        <div className="relative flex justify-center text-xs uppercase">
          <span className="bg-background text-muted-foreground px-2">
            Or continue with
          </span>
        </div>
      </div>

      <div className="grid gap-2">
        <Button
          variant="outline"
          type="button"
          disabled={isLoading || isGithubLoading}
          onClick={() => {
            setIsGithubLoading(true);
            toast({
              title: "Coming soon",
              description: "Social login is not yet supported in this demo",
            });
            setIsGithubLoading(false);
          }}
        >
          {isGithubLoading ? (
            <Icons.spinner className="mr-2 h-4 w-4 animate-spin" />
          ) : (
            <Icons.gitHub className="mr-2 h-4 w-4" />
          )}{" "}
          GitHub
        </Button>
      </div>
    </div>
  );
}
