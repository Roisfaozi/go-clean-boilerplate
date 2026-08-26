import { ServerCrash, Home, RefreshCw } from "lucide-react";
import { useNavigate } from "react-router";
import { NexusButton } from "@casbin/ui";

export default function Error500V1() {
	const navigate = useNavigate();
	return (
		<div className="bg-background flex min-h-screen">
			<div className="bg-destructive/5 relative hidden w-1/2 items-center justify-center overflow-hidden lg:flex">
				<div className="absolute inset-0">
					{[...Array(5)].map((_, i) => (
						<div
							key={i}
							className="border-destructive/10 absolute rounded-full border"
							style={{
								width: `${200 + i * 120}px`,
								height: `${200 + i * 120}px`,
								top: "50%",
								left: "50%",
								transform: "translate(-50%, -50%)",
							}}
						/>
					))}
				</div>
				<div className="relative z-10 space-y-6 px-12 text-center">
					<div className="bg-destructive/10 mx-auto flex h-24 w-24 items-center justify-center rounded-full">
						<ServerCrash className="text-destructive h-12 w-12" />
					</div>
					<h2 className="text-foreground text-3xl font-bold">Server Error</h2>
					<p className="text-muted-foreground max-w-md text-lg">
						Something went wrong on our end. Our team has been notified and is
						working on a fix.
					</p>
				</div>
			</div>

			<div className="flex flex-1 items-center justify-center p-8">
				<div className="w-full max-w-md space-y-8 text-center">
					<div className="space-y-2">
						<p className="text-destructive/20 text-8xl font-black">500</p>
						<h1 className="text-foreground text-3xl font-bold">
							Internal Server Error
						</h1>
						<p className="text-muted-foreground">
							We're experiencing technical difficulties. Please try again in a
							few moments.
						</p>
					</div>
					<div className="flex flex-col justify-center gap-3 sm:flex-row">
						<NexusButton
							variant="primary"
							onClick={() => window.location.reload()}
						>
							<RefreshCw className="h-4 w-4" />
							Try Again
						</NexusButton>
						<NexusButton variant="outline" onClick={() => navigate("/")}>
							<Home className="h-4 w-4" />
							Go Home
						</NexusButton>
					</div>
				</div>
			</div>
		</div>
	);
}
