"use client";

import { useState, useEffect } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { Button } from "~/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "~/components/ui/dialog";
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "~/components/ui/form";
import { Input } from "~/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "~/components/ui/select";
import { toast } from "sonner";
import { authApi } from "~/lib/api/auth"; // For register (create)
import { usersApi } from "~/lib/api/users"; // For update
import { Icon } from "~/components/shared/icon";

const userSchema = z.object({
  name: z.string().min(2, "Name must be at least 2 characters."),
  username: z.string().min(3, "Username must be at least 3 characters."),
  email: z.string().email("Invalid email address."),
  password: z.string().optional(), // Optional for edit
  status: z.enum(["active", "suspended", "banned"]).optional(),
});

type UserFormValues = z.infer<typeof userSchema>;

interface UserDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  user?: any; // If present, edit mode
  onSuccess: () => void;
}

export function UserDialog({
  open,
  onOpenChange,
  user,
  onSuccess,
}: UserDialogProps) {
  const [isLoading, setIsLoading] = useState(false);
  const isEdit = !!user;

  const form = useForm<UserFormValues>({
    resolver: zodResolver(userSchema),
    defaultValues: {
      name: "",
      username: "",
      email: "",
      password: "",
      status: "active",
    },
  });

  // Reset form when user changes (edit mode) or dialog opens/closes
  useEffect(() => {
    if (open) {
      form.reset({
        name: user?.name || "",
        username: user?.username || "",
        email: user?.email || "",
        password: "",
        status: user?.status || "active",
      });
    }
  }, [user, open, form]);

  async function onSubmit(data: UserFormValues) {
    setIsLoading(true);
    try {
      if (isEdit) {
        // Update User
        // Note: The backend update endpoint might separate status and profile updates.
        // Based on users.ts: updateMe (profile) and updateStatus.
        // Admin update might be different or restricted.
        // Assuming we can update status via updateStatus and profile via a different endpoint if available.
        // usersApi.updateMe updates the CURRENT user. Admin updating OTHER users isn't explicitly in users.ts getAll/getById section yet.
        // Let's check swagger.json: PUT /users/{id} exists? No, only /users/me and /users/{id}/status PATCH.
        // Wait, swagger.json has PUT /users/me but DELETE /users/{id}.
        // It seems there is NO admin endpoint to update another user's profile details (name/email), only status.
        // Checking swagger.json again...
        // /users/{id} GET, DELETE.
        // /users/{id}/status PATCH.
        // So Admin can only update Status or Delete.
        // I will only allow Status update for edit mode then.

        if (user.status !== data.status && data.status) {
          await usersApi.updateStatus(
            user.id,
            data.status as "active" | "suspended" | "banned"
          );
        }

        toast.success("User updated successfully");
      } else {
        // Create User (Register)
        if (!data.password) {
          form.setError("password", {
            message: "Password is required for new users",
          });
          setIsLoading(false);
          return;
        }
        await authApi.register({
          name: data.name,
          username: data.username,
          email: data.email,
          password: data.password,
        });
        toast.success("User created successfully");
      }
      onSuccess();
      onOpenChange(false);
    } catch (error: any) {
      toast.error(error.message || "Something went wrong");
    } finally {
      setIsLoading(false);
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[425px]">
        <DialogHeader>
          <DialogTitle>{isEdit ? "Edit User" : "Add New User"}</DialogTitle>
          <DialogDescription>
            {isEdit
              ? "Update user status. Profile details can only be changed by the user."
              : "Create a new user account. They will receive an email to verify."}
          </DialogDescription>
        </DialogHeader>
        <Form {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
            <FormField
              control={form.control}
              name="name"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Full Name</FormLabel>
                  <FormControl>
                    <Input
                      placeholder="John Doe"
                      {...field}
                      disabled={isEdit}
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
            <FormField
              control={form.control}
              name="username"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Username</FormLabel>
                  <FormControl>
                    <Input placeholder="johndoe" {...field} disabled={isEdit} />
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
                    <Input
                      placeholder="john@example.com"
                      type="email"
                      {...field}
                      disabled={isEdit}
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            {!isEdit && (
              <FormField
                control={form.control}
                name="password"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Password</FormLabel>
                    <FormControl>
                      <Input
                        placeholder="••••••••"
                        type="password"
                        {...field}
                      />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
            )}

            {isEdit && (
              <FormField
                control={form.control}
                name="status"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Status</FormLabel>
                    <Select
                      onValueChange={field.onChange}
                      defaultValue={field.value}
                    >
                      <FormControl>
                        <SelectTrigger>
                          <SelectValue placeholder="Select status" />
                        </SelectTrigger>
                      </FormControl>
                      <SelectContent>
                        <SelectItem value="active">Active</SelectItem>
                        <SelectItem value="suspended">Suspended</SelectItem>
                        <SelectItem value="banned">Banned</SelectItem>
                      </SelectContent>
                    </Select>
                    <FormMessage />
                  </FormItem>
                )}
              />
            )}

            <DialogFooter>
              <Button type="submit" disabled={isLoading}>
                {isLoading && (
                  <Icon name="Loader" className="mr-2 h-4 w-4 animate-spin" />
                )}
                {isEdit ? "Save Changes" : "Create User"}
              </Button>
            </DialogFooter>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  );
}
