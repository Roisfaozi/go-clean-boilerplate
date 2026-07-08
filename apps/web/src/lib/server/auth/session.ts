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

export const createSession = async (_token: string, _userId: string) => ({
	id: "new-session-id",
	expiresAt: new Date(Date.now() + 1000 * 60 * 60 * 24 * 30),
});
export const generateSessionToken = () => "placeholder-token";
export const invalidateSession = async (_sessionId: string) => undefined;
export const invalidateAllSessions = async (_userId: string) => undefined;
export const verifyVerificationCode = async (
	_user: { id: string; email: string },
	_code: string,
) => true;
export const generateEmailVerificationCode = async (
	_userId: string,
	_email: string,
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
