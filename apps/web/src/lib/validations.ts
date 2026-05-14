import { z } from "zod";

export const settingsSchema = z.object({
  name: z.string().min(2).max(50),
  email: z.string().email(),
  picture: z.string().optional(),
});

export type SettingsValues = z.infer<typeof settingsSchema>;
