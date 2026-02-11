"use client";

import { useState, use } from "react";
import { useRouter } from "next/navigation";
import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";
import { z } from "zod";
import { Button } from "~/components/ui/button";
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "~/components/ui/form";
import { Input } from "~/components/ui/input";
import { toast } from "sonner";
import { Icon } from "~/components/shared/icon";
import { organizationsApi } from "~/lib/api/organizations";
import Link from "next/link";
import Icons from "~/components/shared/icons";

const inviteSchema = z
  .object({
    name: z.string().min(2, "Name must be at least 2 characters.").optional(),
    password: z
      .string()
      .min(8, "Password must be at least 8 characters.")
      .optional(),
    confirmPassword: z.string().optional(),
  })
  .refine(
    (data) => {
      if (data.password && data.password !== data.confirmPassword) {
        return false;
      }
      return true;
    },
    {
      message: "Passwords do not match",
      path: ["confirmPassword"],
    }
  );

type InviteFormValues = z.infer<typeof inviteSchema>;

interface Props {
  params: Promise<{ token: string; locale: string }>;
}

export default function InvitationPage({ params }: Props) {
  const { token } = use(params);
  const router = useRouter();
  const [isLoading, setIsLoading] = useState(false);

  const form = useForm<InviteFormValues>({
    resolver: zodResolver(inviteSchema),
    defaultValues: {
      name: "",
      password: "",
      confirmPassword: "",
    },
  });

  async function onSubmit(data: InviteFormValues) {
    setIsLoading(true);
    try {
      await organizationsApi.acceptInvitation({
        token,
        name: data.name,
        password: data.password,
      });
      toast.success("Invitation accepted!", {
        description: "You can now log in to your account.",
      });
      router.push("/login");
    } catch (error: any) {
      toast.error(error.message || "Failed to accept invitation");
    } finally {
      setIsLoading(false);
    }
  }

  return (
    <div className="flex min-h-screen w-full flex-col lg:flex-row">
      {/* Left Panel: Functional Zone */}
      <div className="flex flex-1 flex-col justify-center px-6 py-12 md:px-12 lg:px-24 xl:px-32">
        <div className="mx-auto w-full max-w-sm lg:mx-0">
          <Link href="/" className="mb-10 flex items-center gap-2">
            <Icons.logo className="text-primary h-10 w-10" />
            <span className="text-2xl font-bold tracking-tighter">NexusOS</span>
          </Link>

          <div className="mb-8">
            <h1 className="text-3xl font-bold tracking-tight">
              Accept Invitation
            </h1>
            <p className="text-muted-foreground mt-2 text-sm">
              You&apos;ve been invited to join an organization. Complete your
              details below.
            </p>
          </div>

          <Form {...form}>
            <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
              <FormField
                control={form.control}
                name="name"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Full Name</FormLabel>
                    <FormControl>
                      <Input placeholder="John Doe" {...field} />
                    </FormControl>
                    <FormDescription>
                      Optional: Update your display name.
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <div className="space-y-4 border-t pt-4">
                <p className="text-muted-foreground text-[10px] font-bold tracking-widest uppercase">
                  Set Security Credentials
                </p>
                <FormField
                  control={form.control}
                  name="password"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>New Password</FormLabel>
                      <FormControl>
                        <Input
                          type="password"
                          placeholder="••••••••"
                          {...field}
                        />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
                <FormField
                  control={form.control}
                  name="confirmPassword"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>Confirm Password</FormLabel>
                      <FormControl>
                        <Input
                          type="password"
                          placeholder="••••••••"
                          {...field}
                        />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
              </div>

              <Button
                type="submit"
                className="mt-6 w-full"
                disabled={isLoading}
              >
                {isLoading ? (
                  <Icon name="Loader" className="mr-2 h-4 w-4 animate-spin" />
                ) : (
                  <Icon name="Check" className="mr-2 h-4 w-4" />
                )}
                Join Organization
              </Button>
            </form>
          </Form>

          <p className="text-muted-foreground mt-8 text-center text-sm lg:text-left">
            Already have an account?{" "}
            <Link
              href="/login"
              className="text-primary font-medium underline-offset-4 hover:underline"
            >
              Log in instead
            </Link>
          </p>
        </div>
      </div>

      {/* Right Panel: Visual/Branding Zone */}
      <div className="bg-primary relative hidden w-full flex-1 items-center justify-center overflow-hidden lg:flex">
        {/* Background Decorative Elements */}
        <div className="absolute inset-0 bg-linear-to-br from-indigo-600 to-violet-700 opacity-90" />
        <div className="absolute inset-0 bg-[url('https://www.transparenttextures.com/patterns/cubes.png')] opacity-20" />

        <div className="relative z-10 p-12 text-center text-white">
          <div className="mx-auto max-w-md">
            <div className="mx-auto mb-8 flex h-16 w-16 items-center justify-center rounded-2xl bg-white/20 shadow-xl backdrop-blur-md">
              <Icons.logo className="h-10 w-10" />
            </div>
            <h2 className="mb-6 text-4xl leading-tight font-bold tracking-tight">
              Collaborate Better, Scale Faster.
            </h2>
            <p className="mb-10 text-lg text-indigo-100">
              Welcome to the team! NexusOS handles the complexity of access
              control so you can focus on building what matters.
            </p>

            <div className="flex items-center justify-center gap-4 border-y border-white/10 py-6">
              <div className="flex -space-x-2">
                {[1, 2, 3, 4].map((i) => (
                  <div
                    key={i}
                    className="flex h-8 w-8 items-center justify-center rounded-full border-2 border-indigo-600 bg-indigo-200 text-[10px] font-bold text-indigo-800"
                  >
                    U{i}
                  </div>
                ))}
              </div>
              <p className="text-sm font-medium text-indigo-100">
                Join 10+ other team members
              </p>
            </div>
          </div>
        </div>

        {/* Animated Orbs */}
        <div className="absolute -right-24 -bottom-24 h-96 w-96 animate-pulse rounded-full bg-violet-500/20 blur-3xl" />
        <div className="absolute -top-24 -left-24 h-96 w-96 animate-pulse rounded-full bg-indigo-400/20 blur-3xl" />
      </div>
    </div>
  );
}
