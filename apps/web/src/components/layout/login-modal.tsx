"use client";

import { usePathname, useRouter } from "next/navigation";
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@casbin/ui";
import AuthForm from "../auth/login-form";

export default function LoginModal() {
	const router = useRouter();
	const pathname = usePathname();

	const IsOpen = pathname.includes("/login");
	return (
		<Dialog open={IsOpen} onOpenChange={() => router.back()}>
			<DialogContent className="w-full max-w-[400px] rounded-md">
				<DialogHeader>
					<DialogTitle className="font-semibold tracking-tight transition-colors">
						Login
					</DialogTitle>
				</DialogHeader>
				<AuthForm />
			</DialogContent>
		</Dialog>
	);
}
