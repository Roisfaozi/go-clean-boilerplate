"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useState } from "react";
import { useForm } from "react-hook-form";
import { z } from "zod";
import { Button } from "~/components/ui/button";
import { toast } from "~/hooks/use-toast";
import { cn } from "~/lib/utils";
import Icons from "../shared/icons";
import { Input } from "../ui/input";
import { Label } from "../ui/label";
import { authApi } from "~/lib/api/auth";

const registerSchema = z
  .object({
    name: z.string().min(2, "Name must be at least 2 characters."),
    username: z.string().min(3, "Username must be at least 3 characters."),
    email: z.string().email("Please enter a valid email address."),
    password: z.string().min(8, "Password must be at least 8 characters."),
    confirmPassword: z.string(),
  })
  .refine((data) => data.password === data.confirmPassword, {
    message: "Passwords don't match",
    path: ["confirmPassword"],
  });

type FormData = z.infer<typeof registerSchema>;

export default function RegisterForm() {
  const [isLoading, setIsLoading] = useState(false);

  const {
    register,
    handleSubmit,
    watch,
    formState: { errors },
  } = useForm<FormData>({
    resolver: zodResolver(registerSchema),
  });

  const password = watch("password", "");

  const getStrength = (pass: string) => {
    let strength = 0;
    if (pass.length >= 8) strength += 1;
    if (/[0-9]/.test(pass)) strength += 1;
    if (/[^A-Za-z0-9]/.test(pass)) strength += 1;
    return strength;
  };

  const strength = getStrength(password);

  async function onSubmit(data: FormData) {
    setIsLoading(true);

    try {
      await authApi.register({
        name: data.name,
        username: data.username,
        email: data.email,
        password: data.password,
      });

      toast({
        title: "Account created!",
        description: "Please check your email to verify your account.",
      });

      // Redirect to login
      window.location.href = "/login";
    } catch (error) {
      const errorMessage =
        error instanceof Error ? error.message : "Failed to create account";

      toast({
        title: "Registration failed",
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
            <Label htmlFor="name">Full Name</Label>
            <Input
              id="name"
              placeholder="John Doe"
              type="text"
              autoCapitalize="words"
              autoComplete="name"
              autoCorrect="off"
              disabled={isLoading}
              {...register("name")}
            />
            {errors?.name && (
              <p className="text-destructive px-1 text-xs">
                {errors.name.message}
              </p>
            )}
          </div>

          <div className="grid gap-2">
            <Label htmlFor="username">Username</Label>
            <Input
              id="username"
              placeholder="johndoe"
              type="text"
              autoCapitalize="none"
              autoComplete="username"
              autoCorrect="off"
              disabled={isLoading}
              {...register("username")}
            />
            {errors?.username && (
              <p className="text-destructive px-1 text-xs">
                {errors.username.message}
              </p>
            )}
          </div>

          <div className="grid gap-2">
            <Label htmlFor="email">Email</Label>
            <Input
              id="email"
              placeholder="name@example.com"
              type="email"
              autoCapitalize="none"
              autoComplete="email"
              autoCorrect="off"
              disabled={isLoading}
              {...register("email")}
            />
            {errors?.email && (
              <p className="text-destructive px-1 text-xs">
                {errors.email.message}
              </p>
            )}
          </div>

          <div className="grid gap-2">
            <Label htmlFor="password">Password</Label>
            <Input
              id="password"
              placeholder="••••••••"
              type="password"
              autoComplete="new-password"
              disabled={isLoading}
              {...register("password")}
            />
            {errors?.password && (
              <p className="text-destructive px-1 text-xs">
                {errors.password.message}
              </p>
            )}
            <div className="bg-muted mt-1 flex h-1 w-full gap-1 overflow-hidden rounded-full">
              <div
                className={cn(
                  "h-full transition-all duration-500",
                  strength === 1 && "bg-destructive w-1/3",
                  strength === 2 && "bg-warning w-2/3",
                  strength === 3 && "bg-success w-full",
                  strength === 0 && "w-0"
                )}
              />
            </div>
            <p className="text-muted-foreground text-[10px]">
              Minimum 8 characters with numbers and symbols
            </p>
          </div>

          <div className="grid gap-2">
            <Label htmlFor="confirmPassword">Confirm Password</Label>
            <Input
              id="confirmPassword"
              placeholder="••••••••"
              type="password"
              autoComplete="new-password"
              disabled={isLoading}
              {...register("confirmPassword")}
            />
            {errors?.confirmPassword && (
              <p className="text-destructive px-1 text-xs">
                {errors.confirmPassword.message}
              </p>
            )}
          </div>

          <Button type="submit" className="mt-2 w-full" disabled={isLoading}>
            {isLoading ? (
              <Icons.spinner className="mr-2 h-4 w-4 animate-spin" />
            ) : null}
            Create Account
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
          disabled={isLoading}
          onClick={() => {
            toast({
              title: "Coming soon",
              description: "Social login is not yet supported in this demo",
            });
          }}
        >
          <Icons.gitHub className="mr-2 h-4 w-4" />
          GitHub
        </Button>
      </div>
    </div>
  );
}
