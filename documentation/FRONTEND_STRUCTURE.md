# Frontend Structure Analysis

## Next.js 16+ App Router Architecture

---

## Project Structure

```
frontend/
├── .env.local                    # Environment variables
├── .env.example                  # Environment template
├── next.config.ts                # Next.js configuration
├── tailwind.config.ts            # Tailwind CSS config
├── tsconfig.json                 # TypeScript config
├── package.json
│
├── public/                       # Static assets
│   ├── images/
│   ├── icons/
│   └── favicon.ico
│
├── src/
│   ├── app/                      # App Router (Pages)
│   │   ├── layout.tsx            # Root layout
│   │   ├── page.tsx              # Landing page (/)
│   │   ├── loading.tsx           # Global loading
│   │   ├── error.tsx             # Global error
│   │   ├── not-found.tsx         # 404 page
│   │   │
│   │   ├── (auth)/               # Auth group (no layout nesting)
│   │   │   ├── layout.tsx        # Auth layout (centered card)
│   │   │   ├── login/page.tsx
│   │   │   ├── register/page.tsx
│   │   │   ├── forgot-password/page.tsx
│   │   │   ├── reset-password/page.tsx
│   │   │   └── verify-email/page.tsx
│   │   │
│   │   ├── (dashboard)/          # Dashboard group (authenticated)
│   │   │   ├── layout.tsx        # Dashboard layout (sidebar)
│   │   │   ├── dashboard/page.tsx
│   │   │   ├── profile/page.tsx
│   │   │   └── settings/page.tsx
│   │   │
│   │   └── (admin)/              # Admin group (authorized)
│   │       ├── layout.tsx        # Admin layout
│   │       ├── admin/
│   │       │   ├── page.tsx      # Admin dashboard
│   │       │   ├── users/
│   │       │   │   ├── page.tsx
│   │       │   │   └── [id]/page.tsx
│   │       │   ├── roles/
│   │       │   │   ├── page.tsx
│   │       │   │   └── [id]/permissions/page.tsx
│   │       │   ├── permissions/page.tsx
│   │       │   ├── access-rights/page.tsx
│   │       │   ├── endpoints/page.tsx
│   │       │   └── audit-logs/page.tsx
│   │       └── ...
│   │
│   ├── components/               # Reusable components
│   │   ├── ui/                   # shadcn/ui components
│   │   │   ├── button.tsx
│   │   │   ├── input.tsx
│   │   │   ├── card.tsx
│   │   │   ├── dialog.tsx
│   │   │   ├── dropdown-menu.tsx
│   │   │   ├── data-table.tsx
│   │   │   └── ...
│   │   │
│   │   ├── magicui/              # Magic UI components
│   │   │   ├── animated-beam.tsx
│   │   │   ├── bento-grid.tsx
│   │   │   ├── marquee.tsx
│   │   │   ├── shimmer-button.tsx
│   │   │   └── number-ticker.tsx
│   │   │
│   │   ├── layout/               # Layout components
│   │   │   ├── navbar.tsx
│   │   │   ├── sidebar.tsx
│   │   │   ├── footer.tsx
│   │   │   ├── header.tsx
│   │   │   └── breadcrumb.tsx
│   │   │
│   │   ├── landing/              # Landing page components
│   │   │   ├── hero-section.tsx
│   │   │   ├── features-section.tsx
│   │   │   ├── how-it-works.tsx
│   │   │   ├── tech-stack.tsx
│   │   │   ├── testimonials.tsx
│   │   │   ├── pricing-section.tsx
│   │   │   └── cta-section.tsx
│   │   │
│   │   ├── auth/                 # Auth components
│   │   │   ├── login-form.tsx
│   │   │   ├── register-form.tsx
│   │   │   ├── forgot-password-form.tsx
│   │   │   ├── reset-password-form.tsx
│   │   │   └── social-auth-buttons.tsx
│   │   │
│   │   ├── dashboard/            # Dashboard components
│   │   │   ├── stats-card.tsx
│   │   │   ├── activity-feed.tsx
│   │   │   └── quick-actions.tsx
│   │   │
│   │   ├── users/                # User management components
│   │   │   ├── user-table.tsx
│   │   │   ├── user-form.tsx
│   │   │   ├── user-status-badge.tsx
│   │   │   └── user-avatar.tsx
│   │   │
│   │   ├── roles/                # Role management components
│   │   │   ├── role-table.tsx
│   │   │   ├── role-form.tsx
│   │   │   └── role-permissions.tsx
│   │   │
│   │   ├── permissions/          # Permission components
│   │   │   ├── permission-matrix.tsx
│   │   │   ├── permission-toggle.tsx
│   │   │   └── inheritance-tree.tsx
│   │   │
│   │   └── shared/               # Shared components
│   │       ├── loading-spinner.tsx
│   │       ├── empty-state.tsx
│   │       ├── confirm-dialog.tsx
│   │       ├── search-input.tsx
│   │       └── pagination.tsx
│   │
│   ├── lib/                      # Utilities & configurations
│   │   ├── api/                  # API client
│   │   │   ├── client.ts         # Axios/fetch instance
│   │   │   ├── auth.ts           # Auth API calls
│   │   │   ├── users.ts          # User API calls
│   │   │   ├── roles.ts          # Role API calls
│   │   │   ├── permissions.ts    # Permission API calls
│   │   │   └── audit.ts          # Audit API calls
│   │   │
│   │   ├── utils.ts              # Helper functions
│   │   ├── cn.ts                 # classNames utility
│   │   ├── constants.ts          # App constants
│   │   └── validations.ts        # Zod schemas
│   │
│   ├── hooks/                    # Custom React hooks
│   │   ├── use-auth.ts           # Auth state hook
│   │   ├── use-user.ts           # Current user hook
│   │   ├── use-permissions.ts    # Permission check hook
│   │   ├── use-debounce.ts       # Debounce hook
│   │   ├── use-media-query.ts    # Responsive hook
│   │   └── use-websocket.ts      # WebSocket hook
│   │
│   ├── store/                    # Redux Toolkit store
│   │   ├── index.ts              # Store configuration
│   │   ├── provider.tsx          # Redux provider
│   │   ├── slices/
│   │   │   ├── auth-slice.ts     # Auth state
│   │   │   ├── user-slice.ts     # User state
│   │   │   └── ui-slice.ts       # UI state (sidebar, theme)
│   │   └── api/                  # RTK Query APIs
│   │       ├── auth-api.ts
│   │       ├── users-api.ts
│   │       └── roles-api.ts
│   │
│   ├── types/                    # TypeScript types
│   │   ├── api.ts                # API response types
│   │   ├── user.ts               # User types
│   │   ├── role.ts               # Role types
│   │   ├── permission.ts         # Permission types
│   │   └── audit.ts              # Audit log types
│   │
│   ├── styles/                   # Global styles
│   │   └── globals.css           # Tailwind imports + custom
│   │
│   └── middleware.ts             # Next.js middleware (auth guard)
│
└── tests/                        # Test files
    ├── e2e/                      # Playwright E2E tests
    └── unit/                     # Jest unit tests
```

---

## Key Architecture Decisions

### 1. Route Groups

- `(auth)` - Unauthenticated pages dengan centered layout
- `(dashboard)` - Authenticated user pages dengan sidebar
- `(admin)` - Admin pages dengan additional authorization

### 2. State Management

```
Redux Toolkit + RTK Query
├── Global state (auth, user, UI preferences)
├── Server state caching (RTK Query)
└── Form state (React Hook Form + Zod)
```

### 3. API Layer

```typescript
// lib/api/client.ts
const apiClient = axios.create({
  baseURL: process.env.NEXT_PUBLIC_API_URL,
  withCredentials: true, // untuk HttpOnly cookies
});

// Interceptors untuk refresh token
apiClient.interceptors.response.use(
  (response) => response,
  async (error) => {
    if (error.response?.status === 401) {
      // Refresh token logic
    }
    return Promise.reject(error);
  },
);
```

### 4. Authentication Flow

```
1. Login → Set JWT in HttpOnly cookie
2. Middleware checks cookie on protected routes
3. RTK Query fetches /users/me on app load
4. Permission checks via custom hook
```

### 5. Component Patterns

```typescript
// Compound component pattern untuk complex UI
<DataTable>
  <DataTable.Header>
  <DataTable.Body>
  <DataTable.Pagination>
</DataTable>

// Server components untuk data fetching
// Client components untuk interactivity
```

---

## File Naming Conventions

| Type       | Convention       | Example                      |
| ---------- | ---------------- | ---------------------------- |
| Pages      | `page.tsx`       | `app/login/page.tsx`         |
| Layouts    | `layout.tsx`     | `app/(dashboard)/layout.tsx` |
| Components | `kebab-case.tsx` | `user-status-badge.tsx`      |
| Hooks      | `use-*.ts`       | `use-auth.ts`                |
| Types      | `*.ts`           | `user.ts`                    |
| API        | `*-api.ts`       | `users-api.ts`               |
| Slices     | `*-slice.ts`     | `auth-slice.ts`              |

---

## Dependencies

### Core

```json
{
  "next": "^16.0.0",
  "react": "^19.0.0",
  "typescript": "^5.3.0"
}
```

### UI

```json
{
  "tailwindcss": "^4.0.0",
  "@radix-ui/react-*": "latest",
  "class-variance-authority": "^0.7.0",
  "clsx": "^2.1.0",
  "lucide-react": "^0.300.0"
}
```

### State & Data

```json
{
  "@reduxjs/toolkit": "^2.0.0",
  "react-redux": "^9.0.0",
  "@tanstack/react-query": "^5.0.0",
  "axios": "^1.6.0"
}
```

### Forms & Validation

```json
{
  "react-hook-form": "^7.50.0",
  "zod": "^3.22.0",
  "@hookform/resolvers": "^3.3.0"
}
```

### Utilities

```json
{
  "date-fns": "^3.0.0",
  "sonner": "^1.3.0"
}
```

---

## Environment Variables

```env
# .env.local
NEXT_PUBLIC_API_URL=http://localhost:8080/api/v1
NEXT_PUBLIC_WS_URL=ws://localhost:8080/ws
NEXT_PUBLIC_APP_NAME="Go Clean Dashboard"
```

---

## Next Steps

1. [ ] Initialize Next.js project dengan `npx create-next-app@latest`
2. [ ] Setup shadcn/ui: `npx shadcn-ui@latest init`
3. [ ] Install Magic UI components
4. [ ] Configure Redux Toolkit store
5. [ ] Create API client layer
6. [ ] Implement auth middleware
7. [ ] Build Landing Page (Phase 0)
