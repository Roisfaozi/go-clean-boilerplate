# Task: Frontend Implementation (Next.js 16+ + Shadcn UI + Redux)

## 🎯 Objective

Create a modern, type-safe admin dashboard to interact with the Go API, strictly adhering to the architecture defined in `documentation/FRONTEND_STRUCTURE.md`.

## 🛠 Specifications

### 1. Technology Stack

- **Framework**: Next.js 16+ (App Router)
- **Language**: TypeScript
- **Styling**: Tailwind CSS v4
- **Components**: Shadcn UI (Radix Primitives) + Magic UI
- **Icons**: Lucide React
- **State Management**: Redux Toolkit + RTK Query (Global & Server state)
- **Forms**: React Hook Form + Zod
- **HTTP Client**: Axios (with interceptors for Refresh Token)

### 2. Project Structure (Scaffold)

Initialize the project in `frontend/`. The structure must match the documentation:

```
frontend/src/
├── app/
│   ├── (auth)/               # Login, Register, Forgot Password
│   ├── (dashboard)/          # User Dashboard (Sidebar layout)
│   └── (admin)/              # Admin Management (Users, Roles, Audit)
├── components/
│   ├── ui/                   # Shadcn components
│   ├── magicui/              # Magic UI components
│   └── layout/               # Sidebar, Navbar, Footer
├── lib/api/                  # Centralized Axios definitions
├── store/                    # Redux store & RTK Query definitions
└── styles/                   # Global CSS
```

### 3. Core Pages (MVP)

1.  **Auth Module (`src/app/(auth)`)**:
    - `login/page.tsx`: Form with Username & Password. Dispatches `login` thunk/mutation.
    - `layout.tsx`: Centered card layout.
2.  **Dashboard Layout (`src/components/layout`)**:
    - `sidebar.tsx`: Collapsible navigation based on user role.
    - `navbar.tsx`: User profile dropdown and Dark Mode toggle.
3.  **User Management (`src/app/(admin)/admin/users`)**:
    - `page.tsx`: Implements `DataTable` with server-side pagination/sorting.
    - Integration: Uses `users-api.ts` (RTK Query) to fetch data from `POST /api/v1/users/search`.
    - Actions: Edit User, Ban User (Status update), Delete User.

### 4. Implementation Steps

1.  **Initialize**: `npx create-next-app@latest frontend` (Use TypeScript, Tailwind, ESLint, `src/` directory, App Router, `@/*` alias).
2.  **Dependencies**: Install `redux`, `react-redux`, `@reduxjs/toolkit`, `axios`, `lucide-react`, `date-fns`, `sonner`.
3.  **UI Library**:
    - Initialize Shadcn UI: `npx shadcn-ui@latest init`.
    - Install components: `button`, `input`, `table`, `dropdown-menu`, `card`, `dialog`, `form`, `select`.
4.  **State Setup**: Configure `src/store` with `auth-slice` (client state) and `api` service (RTK Query).
5.  **Network Layer**: Configure `src/lib/api/client.ts` with Axios interceptors to handle 401 Unauthorized by attempting to refresh the token via `POST /api/v1/auth/refresh`.
6.  **Develop**: Build the Login page and User Management table first.

### 5. Design System

- **Font**: Inter or Geist Sans.
- **Theme**: Slate/Zinc (Clean enterprise look).
- **Dark Mode**: Enabled by default (System sync).
