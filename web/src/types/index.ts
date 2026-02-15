// Defines global types used in the frontend

export interface User {
  id: string;
  name: string;
  email: string;
  role: string;
  picture?: string;
  avatarUrl?: string;
  status?: "active" | "suspended" | "banned";
}

export interface payload {
  name: string;
  email: string;
  picture?: string;
}

export interface Session {
  user: User;
  accessToken: string;
  expiresAt: string;
}

export interface AuthState {
  user: User | null;
  isAuthenticated: boolean;
  isLoading: boolean;
}

export interface Project {
  id: string;
  organization_id: string;
  user_id: string;
  name: string;
  domain: string;
  status: string;
  created_at: number;
  updated_at: number;
}

export interface SendOTPProps {
  toMail: string;
  code: string;
  userName: string;
}

export interface SendWelcomeEmailProps {
  toMail: string;
  userName: string;
}

export interface SubscriptionPlan {
  name: string;
  description: string;
  stripePriceId: string;
}
