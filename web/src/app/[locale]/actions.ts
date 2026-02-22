"use server";

import { revalidatePath } from "next/cache";
import { redirect } from "next/navigation";
import { authActionClient } from "~/lib/client/safe-action";
import { logoutAction } from "../actions/auth";

export const logout = authActionClient
  .metadata({ actionName: "logout" })
  .action(async () => {
    await logoutAction();
    revalidatePath("/");
    return redirect("/login");
  });
