import { PageHeader } from "@/components/layout/page-header";
import {
	NexusCard,
	NexusCardHeader,
	NexusCardTitle,
	NexusCardContent,
} from "@casbin/ui";
import { NexusButton } from "@casbin/ui";
import { NexusInput } from "@casbin/ui";
import { FormGroup } from "@/components/patterns/form-group";
import { useUIStore } from "@/stores";
import { Sun, Moon, Maximize2, Minimize2 } from "lucide-react";

export default function SettingsPage() {
	const { theme, setTheme, density, setDensity } = useUIStore();

	return (
		<div className="space-y-6">
			<PageHeader
				title="Settings"
				description="Manage your workspace preferences."
			/>

			<div className="gap-gap grid grid-cols-1 lg:grid-cols-2">
				<NexusCard>
					<NexusCardHeader>
						<NexusCardTitle>Appearance</NexusCardTitle>
					</NexusCardHeader>
					<NexusCardContent className="space-y-4">
						<div className="flex items-center justify-between">
							<div>
								<p className="text-body font-medium">Theme</p>
								<p className="text-caption text-muted-foreground">
									Toggle between light and dark mode
								</p>
							</div>
							<NexusButton
								variant="outline"
								size="sm"
								onClick={() => setTheme(theme === "light" ? "dark" : "light")}
							>
								{theme === "light" ? (
									<Moon className="mr-2 h-4 w-4" />
								) : (
									<Sun className="mr-2 h-4 w-4" />
								)}
								{theme === "light" ? "Dark" : "Light"}
							</NexusButton>
						</div>
						<div className="flex items-center justify-between">
							<div>
								<p className="text-body font-medium">Density</p>
								<p className="text-caption text-muted-foreground">
									Adjust UI spacing density
								</p>
							</div>
							<NexusButton
								variant="outline"
								size="sm"
								onClick={() =>
									setDensity(density === "comfort" ? "compact" : "comfort")
								}
							>
								{density === "comfort" ? (
									<Minimize2 className="mr-2 h-4 w-4" />
								) : (
									<Maximize2 className="mr-2 h-4 w-4" />
								)}
								{density === "comfort" ? "Compact" : "Comfort"}
							</NexusButton>
						</div>
					</NexusCardContent>
				</NexusCard>

				<NexusCard>
					<NexusCardHeader>
						<NexusCardTitle>Profile</NexusCardTitle>
					</NexusCardHeader>
					<NexusCardContent className="space-y-4">
						<FormGroup label="Display Name">
							<NexusInput placeholder="Your name" defaultValue="Admin User" />
						</FormGroup>
						<FormGroup label="Email">
							<NexusInput
								placeholder="Email"
								defaultValue="admin@nexus.io"
								type="email"
							/>
						</FormGroup>
						<NexusButton>Save Changes</NexusButton>
					</NexusCardContent>
				</NexusCard>
			</div>
		</div>
	);
}
