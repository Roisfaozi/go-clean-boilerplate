"use server";

import { revalidatePath } from "next/cache";
import { usersApi } from "~/lib/api/users";

export const updateUser = async (id: string, payload: any) => {
  // Use our Go API instead of Prisma
  await usersApi.updateMe({
    name: payload.name,
    // Add other fields as needed by backend
  });

  revalidatePath("/dashboard/settings");
};

export async function removeUserOldImageFromCDN(
  newImageUrl: string,
  currentImageUrl: string
) {
  // Placeholder logic if we are not using Uploadthing with Go yet
  console.log("Removing old image:", currentImageUrl);
}

export async function removeNewImageFromCDN(image: string) {
  console.log("Removing new image:", image);
}