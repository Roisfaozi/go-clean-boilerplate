# Task: Frontend Implementation (Next.js + Shadcn UI)

## 🎯 Objective
Create a modern, type-safe admin dashboard to interact with the Go API.

## 🛠 Specifications

### 1. Technology Stack
- **Framework**: Next.js 14 (App Router)
- **Language**: TypeScript
- **Styling**: Tailwind CSS
- **Components**: Shadcn UI (Radix Primitives)
- **Icons**: Lucide React
- **Data Fetching**: TanStack Query (React Query) v5
- **Auth**: NextAuth.js (configured for custom backend) OR manual JWT handling in storage/cookies.

### 2. Project Structure (Scaffold)
Create a new folder `frontend/` in the root of the repo.

### 3. Core Pages (MVP)
1.  **Login Page (`/login`)**:
    - Form: Username & Password.
    - Action: Call `POST /api/v1/auth/login`.
    - Handle 2FA (Future proofing) or Account Lockout errors gracefully.
2.  **Dashboard Layout**:
    - Sidebar (Collapsible).
    - Topbar (User Profile, Dark Mode Toggle).
    - Protected Route wrapper (redirect to login if no token).
3.  **User Management (`/users`)**:
    - Data Table (TanStack Table).
    - Server-side Pagination, Sorting, Filtering (Connect to `POST /api/v1/users/search`).
    - Actions: Edit, Ban (Status), Delete.

### 4. Implementation Steps
1.  Initialize Next.js app.
2.  Setup `axios` instance with Interceptors (auto-refresh token on 401).
3.  Install Shadcn components: `button`, `input`, `table`, `dropdown-menu`, `toast`.
4.  Build Login page.
5.  Build User Table.

### 5. Design System
- **Font**: Inter or Geist Sans.
- **Theme**: Slate/Zinc (Clean enterprise look).
- **Dark Mode**: Enabled by default support.
