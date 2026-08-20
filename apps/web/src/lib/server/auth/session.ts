import { cookies } from "next/headers";
import { redirect } from "next/navigation";

const BACKEND_URL =
	process.env.NEXT_PUBLIC_API_URL || "http://127.0.0.1:8080/api/v1";

type SessionUser = {
	id: string;
	name: string;
	email: string;
	username: string;
	role: string;
	avatar_url?: string;
	emailVerifiedAt: number | null;
};

type SessionResult = {
	session: { id: string } | null;
	user: SessionUser | null;
};

const getResponseData = async (response: Response) => {
	const payload = await response.json().catch(() => null);
	return payload?.data ?? payload;
};

export const getCurrentSession = async (): Promise<SessionResult> => {
	const cookieStore = await cookies();
	let accessToken = cookieStore.get("access_token")?.value;
	const cookieHeader = cookieStore
		.getAll()
		.map((cookie) => `${cookie.name}=${cookie.value}`)
		.join("; ");

	if (!accessToken) {
		return { session: null, user: null };
	}

	const getAuthUser = async (token: string) => {
		const response = await fetch(`${BACKEND_URL}/auth/me`, {
			headers: {
				Authorization: `Bearer ${token}`,
				Cookie: cookieHeader,
			},
			cache: "no-store",
		});
		if (!response.ok) return null;
		return (await getResponseData(response))?.user ?? null;
	};

	const authUser = await getAuthUser(accessToken);

	if (!authUser?.id) {
		return { session: null, user: null };
	}

	let profileUser: Partial<SessionUser> = {};
	const profileResponse = await fetch(`${BACKEND_URL}/users/me`, {
		headers: {
			Authorization: `Bearer ${accessToken}`,
			Cookie: cookieHeader,
		},
		cache: "no-store",
	});

	if (profileResponse.ok) {
		profileUser = await getResponseData(profileResponse);
	}

	return {
		session: { id: "cookie-session" },
		user: {
			id: authUser.id,
			name: profileUser.name ?? "",
			email: profileUser.email ?? "",
			username: profileUser.username ?? authUser.username ?? "",
			role: authUser.role ?? "user",
			avatar_url: profileUser.avatar_url,
			emailVerifiedAt: profileUser.emailVerifiedAt ?? null,
		},
	};
};

export const requireAuth = async (locale = "en") => {
	const { session, user } = await getCurrentSession();

	if (!session || !user) {
		redirect(
			`/${locale}/login?returnTo=${encodeURIComponent(`/${locale}/dashboard`)}`,
		);
	}

	return { session, user };
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
