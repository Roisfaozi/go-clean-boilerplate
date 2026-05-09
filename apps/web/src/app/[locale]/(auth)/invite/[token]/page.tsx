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
import { AuthLayoutShell } from "~/components/auth/auth-layout-shell";

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
    <AuthLayoutShell
      title="Accept Invitation"
      description="You've been invited to join an organization. Complete your details below."
      brandingTitle="Collaborate Better, Scale Faster."
      brandingDescription="Welcome to the team! NexusOS handles the complexity of access control so you can focus on building what matters."
      footer={
        <p className="text-muted-foreground text-center text-sm lg:text-left">
          Already have an account?{" "}
          <Link
            href="/login"
            className="text-primary font-medium underline-offset-4 hover:underline"
          >
            Log in instead
          </Link>
        </p>
      }
    >
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
                    <Input type="password" placeholder="••••••••" {...field} />
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
                    <Input type="password" placeholder="••••••••" {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
          </div>

          <Button type="submit" className="mt-6 w-full" disabled={isLoading}>
            {isLoading ? (
              <Icon name="Loader" className="mr-2 h-4 w-4 animate-spin" />
            ) : (
              <Icon name="Check" className="mr-2 h-4 w-4" />
            )}
            Join Organization
          </Button>
        </form>
      </Form>
    </AuthLayoutShell>
  );
}
