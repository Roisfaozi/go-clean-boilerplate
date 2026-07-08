import { Badge } from "@casbin/ui";
import {
	Select,
	SelectContent,
	SelectItem,
	SelectTrigger,
	SelectValue,
} from "@casbin/ui";

const roles = [
	{ value: "owner", label: "Owner", color: "destructive" as const },
	{ value: "admin", label: "Admin", color: "default" as const },
	{ value: "editor", label: "Editor", color: "secondary" as const },
	{ value: "viewer", label: "Viewer", color: "outline" as const },
];

interface MemberRoleSelectorProps {
	value: string;
	onChange: (value: string) => void;
	disabled?: boolean;
}

export function MemberRoleSelector({
	value,
	onChange,
	disabled,
}: MemberRoleSelectorProps) {
	return (
		<Select value={value} onValueChange={onChange} disabled={disabled}>
			<SelectTrigger className="h-8 w-[120px] text-xs">
				<SelectValue />
			</SelectTrigger>
			<SelectContent>
				{roles.map((role) => (
					<SelectItem key={role.value} value={role.value}>
						<div className="flex items-center gap-2">
							<Badge variant={role.color} className="px-1.5 py-0 text-[10px]">
								{role.label}
							</Badge>
						</div>
					</SelectItem>
				))}
			</SelectContent>
		</Select>
	);
}

export function MemberRoleBadge({ role }: { role: string }) {
	const found = roles.find((r) => r.value === role);
	return (
		<Badge variant={found?.color ?? "outline"} className="text-[10px]">
			{found?.label ?? role}
		</Badge>
	);
}
