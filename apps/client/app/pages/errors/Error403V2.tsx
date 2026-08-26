import { ShieldOff, ArrowLeft, Home } from "lucide-react";
import { useNavigate } from "react-router";
import { NexusButton } from "@casbin/ui";

export default function Error403V2() {
	const navigate = useNavigate();
	return (
		<div className="relative flex min-h-screen items-center justify-center overflow-hidden p-8">
			<div className="from-danger/20 via-warning/10 to-primary/10 absolute inset-0 animate-pulse bg-gradient-to-br" />
			<div className="bg-danger/10 absolute top-1/3 -left-20 h-72 w-72 rounded-full blur-3xl" />
			<div className="bg-warning/10 absolute -right-20 bottom-1/4 h-72 w-72 rounded-full blur-3xl" />

			<div className="bg-card/80 border-border relative z-10 w-full max-w-lg space-y-8 rounded-2xl border p-10 text-center shadow-xl backdrop-blur-xl">
				<div className="bg-danger/10 mx-auto flex h-20 w-20 rotate-12 items-center justify-center rounded-2xl">
					<ShieldOff className="text-danger h-10 w-10 -rotate-12" />
				</div>
				<div className="space-y-2">
					<p className="text-danger text-sm font-semibold tracking-widest uppercase">
						Error 403
					</p>
					<h1 className="text-foreground text-4xl font-bold">
						Access Forbidden
					</h1>
					<p className="text-muted-foreground leading-relaxed">
						Your account doesn't have the required permissions. Contact an admin
						to request access.
					</p>
				</div>
				<div className="flex flex-col gap-3">
					<NexusButton
						variant="primary"
						size="lg"
						className="w-full"
						onClick={() => navigate("/")}
					>
						<Home className="h-4 w-4" />
						Return to Dashboard
					</NexusButton>
					<NexusButton variant="ghost" onClick={() => navigate(-1)}>
						<ArrowLeft className="h-4 w-4" />
						Go Back
					</NexusButton>
				</div>
			</div>
		</div>
	);
}
