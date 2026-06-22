"use client";

import { useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
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
import { Avatar, AvatarFallback, AvatarImage } from "~/components/ui/avatar";
import { toast } from "sonner";
import { Icon } from "~/components/shared/icon";
import { usersApi } from "~/lib/api/users";

const profileSchema = z.object({
  name: z.string().min(2, "Name must be at least 2 characters."),
  email: z.string().email(),
  username: z.string().min(3),
});

type ProfileFormValues = z.infer<typeof profileSchema>;

interface ProfileFormProps {
  user: any;
}

export function ProfileForm({ user }: ProfileFormProps) {
  const [isLoading, setIsLoading] = useState(false);
  const [isUploading, setIsUploading] = useState(false);

  const form = useForm<ProfileFormValues>({
    resolver: zodResolver(profileSchema),
    defaultValues: {
      name: user?.name || "",
      email: user?.email || "",
      username: user?.username || "",
    },
  });

  async function onSubmit(data: ProfileFormValues) {
    setIsLoading(true);
    try {
      await usersApi.updateMe({ name: data.name });
      toast.success("Profile updated successfully");
    } catch (error: any) {
      toast.error(error.message || "Failed to update profile");
    } finally {
      setIsLoading(false);
    }
  }

  const handleAvatarChange = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;

    setIsUploading(true);
    try {
      await usersApi.uploadAvatar(file);
      toast.success("Avatar updated successfully");
      // Reload page or update local state to see new avatar
      window.location.reload();
    } catch (error: any) {
      toast.error(error.message || "Failed to upload avatar");
    } finally {
      setIsUploading(false);
    }
  };

  return (
    <Form {...form}>
      <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-8">
        {/* Avatar Section */}
        <div className="flex items-center gap-x-6">
          <Avatar className="border-muted h-20 w-20 border-2">
            <AvatarImage src={user?.avatar_url} />
            <AvatarFallback className="bg-primary/5 text-lg">
              {user?.username?.slice(0, 2).toUpperCase() || "JD"}
            </AvatarFallback>
          </Avatar>
          <div className="relative">
            <Input
              type="file"
              className="hidden"
              id="avatar-upload"
              accept="image/*"
              onChange={handleAvatarChange}
              disabled={isUploading}
            />
            <label htmlFor="avatar-upload">
              <Button
                variant="outline"
                type="button"
                size="sm"
                className="cursor-pointer"
                asChild
              >
                <span>
                  {isUploading ? (
                    <Icon name="Loader" className="mr-2 h-4 w-4 animate-spin" />
                  ) : (
                    <Icon name="Upload" className="mr-2 h-4 w-4" />
                  )}
                  Change Avatar
                </span>
              </Button>
            </label>
          </div>
        </div>

        <FormField
          control={form.control}
          name="username"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Username</FormLabel>
              <FormControl>
                <Input {...field} disabled />
              </FormControl>
              <FormDescription>
                This is your public display name. It can be your real name or a
                pseudonym.
              </FormDescription>
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name="name"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Full Name</FormLabel>
              <FormControl>
                <Input placeholder="John Doe" {...field} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name="email"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Email</FormLabel>
              <FormControl>
                <Input placeholder="john@example.com" {...field} disabled />
              </FormControl>
              <FormDescription>
                Email cannot be changed directly. Contact support if needed.
              </FormDescription>
              <FormMessage />
            </FormItem>
          )}
        />

        <Button type="submit" disabled={isLoading}>
          {isLoading && (
            <Icon name="Loader" className="mr-2 h-4 w-4 animate-spin" />
          )}
          Update Profile
        </Button>
      </form>
    </Form>
  );
}
