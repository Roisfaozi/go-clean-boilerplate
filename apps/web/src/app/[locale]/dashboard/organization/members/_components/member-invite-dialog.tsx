"use client";

import { useState } from "react";
import { useMembers } from "./members-context";
import { useOrganizationStore } from "~/stores/use-organization-store";
import { Button } from "@casbin/ui";
import { Icon } from "~/components/shared/icon";
import {
	Dialog,
	DialogContent,
	DialogDescription,
	DialogFooter,
	DialogHeader,
	DialogTitle,
	DialogTrigger,
} from "@casbin/ui";
import {
	Select,
	SelectContent,
	SelectItem,
	SelectTrigger,
	SelectValue,
} from "@casbin/ui";
import { Input } from "@casbin/ui";
import { Label } from "@casbin/ui";

export function MemberInviteDialog() {
	const { currentOrganization } = useOrganizationStore();
	const { roles, inviteMember, isInviting } = useMembers();
	const [email, setEmail] = useState("");
	const [roleId, setRoleId] = useState("");
	const [isOpen, setIsOpen] = useState(false);

	const handleSubmit = async () => {
		if (!email || !roleId) return;
		try {
			await inviteMember(email, roleId);
			setIsOpen(false);
			setEmail("");
			setRoleId("");
		} catch (_error) {
			// Error handled in context
		}
	};

	if (!currentOrganization) return null;

	return (
		<Dialog open={isOpen} onOpenChange={setIsOpen}>
			<DialogTrigger
				render={
					<Button size="sm">
						<Icon name="UserPlus" className="mr-2 h-4 w-4" />
						Invite Member
					</Button>
				}
			/>
			<DialogContent>
				<DialogHeader>
					<DialogTitle>Invite new member</DialogTitle>
					<DialogDescription>
						Invite someone to join {currentOrganization.name} by their email
						address.
					</DialogDescription>
				</DialogHeader>
				<div className="grid gap-4 py-4">
					<div className="grid gap-2">
						<Label htmlFor="email">Email address</Label>
						<Input
							id="email"
							type="email"
							placeholder="colleague@example.com"
							value={email}
							onChange={(e) => setEmail(e.target.value)}
						/>
					</div>
					<div className="grid gap-2">
						<Label htmlFor="role">Initial Role</Label>
						<Select value={roleId} onValueChange={setRoleId}>
							<SelectTrigger id="role">
								<SelectValue placeholder="Select a role" />
							</SelectTrigger>
							<SelectContent>
								{roles.map((role) => (
									<SelectItem key={role.id} value={role.id}>
										{role.name}
									</SelectItem>
								))}
							</SelectContent>
						</Select>
					</div>
				</div>
				<DialogFooter>
					<Button variant="outline" onClick={() => setIsOpen(false)}>
						Cancel
					</Button>
					<Button
						onClick={handleSubmit}
						disabled={isInviting || !email || !roleId}
					>
						{isInviting && (
							<Icon name="Loader" className="mr-2 h-4 w-4 animate-spin" />
						)}
						Send Invitation
					</Button>
				</DialogFooter>
			</DialogContent>
		</Dialog>
	);
}
