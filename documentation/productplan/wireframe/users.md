# User Management Wireframe

## Overview

Full CRUD interface for managing users with Hyper-Grid and detail panels.

---

## User List (Desktop - Full Grid)

```
┌────────────────────────────────────────────────────────────────────────────────────┐
│ NAVBAR                                                                             │
├────────────┬───────────────────────────────────────────────────────────────────────┤
│            │                                                                       │
│  SIDEBAR   │  ┌─── PAGE HEADER ───────────────────────────────────────────────┐   │
│            │  │                                                               │   │
│            │  │  👥 User Management                     [+ Add User]         │   │
│            │  │  Manage all system users and their roles                     │   │
│  ● Users   │  │                                                               │   │
│            │  └───────────────────────────────────────────────────────────────┘   │
│            │                                                                       │
│            │  ┌─── TOOLBAR ───────────────────────────────────────────────────┐   │
│            │  │                                                               │   │
│            │  │  [🔍 Search users...]  [Role ▼]  [Status ▼]  [Columns ▼]     │   │
│            │  │                                                               │   │
│            │  │  Selected: 0          [◐ Comfort ○ Compact]     [📤 Export]  │   │
│            │  │                                                               │   │
│            │  └───────────────────────────────────────────────────────────────┘   │
│            │                                                                       │
│            │  ┌─── HYPER-GRID ────────────────────────────────────────────────┐   │
│            │  │                                                               │   │
│            │  │ ☐ │ User           │ Email              │ Role   │ Status │ ⋮│   │
│            │  │───│────────────────│────────────────────│────────│────────│──│   │
│            │  │ ☐ │ ● John Doe     │ john@example.com   │ Admin  │ Active │⋮ │   │
│            │  │ ☐ │ ● Jane Smith   │ jane@example.com   │ Editor │ Active │⋮ │   │
│            │  │ ☐ │ ● Alex Johnson │ alex@example.com   │ Viewer │ Inactive│⋮ │   │
│            │  │ ☐ │ ● Maya Chen    │ maya@example.com   │ Admin  │ Active │⋮ │   │
│            │  │ ☐ │ ● Dev User     │ dev@example.com    │ Editor │ Pending │⋮ │   │
│            │  │                                                               │   │
│            │  └───────────────────────────────────────────────────────────────┘   │
│            │                                                                       │
│            │  ┌─── PAGINATION ────────────────────────────────────────────────┐   │
│            │  │  Rows per page: [10 ▼]    1-10 of 245      [←] 1 2 3 ... [→] │   │
│            │  └───────────────────────────────────────────────────────────────┘   │
│            │                                                                       │
└────────────┴───────────────────────────────────────────────────────────────────────┘
```

---

## Row Hover State & Actions

```
Normal Row:
│ ☐ │ ● John Doe     │ john@example.com   │ Admin  │ Active │   │

Hovered Row (actions appear):
│ ☐ │ ● John Doe     │ john@example.com   │ Admin  │ Active │[👁][✏️][⋮]│
                                                              ↑
                                                    View | Edit | More

More Menu Dropdown:
┌─────────────────┐
│ 👁 View Profile │
│ ✏️ Edit User    │
│ 🔐 Change Role  │
│ ─────────────── │
│ 🔒 Deactivate   │
│ 🗑️ Delete       │
└─────────────────┘
```

---

## Bulk Selection State

```
When rows are selected, floating action bar appears:

┌───────────────────────────────────────────────────────────────────────┐
│  ┌────────────────────────────────────────────────────────────────┐  │
│  │  ☑ 3 users selected    [Change Role ▼]  [Deactivate]  [Delete]│  │
│  │                                                         [✕]   │  │
│  └────────────────────────────────────────────────────────────────┘  │
│                              ↑                                       │
│                    Floating sticky bar at bottom                     │
└───────────────────────────────────────────────────────────────────────┘
```

---

## Add/Edit User Modal

```
┌─── Add New User ────────────────────────────────────────────┐
│                                                       [✕]   │
│                                                             │
│  ┌─ Personal Information ─────────────────────────────────┐ │
│  │                                                        │ │
│  │  Full Name *                                           │ │
│  │  ┌────────────────────────────────────────────────┐   │ │
│  │  │ John Doe                                       │   │ │
│  │  └────────────────────────────────────────────────┘   │ │
│  │                                                        │ │
│  │  Email Address *                                       │ │
│  │  ┌────────────────────────────────────────────────┐   │ │
│  │  │ john@example.com                               │   │ │
│  │  └────────────────────────────────────────────────┘   │ │
│  │                                                        │ │
│  │  Password *                     [🔑 Generate]          │ │
│  │  ┌────────────────────────────────────────────────┐   │ │
│  │  │ ●●●●●●●●                              [👁]     │   │ │
│  │  └────────────────────────────────────────────────┘   │ │
│  │                                                        │ │
│  └────────────────────────────────────────────────────────┘ │
│                                                             │
│  ┌─ Access & Permissions ─────────────────────────────────┐ │
│  │                                                        │ │
│  │  Assign Role *                                         │ │
│  │  ┌────────────────────────────────────────────────┐   │ │
│  │  │ Admin                                      [▼] │   │ │
│  │  └────────────────────────────────────────────────┘   │ │
│  │                                                        │ │
│  │  Status                                                │ │
│  │  ○ Active   ○ Pending   ○ Inactive                    │ │
│  │                                                        │ │
│  │  ☐ Send welcome email with login credentials          │ │
│  │                                                        │ │
│  └────────────────────────────────────────────────────────┘ │
│                                                             │
│  ┌────────────────────────────────────────────────────────┐ │
│  │  [Cancel]                               [Create User]  │ │
│  └────────────────────────────────────────────────────────┘ │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

---

## User Profile Card (Slide-over Detail)

```
┌─── User: John Doe ─────────────────────────────────────────┐
│                                                      [✕]   │
│                                                            │
│  ┌──────────────────────────────────────────────────────┐ │
│  │                                                      │ │
│  │           ┌───────┐                                  │ │
│  │           │  JD   │   ← Avatar 80px                  │ │
│  │           └───────┘                                  │ │
│  │                                                      │ │
│  │        John Doe                                      │ │
│  │        john@example.com                              │ │
│  │        🟢 Active                                     │ │
│  │                                                      │ │
│  │      [✏️ Edit Profile]   [🔐 Reset Password]         │ │
│  │                                                      │ │
│  └──────────────────────────────────────────────────────┘ │
│                                                            │
│  ▼ Account Information                                     │
│  ┌──────────────────────────────────────────────────────┐ │
│  │  Created:        January 15, 2026                    │ │
│  │  Last Login:     2 hours ago                         │ │
│  │  IP Address:     192.168.1.1                         │ │
│  │  User Agent:     Chrome/120 (Windows)                │ │
│  └──────────────────────────────────────────────────────┘ │
│                                                            │
│  ▼ Assigned Roles                                          │
│  ┌──────────────────────────────────────────────────────┐ │
│  │  ┌───────┐                                           │ │
│  │  │ Admin │ ✕                                         │ │
│  │  └───────┘                                           │ │
│  │                                                      │ │
│  │  [+ Assign Role]                                     │ │
│  └──────────────────────────────────────────────────────┘ │
│                                                            │
│  ▼ Recent Activity                                         │
│  ┌──────────────────────────────────────────────────────┐ │
│  │  • Updated role settings - 2h ago                    │ │
│  │  • Created user maya@example.com - 5h ago            │ │
│  │  • Logged in - 1d ago                                │ │
│  │                                                      │ │
│  │  [View All Activity →]                               │ │
│  └──────────────────────────────────────────────────────┘ │
│                                                            │
│  ─────────────────────────────────────────────────────── │
│                                                            │
│  [🗑️ Delete User]                                          │
│                                                            │
└────────────────────────────────────────────────────────────┘
```

---

## Status Badges

| Status   | Light Mode                       | Dark Mode                            | Icon |
| :------- | :------------------------------- | :----------------------------------- | :--- |
| Active   | `bg-emerald-50 text-emerald-700` | `bg-emerald-500/10 text-emerald-400` | 🟢   |
| Pending  | `bg-amber-50 text-amber-700`     | `bg-amber-500/10 text-amber-400`     | 🟡   |
| Inactive | `bg-slate-50 text-slate-500`     | `bg-slate-500/10 text-slate-400`     | ⚪   |
| Locked   | `bg-red-50 text-red-700`         | `bg-red-500/10 text-red-400`         | 🔴   |

---

## Empty State

```
┌─────────────────────────────────────────────────────────┐
│                                                         │
│                        👥                               │
│                   No users found                        │
│                                                         │
│     Try adjusting your filters or add a new user       │
│                                                         │
│              [+ Add Your First User]                    │
│                                                         │
└─────────────────────────────────────────────────────────┘
```

---

## Density Mode Variations

### Comfort vs Compact Comparison

| Element          | Comfort Mode     | Compact Mode |
| :--------------- | :--------------- | :----------- |
| **Row Height**   | 64px             | 40px         |
| **Avatar Size**  | 40px circle      | 24px circle  |
| **Font Size**    | 14px             | 13px         |
| **Padding**      | 16px             | 8px          |
| **Action Icons** | 20px             | 16px         |
| **Status Badge** | Full text + icon | Icon only    |

### Compact Mode Grid

```
┌──────────────────────────────────────────────────────────────────────────┐
│ 👥 Users                               [+ Add]   [◐] [📤]                │
├──────────────────────────────────────────────────────────────────────────┤
│ [🔍 Search...] [Role ▼] [Status ▼]                    Selected: 0        │
├──────────────────────────────────────────────────────────────────────────┤
│ ☐ │ ●│ John Doe      │ john@example.com   │ Admin │🟢│⋮│                 │
│ ☐ │ ●│ Jane Smith    │ jane@example.com   │Editor │🟢│⋮│  ← 40px rows    │
│ ☐ │ ●│ Alex Johnson  │ alex@example.com   │Viewer │⚪│⋮│                 │
│ ☐ │ ●│ Maya Chen     │ maya@example.com   │ Admin │🟢│⋮│                 │
│ ☐ │ ●│ Dev User      │ dev@example.com    │Editor │🟡│⋮│                 │
├──────────────────────────────────────────────────────────────────────────┤
│ [10 ▼] 1-10 of 245                              [←] 1 2 3 ... 25 [→]     │
└──────────────────────────────────────────────────────────────────────────┘

Key Differences:
- Smaller avatars (24px vs 40px)
- Status shown as icon only (🟢🟡⚪🔴)
- Abbreviated column headers
- Tighter row padding
- More rows visible per screen
```

### Compact Mode Modal

```
┌─── Add User ────────────────────────────────────────┐
│                                                [✕]  │
│  Name *         [________________]                  │
│  Email *        [________________]                  │  ← Single column
│  Password *     [________] [🔑]                     │    inline labels
│  Role *         [Admin        ▼]                    │    32px inputs
│  Status         ○ Active ○ Pending ○ Inactive       │
│  ☐ Send welcome email                               │
│                                                     │
│           [Cancel]          [Create]                │
└─────────────────────────────────────────────────────┘
```
