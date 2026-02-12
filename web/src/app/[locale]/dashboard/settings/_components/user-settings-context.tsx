"use client";

import { createContext, useContext, ReactNode } from "react";
import { User } from "~/lib/api/users";

interface UserSettingsContextType {
  user: any;
}

const UserSettingsContext = createContext<UserSettingsContextType | undefined>(undefined);

export function UserSettingsProvider({ user, children }: { user: any, children: ReactNode }) {
  return (
    <UserSettingsContext.Provider value={{ user }}>
      {children}
    </UserSettingsContext.Provider>
  );
}

export function useUserSettings() {
  const context = useContext(UserSettingsContext);
  if (context === undefined) {
    throw new Error("useUserSettings must be used within a UserSettingsProvider");
  }
  return context;
}
