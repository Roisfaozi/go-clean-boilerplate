"use client";

import { createContext, useContext, ReactNode } from "react";
import { User } from "~/lib/api/users";

interface ProfileContextType {
  user: User | any;
}

const ProfileContext = createContext<ProfileContextType | undefined>(undefined);

export function ProfileProvider({
  user,
  children,
}: {
  user: User | any;
  children: ReactNode;
}) {
  return (
    <ProfileContext.Provider value={{ user }}>
      {children}
    </ProfileContext.Provider>
  );
}

export function useProfile() {
  const context = useContext(ProfileContext);
  if (context === undefined) {
    throw new Error("useProfile must be used within a ProfileProvider");
  }
  return context;
}
