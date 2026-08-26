import { redirect } from "next/navigation";

export default async function AcceptInvite({
	params,
	searchParams,
}: {
	params: Promise<{ locale: string }>;
	searchParams: Promise<{ token?: string | string[] }>;
}) {
	const { locale } = await params;
	const { token: rawToken } = await searchParams;
	const token = Array.isArray(rawToken) ? rawToken[0] : rawToken;

	if (!token) {
		redirect(`/${locale}/login`);
	}

	redirect(`/${locale}/invite/${encodeURIComponent(token)}`);
}
