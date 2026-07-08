import { useState } from "react";
import { NexusCard } from "@casbin/ui";
import { NexusInput } from "@casbin/ui";
import { NexusButton } from "@casbin/ui";
import { Avatar, AvatarFallback, AvatarImage } from "@casbin/ui";
import { Badge } from "@casbin/ui";
import { MemberRoleSelector } from "./member-role-selector";
import {
	DropdownMenu,
	DropdownMenuContent,
	DropdownMenuItem,
	DropdownMenuTrigger,
} from "@casbin/ui";
import { useToast } from "@casbin/ui";
import { Search, MoreHorizontal, UserMinus, Mail } from "lucide-react";
import { cn } from "@casbin/ui";

interface Member {
	id: string;
	name: string;
	email: string;
	avatar?: string;
	role: string;
	status: "online" | "away" | "offline";
	joinedAt: string;
}

const mockMembers: Member[] = [
	{
		id: "1",
		name: "Alice Johnson",
		email: "alice@acme.com",
		role: "owner",
		status: "online",
		joinedAt: "2024-01-15",
	},
	{
		id: "2",
		name: "Bob Smith",
		email: "bob@acme.com",
		role: "admin",
		status: "online",
		joinedAt: "2024-02-20",
	},
	{
		id: "3",
		name: "Carol Williams",
		email: "carol@acme.com",
		role: "editor",
		status: "away",
		joinedAt: "2024-03-01",
	},
	{
		id: "4",
		name: "David Brown",
		email: "david@acme.com",
		role: "editor",
		status: "offline",
		joinedAt: "2024-03-05",
	},
	{
		id: "5",
		name: "Eva Martinez",
		email: "eva@acme.com",
		role: "viewer",
		status: "online",
		joinedAt: "2024-03-08",
	},
	{
		id: "6",
		name: "Frank Lee",
		email: "frank@acme.com",
		role: "viewer",
		status: "offline",
		joinedAt: "2024-03-10",
	},
];

const statusColors: Record<string, string> = {
	online: "bg-success",
	away: "bg-warning",
	offline: "bg-muted-foreground/40",
};

export function MembersTable() {
	const { toast } = useToast();
	const [search, setSearch] = useState("");
	const [members, setMembers] = useState<Member[]>(mockMembers);

	const filtered = members.filter(
		(m) =>
			m.name.toLowerCase().includes(search.toLowerCase()) ||
			m.email.toLowerCase().includes(search.toLowerCase()),
	);

	const handleRoleChange = (memberId: string, newRole: string) => {
		setMembers((prev) =>
			prev.map((m) => (m.id === memberId ? { ...m, role: newRole } : m)),
		);
		toast({ title: "Role Updated", description: `Changed to ${newRole}` });
	};

	const handleRemove = (member: Member) => {
		setMembers((prev) => prev.filter((m) => m.id !== member.id));
		toast({ title: "Member Removed", description: member.email });
	};

	const addMember = (email: string, role: string) => {
		const newMember: Member = {
			id: String(Date.now()),
			name: email.split("@")[0],
			email,
			role,
			status: "offline",
			joinedAt: new Date().toISOString().slice(0, 10),
		};
		setMembers((prev) => [...prev, newMember]);
	};

	return {
		filtered,
		search,
		setSearch,
		handleRoleChange,
		handleRemove,
		addMember,
		members,
	};
}

export function MembersTableUI() {
	const { filtered, search, setSearch, handleRoleChange, handleRemove } =
		MembersTable();

	return (
		<NexusCard className="overflow-hidden">
			<div className="border-border border-b p-4">
				<div className="relative max-w-sm">
					<Search className="text-muted-foreground absolute top-1/2 left-3 h-4 w-4 -translate-y-1/2" />
					<NexusInput
						placeholder="Search members..."
						value={search}
						onChange={(e) => setSearch(e.target.value)}
						className="h-9 pl-9"
					/>
				</div>
			</div>
			<div className="overflow-x-auto">
				<table className="w-full text-sm">
					<thead>
						<tr className="border-border bg-muted/50 border-b">
							<th className="text-muted-foreground px-4 py-3 text-left font-medium">
								Member
							</th>
							<th className="text-muted-foreground px-4 py-3 text-left font-medium">
								Status
							</th>
							<th className="text-muted-foreground px-4 py-3 text-left font-medium">
								Role
							</th>
							<th className="text-muted-foreground px-4 py-3 text-left font-medium">
								Joined
							</th>
							<th className="text-muted-foreground w-12 px-4 py-3 text-right font-medium"></th>
						</tr>
					</thead>
					<tbody>
						{filtered.map((member) => (
							<tr
								key={member.id}
								className="border-border hover:bg-muted/20 border-b transition-colors last:border-0"
							>
								<td className="px-4 py-3">
									<div className="flex items-center gap-3">
										<div className="relative">
											<Avatar className="h-8 w-8">
												<AvatarImage src={member.avatar} />
												<AvatarFallback className="bg-primary/10 text-primary text-xs">
													{member.name
														.split(" ")
														.map((n) => n[0])
														.join("")}
												</AvatarFallback>
											</Avatar>
											<span
												className={cn(
													"border-background absolute -right-0.5 -bottom-0.5 h-3 w-3 rounded-full border-2",
													statusColors[member.status],
												)}
											/>
										</div>
										<div>
											<p className="text-foreground font-medium">
												{member.name}
											</p>
											<p className="text-muted-foreground text-xs">
												{member.email}
											</p>
										</div>
									</div>
								</td>
								<td className="px-4 py-3">
									<Badge
										variant={member.status === "online" ? "default" : "outline"}
										className={cn(
											"text-[10px] capitalize",
											member.status === "online" &&
												"bg-success/10 text-success border-success/20",
											member.status === "away" &&
												"bg-warning/10 text-warning border-warning/20",
										)}
									>
										{member.status}
									</Badge>
								</td>
								<td className="px-4 py-3">
									<MemberRoleSelector
										value={member.role}
										onChange={(val) => handleRoleChange(member.id, val)}
										disabled={member.role === "owner"}
									/>
								</td>
								<td className="text-muted-foreground px-4 py-3 text-xs">
									{member.joinedAt}
								</td>
								<td className="px-4 py-3 text-right">
									<DropdownMenu>
										<DropdownMenuTrigger asChild>
											<NexusButton
												variant="ghost"
												size="icon"
												className="h-8 w-8"
											>
												<MoreHorizontal className="h-4 w-4" />
											</NexusButton>
										</DropdownMenuTrigger>
										<DropdownMenuContent align="end">
											<DropdownMenuItem>
												<Mail className="mr-2 h-4 w-4" />
												Resend Invite
											</DropdownMenuItem>
											<DropdownMenuItem
												className="text-destructive focus:text-destructive"
												onClick={() => handleRemove(member)}
												disabled={member.role === "owner"}
											>
												<UserMinus className="mr-2 h-4 w-4" />
												Remove Member
											</DropdownMenuItem>
										</DropdownMenuContent>
									</DropdownMenu>
								</td>
							</tr>
						))}
						{filtered.length === 0 && (
							<tr>
								<td
									colSpan={5}
									className="text-muted-foreground px-4 py-8 text-center"
								>
									No members found
								</td>
							</tr>
						)}
					</tbody>
				</table>
			</div>
		</NexusCard>
	);
}
