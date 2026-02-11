# Permission Matrix Wireframe

## Overview

Spreadsheet-style RBAC interface combining Matrix overview with Card detail editing.

---

## Permission Matrix (Full View)

```
┌────────────────────────────────────────────────────────────────────────────────────┐
│ NAVBAR                                                                             │
├────────────┬───────────────────────────────────────────────────────────────────────┤
│            │                                                                       │
│  SIDEBAR   │  ┌─── PAGE HEADER ───────────────────────────────────────────────┐   │
│            │  │                                                               │   │
│            │  │  🛡️ Access Control                [+ Add Role] [+ Add Resource] │   │
│  ● Access  │  │  Configure role permissions and access policies              │   │
│            │  │                                                               │   │
│            │  └───────────────────────────────────────────────────────────────┘   │
│            │                                                                       │
│            │  ┌─── VIEW TABS ─────────────────────────────────────────────────┐   │
│            │  │  [Matrix View]  [Role Cards]  [Policy Editor]                │   │
│            │  │       ●                                                       │   │
│            │  └───────────────────────────────────────────────────────────────┘   │
│            │                                                                       │
│            │  ┌─── MATRIX GRID ───────────────────────────────────────────────┐   │
│            │  │                                                               │   │
│            │  │ ┌───────────┬────────┬────────┬────────┬────────┬────────┐   │   │
│            │  │ │   Role    │/users  │/roles  │/audit  │/access │/config │   │   │
│            │  │ ├───────────┼────────┼────────┼────────┼────────┼────────┤   │   │
│            │  │ │superadmin │ ████   │ ████   │ ████   │ ████   │ ████   │   │   │
│            │  │ │ 2 members │        │        │        │        │        │   │   │
│            │  │ ├───────────┼────────┼────────┼────────┼────────┼────────┤   │   │
│            │  │ │admin      │ ███░   │ ██░░   │ ██░░   │ ░░░░   │ ██░░   │   │   │
│            │  │ │ 5 members │        │        │        │        │        │   │   │
│            │  │ ├───────────┼────────┼────────┼────────┼────────┼────────┤   │   │
│            │  │ │editor     │ █░░░   │ ░░░░   │ █░░░   │ ░░░░   │ ░░░░   │   │   │
│            │  │ │ 12 members│        │        │        │        │        │   │   │
│            │  │ ├───────────┼────────┼────────┼────────┼────────┼────────┤   │   │
│            │  │ │viewer     │ █░░░   │ ░░░░   │ █░░░   │ ░░░░   │ ░░░░   │   │   │
│            │  │ │ 45 members│        │        │        │        │        │   │   │
│            │  │ └───────────┴────────┴────────┴────────┴────────┴────────┘   │   │
│            │  │                                                               │   │
│            │  │  Legend:  █ = Enabled   ░ = Disabled                          │   │
│            │  │           C = Create  R = Read  U = Update  D = Delete        │   │
│            │  │                                                               │   │
│            │  └───────────────────────────────────────────────────────────────┘   │
│            │                                                                       │
└────────────┴───────────────────────────────────────────────────────────────────────┘
```

---

## Permission Cell Detail

### Visual Representation

```
Full Permission (CRUD):     ████    = Read, Create, Update, Delete all enabled
Read + Create Only:         ██░░    = Read ✓, Create ✓, Update ✗, Delete ✗
Read Only:                  █░░░    = Read ✓, rest disabled
No Permission:              ░░░░    = All disabled

Color Coding:
- Enabled:   Primary/Indigo-500 (filled block)
- Disabled:  Slate-200 (light) / Slate-700 (dark)
```

### Cell Click Popup

```
┌─────────────────────────────────────────┐
│  /users Permissions for: admin          │
├─────────────────────────────────────────┤
│                                         │
│  ☑ Read      View user list and        │
│              profile details            │
│                                         │
│  ☑ Create   Add new users to the       │
│              system                     │
│                                         │
│  ☑ Update   Modify user profiles       │
│              and settings               │
│                                         │
│  ☐ Delete   Remove users from the      │
│              system permanently         │
│                                         │
├─────────────────────────────────────────┤
│      [Cancel]           [Apply]         │
└─────────────────────────────────────────┘
```

---

## Role Row Click → Slide-over Panel

```
When clicking a Role row, a slide-over panel appears from the right:

┌─── Role: admin ──────────────────────────────────────────────┐
│                                                        [✕]   │
│                                                              │
│  ┌─ Basic Information ─────────────────────────────────────┐ │
│  │                                                         │ │
│  │  Role Name *                                            │ │
│  │  ┌───────────────────────────────────────────────────┐ │ │
│  │  │ admin                                             │ │ │
│  │  └───────────────────────────────────────────────────┘ │ │
│  │                                                         │ │
│  │  Description                                            │ │
│  │  ┌───────────────────────────────────────────────────┐ │ │
│  │  │ Administrator with full system access            │ │ │
│  │  └───────────────────────────────────────────────────┘ │ │
│  │                                                         │ │
│  │  Inherits From                                          │ │
│  │  ┌───────────────────────────────────────────────────┐ │ │
│  │  │ None (Base Role)                              [▼] │ │ │
│  │  └───────────────────────────────────────────────────┘ │ │
│  │                                                         │ │
│  └─────────────────────────────────────────────────────────┘ │
│                                                              │
│  ┌─ Members (5) ───────────────────────────── [+ Assign] ─┐ │
│  │                                                         │ │
│  │  ┌─────────────────────────────────────────────────┐   │ │
│  │  │ ● John Doe         john@example.com     [✕]     │   │ │
│  │  │ ● Jane Smith       jane@example.com     [✕]     │   │ │
│  │  │ ● Alex Johnson     alex@example.com     [✕]     │   │ │
│  │  │ ● Maya Chen        maya@example.com     [✕]     │   │ │
│  │  │ ● Dev User         dev@example.com      [✕]     │   │ │
│  │  └─────────────────────────────────────────────────┘   │ │
│  │                                                         │ │
│  └─────────────────────────────────────────────────────────┘ │
│                                                              │
│  ┌─ Permissions ───────────────────────────────────────────┐ │
│  │                                                         │ │
│  │  ▸ /users                                               │ │
│  │  ┌─────────────────────────────────────────────────┐   │ │
│  │  │  ☑ Read   ☑ Create   ☑ Update   ☐ Delete      │   │ │
│  │  └─────────────────────────────────────────────────┘   │ │
│  │                                                         │ │
│  │  ▸ /roles                                               │ │
│  │  ┌─────────────────────────────────────────────────┐   │ │
│  │  │  ☑ Read   ☑ Create   ☐ Update   ☐ Delete      │   │ │
│  │  └─────────────────────────────────────────────────┘   │ │
│  │                                                         │ │
│  │  ▸ /audit                                               │ │
│  │  ┌─────────────────────────────────────────────────┐   │ │
│  │  │  ☑ Read   ☐ Create   ☑ Update   ☐ Delete      │   │ │
│  │  └─────────────────────────────────────────────────┘   │ │
│  │                                                         │ │
│  │  [+ Add Resource Permission]                            │ │
│  │                                                         │ │
│  └─────────────────────────────────────────────────────────┘ │
│                                                              │
│  ───────────────────────────────────────────────────────── │
│                                                              │
│  [🗑️ Delete Role]                       [💾 Save Changes]   │
│                                                              │
└──────────────────────────────────────────────────────────────┘
```

---

## Role Cards View (Alternative Tab)

```
┌─────────────────────────────────────────────────────────────────────────┐
│                                                                         │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐         │
│  │ 🛡️ superadmin   │  │ 🛡️ admin        │  │ 🛡️ editor       │         │
│  │                 │  │                 │  │                 │         │
│  │ Full system     │  │ System admin    │  │ Content mgmt    │         │
│  │ access          │  │ with limits     │  │ access          │         │
│  │                 │  │                 │  │                 │         │
│  │ 👥 2 members    │  │ 👥 5 members    │  │ 👥 12 members   │         │
│  │ 🔐 5 resources  │  │ 🔐 5 resources  │  │ 🔐 3 resources  │         │
│  │                 │  │                 │  │                 │         │
│  │   [Edit →]      │  │   [Edit →]      │  │   [Edit →]      │         │
│  └─────────────────┘  └─────────────────┘  └─────────────────┘         │
│                                                                         │
│  ┌─────────────────┐  ┌─────────────────┐                              │
│  │ 🛡️ viewer       │  │ + Create Role   │                              │
│  │                 │  │                 │                              │
│  │ Read-only       │  │    [+]          │                              │
│  │ access          │  │                 │                              │
│  │                 │  │  Add a new      │                              │
│  │ 👥 45 members   │  │  role to the    │                              │
│  │ 🔐 2 resources  │  │  system         │                              │
│  │                 │  │                 │                              │
│  │   [Edit →]      │  └─────────────────┘                              │
│  └─────────────────┘                                                    │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## Add New Role Modal

```
┌─── Create New Role ─────────────────────────────────────────┐
│                                                       [✕]   │
│                                                             │
│  Role Name *                                                │
│  ┌─────────────────────────────────────────────────────┐   │
│  │ moderator                                           │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                             │
│  Description                                                │
│  ┌─────────────────────────────────────────────────────┐   │
│  │ Community moderator with limited admin access       │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                             │
│  Copy Permissions From (optional)                           │
│  ┌─────────────────────────────────────────────────────┐   │
│  │ Select a role...                                [▼] │   │
│  └─────────────────────────────────────────────────────┘   │
│  ℹ️ You can customize permissions after creation            │
│                                                             │
│  Inherits From (optional)                                   │
│  ┌─────────────────────────────────────────────────────┐   │
│  │ viewer                                          [▼] │   │
│  └─────────────────────────────────────────────────────┘   │
│  ℹ️ Role will inherit all permissions from parent           │
│                                                             │
│  ┌─────────────────────────────────────────────────────┐   │
│  │  [Cancel]                           [Create Role]   │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

---

## Add Resource Modal

```
┌─── Add Resource Permission ─────────────────────────────────┐
│                                                       [✕]   │
│                                                             │
│  Resource Path *                                            │
│  ┌─────────────────────────────────────────────────────┐   │
│  │ /api/v1/                                            │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                             │
│  Or select from existing:                                   │
│  ┌─────────────────────────────────────────────────────┐   │
│  │  ☐ /users                                           │   │
│  │  ☐ /users/*                                         │   │
│  │  ☐ /roles                                           │   │
│  │  ☐ /audit                                           │   │
│  │  ☐ /config                                          │   │
│  │  ☐ /api/v1/*                                        │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                             │
│  Description (optional)                                     │
│  ┌─────────────────────────────────────────────────────┐   │
│  │ API endpoints for version 1                         │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                             │
│  ┌─────────────────────────────────────────────────────┐   │
│  │  [Cancel]                          [Add Resource]   │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

---

## Interaction States

### Matrix Cell Hover

```
┌────────┐         ┌────────┐
│ ███░   │   →     │ ███░   │  + Tooltip: "Read, Create, Update"
└────────┘         └────────┘    + Cursor: pointer
                                  + Border: primary-500
```

### Role Row Hover

```
Entire row highlights with bg-slate-50 (light) / bg-slate-800 (dark)
"Edit" link appears at row end
```

### Permission Toggle Animation

```
When toggling permission:
1. Optimistic UI update (instant visual change)
2. Background API call
3. Success: Subtle pulse animation
4. Error: Revert + Error toast
```

---

## Role Inheritance Tree View (Policy Editor Tab)

Visual representation of how roles inherit permissions from parent roles.

### Inheritance Tree Structure

```
┌─────────────────────────────────────────────────────────────────────────────────────┐
│  🌳 Role Inheritance Tree                                         [Expand All] [−] │
├─────────────────────────────────────────────────────────────────────────────────────┤
│                                                                                     │
│  📁 superadmin                                 ← Root role (no parent)             │
│  │   ├── 🔐 /users: CRUD                                                           │
│  │   ├── 🔐 /roles: CRUD                                                           │
│  │   ├── 🔐 /audit: CRUD                                                           │
│  │   ├── 🔐 /access: CRUD                                                          │
│  │   └── 🔐 /config: CRUD                                                          │
│  │                                                                                  │
│  └── 📁 admin                                  ← Inherits from superadmin          │
│      │   ├── 🔐 /users: CRU- (inherited + override: no Delete)                     │
│      │   ├── 🔐 /roles: CR-- (inherited + override)                                │
│      │   └── 🔐 /audit: R--- (override: Read only)                                 │
│      │                                                                             │
│      ├── 📁 editor                             ← Inherits from admin               │
│      │   │   ├── 🔐 /users: R--- (inherited, reduced)                              │
│      │   │   ├── 🔐 /content: CRUD (own permission)                                │
│      │   │   └── 🔐 /media: CRU- (own permission)                                  │
│      │   │                                                                         │
│      │   └── 📁 author                         ← Inherits from editor              │
│      │           ├── 🔐 /content: -RU- (inherited, reduced)                        │
│      │           └── 🔐 /media: -R-- (inherited, reduced)                          │
│      │                                                                             │
│      └── 📁 viewer                             ← Inherits from admin               │
│              ├── 🔐 /users: R--- (inherited)                                       │
│              ├── 🔐 /audit: R--- (inherited)                                       │
│              └── 🔐 /reports: R--- (own permission)                                │
│                                                                                     │
│  📁 api_client (standalone)                    ← Root role (no parent)             │
│      ├── 🔐 /api/v1/*: CR--                                                        │
│      └── 🔐 /webhooks: CRUD                                                        │
│                                                                                     │
└─────────────────────────────────────────────────────────────────────────────────────┘
```

### Legend & Indicators

```
Permission Display:
  C = Create  R = Read  U = Update  D = Delete
  - = Denied

Inheritance Indicators:
  📁 Role name           = Role node (click to expand/collapse)
  🔐 Resource: CRUD      = Permission definition
  (inherited)            = Permission comes from parent
  (own permission)       = Permission defined on this role
  (inherited, reduced)   = Inherited but with restrictions applied
  (override)             = Explicitly overrides parent permission
```

### Interactive Node States

```
Collapsed Node:
├── 📁 admin ▸                     5 children, 3 resources

Expanded Node:
├── 📁 admin ▾
│   ├── 🔐 /users: CRU-
│   └── ...

Selected Node (editing):
├── 📁 admin ◉ ─────────────────────
│   │ ┌────────────────────────────┐
│   │ │ ✏️ Click to edit role      │
│   │ └────────────────────────────┘

Hover State:
├── 📁 admin ▾  [Edit] [+ Child Role]    ← Actions appear on hover
```

### Add Child Role Flow

```
When clicking [+ Child Role] on a parent:

┌─── Create Child Role ──────────────────────────────────────────┐
│                                                          [✕]   │
│                                                                │
│  Creating child role under: admin                              │
│                                                                │
│  Role Name *                                                   │
│  ┌──────────────────────────────────────────────────────────┐ │
│  │ content_manager                                          │ │
│  └──────────────────────────────────────────────────────────┘ │
│                                                                │
│  ┌────────────────────────────────────────────────────────┐   │
│  │ ℹ️ This role will inherit all permissions from "admin"  │   │
│  │   You can customize or restrict permissions after      │   │
│  │   creation.                                            │   │
│  └────────────────────────────────────────────────────────┘   │
│                                                                │
│  Initial Permission Mode:                                      │
│  ○ Copy all permissions from parent                           │
│  ● Inherit with ability to restrict                           │
│  ○ Start with no permissions (add manually)                   │
│                                                                │
│                              [Cancel]      [Create Child Role] │
│                                                                │
└────────────────────────────────────────────────────────────────┘
```

### Effective Permissions Calculation

```
When hovering/clicking a role, show effective permissions panel:

┌─── Effective Permissions: editor ──────────────────────────────┐
│                                                                │
│  Inheritance Chain: superadmin → admin → editor               │
│                                                                │
│  Resource      │ Own    │ Inherited │ Effective │ Source      │
│  ──────────────│────────│───────────│───────────│─────────────│
│  /users        │ R---   │ CRU-      │ R---      │ editor      │
│  /roles        │ ----   │ CR--      │ CR--      │ admin       │
│  /audit        │ ----   │ R---      │ R---      │ admin       │
│  /content      │ CRUD   │ ----      │ CRUD      │ editor      │
│  /media        │ CRU-   │ ----      │ CRU-      │ editor      │
│                                                                │
│  Legend:                                                       │
│  🟢 Own = Defined directly on this role                        │
│  🔵 Inherited = Comes from parent role                         │
│  🟡 Effective = Final computed permission                      │
│                                                                │
└────────────────────────────────────────────────────────────────┘
```

### Drag & Drop Reparenting

```
Users can drag a role to a new parent:

                    📁 admin
                      │
    ╭─────────────────┼─────────────────╮
    │                 │                 │
  📁 editor    📁 viewer     📁 moderator ← Dragging
    │                                   │
    │     ╔═══════════════════════╗     │
    └────►║ Drop here to reparent ║◄────┘
          ╚═══════════════════════╝

Confirmation Dialog:
┌───────────────────────────────────────────────────────────────┐
│ ⚠️ Change Role Parent                                          │
│                                                               │
│ Moving "moderator" from "admin" to "editor" will:            │
│                                                               │
│ • Change inheritance chain                                    │
│ • May affect 3 users' effective permissions                  │
│ • Recalculate all permission grants                          │
│                                                               │
│                              [Cancel]           [Confirm Move] │
└───────────────────────────────────────────────────────────────┘
```

---

## Density Mode Variations

### Comfort vs Compact Comparison

| Element              | Comfort Mode | Compact Mode |
| :------------------- | :----------- | :----------- |
| **Cell Size**        | 60px × 48px  | 40px × 32px  |
| **Role Column**      | 160px        | 100px        |
| **Font Size**        | 14px         | 12px         |
| **Permission Icons** | 16px blocks  | 12px blocks  |
| **Member Count**     | "5 members"  | "5"          |
| **Row Padding**      | 12px         | 6px          |

### Compact Mode Matrix

```
┌──────────────────────────────────────────────────────────────────────────────┐
│ 🛡️ Access Control                           [+ Role] [+ Resource] [◐]       │
├──────────────────────────────────────────────────────────────────────────────┤
│ [Matrix]  [Cards]  [Policy]                                                  │
├──────────────────────────────────────────────────────────────────────────────┤
│ Role      │/users │/roles │/audit │/access│/config│/api   │/ws    │/batch  │
│───────────│───────│───────│───────│───────│───────│───────│───────│────────│
│superadmin │ ████  │ ████  │ ████  │ ████  │ ████  │ ████  │ ████  │ ████   │
│ (2)       │       │       │       │       │       │       │       │        │
│admin (5)  │ ███░  │ ██░░  │ ██░░  │ ░░░░  │ ██░░  │ ███░  │ █░░░  │ ██░░   │
│editor (12)│ █░░░  │ ░░░░  │ █░░░  │ ░░░░  │ ░░░░  │ █░░░  │ ░░░░  │ ░░░░   │
│viewer (45)│ █░░░  │ ░░░░  │ █░░░  │ ░░░░  │ ░░░░  │ █░░░  │ ░░░░  │ ░░░░   │
│api_client │ ░░░░  │ ░░░░  │ ░░░░  │ ░░░░  │ ░░░░  │ ████  │ ░░░░  │ ░░░░   │
│ (3)       │       │       │       │       │       │       │       │        │
└──────────────────────────────────────────────────────────────────────────────┘

Key Differences:
- More resources visible horizontally (8 vs 5 columns)
- Member count abbreviated: "(5)" instead of "5 members"
- Smaller permission blocks
- Tighter spacing for data-dense view
- Horizontal scroll for many resources
```

### Compact Mode Role Cards

```
┌─────────────────┐ ┌─────────────────┐ ┌─────────────────┐
│ 🛡️ superadmin   │ │ 👤 admin        │ │ ✏️ editor       │
│ CRUD all        │ │ CRU most        │ │ R content       │
│ 2 members       │ │ 5 members       │ │ 12 members      │
│ [Edit]          │ │ [Edit]          │ │ [Edit]          │
└─────────────────┘ └─────────────────┘ └─────────────────┘

← 120px cards instead of 200px
← Single line permission summary
```
