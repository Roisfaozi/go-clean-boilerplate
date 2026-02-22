import { cookies } from "next/headers";

export const getCurrentSession = async () => {
  const cookieStore = await cookies();
  const accessToken = cookieStore.get("access_token")?.value;
  const refreshToken = cookieStore.get("refresh_token")?.value;

  if (!accessToken && !refreshToken) {
    return { session: null, user: null };
  }

  return {
    session: { id: "cookie-session" },
    user: {
      id: "current-user",
      email: "",
      name: "",
      role: "user",
      emailVerifiedAt: null,
    },
  };
};

export const createSession = async (token: string, userId: string) => ({
  id: "new-session-id",
  expiresAt: new Date(Date.now() + 1000 * 60 * 60 * 24 * 30),
});
export const generateSessionToken = () => "placeholder-token";
export const invalidateSession = async (sessionId: string) => {};
export const invalidateAllSessions = async (userId: string) => {};
export const verifyVerificationCode = async (
  user: { id: string; email: string },
  code: string
) => true;
export const generateEmailVerificationCode = async (
  userId: string,
  email: string
) => "123456";

export const authMiddleware = async ({ next }: { next: any }) => {
  const { session, user } = await getCurrentSession();
  return next({
    ctx: {
      sessionId: session?.id ?? "",
      user,
    },
  });
};
