"use client";

import { createContext, useContext, type ReactNode } from "react";

interface UserSettingsContextType {
	user: any;
}

const UserSettingsContext = createContext<UserSettingsContextType | undefined>(
	undefined,
);

export function UserSettingsProvider({
	user,
	children,
}: {
	user: any;
	children: ReactNode;
}) {
	return (
		<UserSettingsContext.Provider value={{ user }}>
			{children}
		</UserSettingsContext.Provider>
	);
}

export function useUserSettings() {
	const context = useContext(UserSettingsContext);
	if (context === undefined) {
		throw new Error(
			"useUserSettings must be used within a UserSettingsProvider",
		);
	}
	return context;
}
