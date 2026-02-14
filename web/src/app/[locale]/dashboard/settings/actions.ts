"use server";

import { revalidatePath } from "next/cache";
import { usersApi } from "~/lib/api/users";
import { authActionClient } from "~/lib/client/safe-action";
import { z } from "zod";

const updateUserSchema = z.object({
  name: z.string().min(1, "Name is required").optional(),
  username: z.string().min(3, "Username must be at least 3 characters").optional(),
});

export const updateUserAction = authActionClient
  .schema(updateUserSchema)
  .action(async ({ parsedInput }) => {
    // Use our Go API instead of Prisma
    const result = await usersApi.updateMe(parsedInput);
    revalidatePath("/dashboard/settings");
    return result;
  });

export async function removeUserOldImageFromCDN(
  newImageUrl: string,
  currentImageUrl: string
) {
  // Placeholder logic for now
  console.log("Removing old image:", currentImageUrl);
}

export async function removeNewImageFromCDN(image: string) {
  console.log("Removing new image:", image);
}
