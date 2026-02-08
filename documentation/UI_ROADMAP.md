# UI Roadmap Specification

## Go Clean Boilerplate - Admin Dashboard & User Portal

---

## Overview

This document outlines the phased UI development roadmap for the Go Clean Boilerplate application, covering both the **Admin Dashboard** and **User Portal**.

**Tech Stack Recommendation:**

- **Framework**: Next.js 16+ (App Router)
- **Styling**: Tailwind CSS + shadcn/ui + Magic UI
- **State Management**: Redux Toolkit / TanStack Query
- **Auth**: JWT with HttpOnly cookies

---

## Phase 0: Landing Page

**Duration**: 1 week | **Priority**: 🔴 Critical

### Route

| Page    | Route | Description                           |
| ------- | ----- | ------------------------------------- |
| Landing | `/`   | Public homepage with product overview |

### Sections

#### 1. Hero Section

- **Headline**: Tagline utama aplikasi
- **Subheadline**: Deskripsi singkat value proposition
- **CTA Buttons**: "Get Started" → `/register`, "Learn More" → scroll to features
- **Hero Visual**: Ilustrasi/screenshot dashboard (Magic UI animated)
- **Background**: Gradient mesh atau animated particles

#### 2. Features Section

- Grid 3-4 kolom dengan icon + title + description
- Features: RBAC, Audit Logging, Real-time, API-First
- Hover animations (Magic UI)

#### 3. How It Works

- Step-by-step flow (1-2-3 numbered)
- Visual diagram atau ilustrasi proses

#### 4. Tech Stack / Integration

- Logo grid teknologi yang didukung
- Go, MySQL, Redis, WebSocket, JWT

#### 5. Testimonials / Use Cases (Optional)

- Carousel atau grid testimonial
- Company logos jika ada

#### 6. Pricing Section (Optional)

- Pricing tiers jika aplikasi berbayar
- Free / Pro / Enterprise

#### 7. CTA Section

- Call-to-action final sebelum footer
- "Start Building Today" button

#### 8. Footer

- Navigation links
- Social media icons
- Copyright & legal links

### Components

- [ ] Navbar (transparent → solid on scroll)
- [ ] Hero with animated background
- [ ] Feature Card with hover effects
- [ ] Step Timeline component
- [ ] Logo Marquee (tech stack)
- [ ] Testimonial Carousel
- [ ] Pricing Table
- [ ] Footer with sitemap

### Design Guidelines

- **Style**: Modern, glassmorphism, dark mode primary
- **Animations**: Scroll-triggered, subtle micro-interactions
- **Performance**: Lazy load images, optimize LCP < 2.5s
- **SEO**: Meta tags, Open Graph, structured data
- **Mobile**: Fully responsive, mobile-first

### Magic UI Components to Use

- `AnimatedBeam` for hero background
- `BentoGrid` for features
- `Marquee` for tech logos
- `NumberTicker` for stats
- `ShimmerButton` for CTAs

---

## Phase 1: Authentication & Core Layout

**Duration**: 1-2 weeks

### Pages

| Page               | Route              | Description                              |
| ------------------ | ------------------ | ---------------------------------------- |
| Login              | `/login`           | Email/username + password authentication |
| Register           | `/register`        | User registration with validation        |
| Forgot Password    | `/forgot-password` | Request password reset email             |
| Reset Password     | `/reset-password`  | Set new password with token              |
| Email Verification | `/verify-email`    | Verify email with token                  |

### Components

- [ ] Auth Layout (centered card)
- [ ] Login Form with validation
- [ ] Register Form with password strength indicator
- [ ] Toast notifications
- [ ] Loading states & error handling

### API Integration

```
POST /api/v1/auth/login
POST /api/v1/users/register
POST /api/v1/auth/forgot-password
POST /api/v1/auth/reset-password
POST /api/v1/auth/verify-email
```

---

## Phase 2: User Portal

**Duration**: 1-2 weeks

### Pages

| Page      | Route        | Description                          |
| --------- | ------------ | ------------------------------------ |
| Dashboard | `/dashboard` | User home with summary stats         |
| Profile   | `/profile`   | View/edit profile, avatar upload     |
| Settings  | `/settings`  | Account settings, email verification |

### Components

- [ ] App Layout (sidebar + header)
- [ ] User Avatar Upload
- [ ] Profile Edit Form
- [ ] Email Verification Banner
- [ ] Session Management

### API Integration

```
GET    /api/v1/users/me
PUT    /api/v1/users/me
PATCH  /api/v1/users/me/avatar
POST   /api/v1/auth/resend-verification
POST   /api/v1/auth/logout
```

---

## Phase 3: Admin - User Management

**Duration**: 2 weeks

### Pages

| Page        | Route               | Description                      |
| ----------- | ------------------- | -------------------------------- |
| User List   | `/admin/users`      | Paginated, searchable user table |
| User Detail | `/admin/users/[id]` | View/edit user details           |

### Components

- [ ] Data Table with sorting, filtering, pagination
- [ ] Dynamic Search Builder (uses `/search` endpoint)
- [ ] User Status Badge (active/banned/pending)
- [ ] Ban/Reactivate User Modal
- [ ] Delete User Confirmation

### API Integration

```
GET    /api/v1/users
POST   /api/v1/users/search
GET    /api/v1/users/:id
PATCH  /api/v1/users/:id/status
DELETE /api/v1/users/:id
```

---

## Phase 4: Admin - RBAC Management

**Duration**: 2-3 weeks

### Pages

| Page              | Route                           | Description                 |
| ----------------- | ------------------------------- | --------------------------- |
| Role List         | `/admin/roles`                  | All roles with member count |
| Role Permissions  | `/admin/roles/[id]/permissions` | Manage role permissions     |
| Permission Matrix | `/admin/permissions`            | Visual permission grid      |

### Components

- [ ] Role CRUD Modal
- [ ] Assign Role to User
- [ ] Revoke Role from User
- [ ] Permission Grant/Revoke Toggle
- [ ] Role Inheritance Tree (parent/child)
- [ ] Permission Batch Check

### API Integration

```
POST   /api/v1/roles
GET    /api/v1/roles
DELETE /api/v1/roles/:id
POST   /api/v1/permissions/assign-role
DELETE /api/v1/permissions/revoke-role
POST   /api/v1/permissions/grant
DELETE /api/v1/permissions/revoke
GET    /api/v1/permissions
GET    /api/v1/permissions/:role
POST   /api/v1/permissions/inheritance
DELETE /api/v1/permissions/inheritance
POST   /api/v1/permissions/check-batch
```

---

## Phase 5: Admin - Access Rights & Endpoints

**Status**: ✅ Completed (UI Overhaul with Accordion & Checkboxes)

### Pages

| Page          | Route                  | Description                     |
| ------------- | ---------------------- | ------------------------------- |
| Access Rights | `/admin/access-rights` | Manage access right definitions |
| API Endpoints | `/admin/endpoints`     | Manage registered endpoints     |

### Components

- [x] Access Right CRUD (Accordion-based)
- [x] Endpoint Registration Form
- [x] Link/Unlink Endpoint to Access Right (Inline Toggle)
- [x] Module-based Endpoint Grouping
- [ ] Endpoint Discovery (auto-detect routes)

### API Integration

```
POST   /api/v1/access-rights
GET    /api/v1/access-rights
DELETE /api/v1/access-rights/:id
POST   /api/v1/access-rights/link
POST   /api/v1/access-rights/unlink
POST   /api/v1/endpoints
DELETE /api/v1/endpoints/:id
```

---

## Phase 6: Admin - Audit Logs

**Duration**: 1 week

### Pages

| Page       | Route               | Description              |
| ---------- | ------------------- | ------------------------ |
| Audit Logs | `/admin/audit-logs` | Searchable activity logs |

### Components

- [ ] Log Table with infinite scroll
- [ ] Advanced Filter Panel (date, action, user)
- [ ] Log Detail Modal
- [ ] Export to CSV

### API Integration

```
POST /api/v1/audit-logs/search
```

---

## Phase 7: Real-time Features

**Duration**: 1-2 weeks

### Features

- [ ] WebSocket Connection Manager
- [ ] Real-time Notifications
- [ ] SSE Event Subscription
- [ ] Online User Presence

### Endpoints

```
WS  /ws
SSE /events
```

---

## Priority Matrix

| Phase | Priority    | Complexity | Dependencies |
| ----- | ----------- | ---------- | ------------ |
| 0     | 🔴 Critical | Medium     | None         |
| 1     | 🔴 Critical | Low        | Phase 0      |
| 2     | 🔴 Critical | Medium     | Phase 1      |
| 3     | 🟡 High     | Medium     | Phase 2      |
| 4     | 🟡 High     | High       | Phase 3      |
| 5     | 🟢 Medium   | Medium     | Phase 4      |
| 6     | 🟢 Medium   | Low        | Phase 3      |
| 7     | 🔵 Low      | High       | Phase 2      |

---

## Milestones

| Milestone | Target  | Deliverables                     |
| --------- | ------- | -------------------------------- |
| Launch    | Week 1  | Landing Page                     |
| MVP       | Week 5  | Auth + User Portal + Basic Admin |
| Beta      | Week 9  | Full Admin Dashboard             |
| v1.0      | Week 11 | Real-time + Polish               |
