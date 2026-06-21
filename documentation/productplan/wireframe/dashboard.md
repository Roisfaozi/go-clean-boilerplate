# Dashboard Wireframe (Hybrid Layout)

## Overview

The main dashboard implements the **Hybrid** approach with KPI Cards + Hyper-Grid + Quick Actions.

---

## Full Layout (Desktop - 1440px)

```
┌────────────────────────────────────────────────────────────────────────────────────┐
│  ┌─ NAVBAR ──────────────────────────────────────────────────────────────────────┐ │
│  │ [≡] NexusOS        [🔍 Search... ⌘K]     [◐ Comfort]  [🌙]  [🔔 3] [👤 Admin ▼]│ │
│  └────────────────────────────────────────────────────────────────────────────────┘ │
├────────────┬───────────────────────────────────────────────────────────────────────┤
│            │                                                                       │
│  SIDEBAR   │  ┌─── KPI CARDS (Zone A) ────────────────────────────────────────┐   │
│  280px     │  │                                                               │   │
│            │  │  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐      │   │
│ ┌────────┐ │  │  │  👥      │  │  🛡️      │  │  📊      │  │  ⚠️       │      │   │
│ │ 📊     │ │  │  │  245     │  │  12      │  │  1,847   │  │  3        │      │   │
│ │Dashboard│ │  │  │ Users    │  │ Roles    │  │ Actions  │  │ Alerts   │      │   │
│ │◄ Active│ │  │  │  +12%    │  │  stable  │  │  today   │  │  pending │      │   │
│ └────────┘ │  │  └──────────┘  └──────────┘  └──────────┘  └──────────┘      │   │
│            │  │                                                               │   │
│ ┌────────┐ │  └───────────────────────────────────────────────────────────────┘   │
│ │ 👥     │ │                                                                       │
│ │ Users  │ │  ┌─── HYPER-GRID: Recent Activity (Zone B) ──────────────────────┐   │
│ └────────┘ │  │                                                               │   │
│            │  │  ┌─────────────────────────────────────────────────────────┐  │   │
│ ┌────────┐ │  │  │ Time      │ User        │ Action  │ Resource   │ Status│  │   │
│ │ 🔐     │ │  │  ├─────────────────────────────────────────────────────────┤  │   │
│ │ Roles  │ │  │  │ 2m ago    │ ● john      │ CREATE  │ /users/5   │ ✓ OK  │  │   │
│ └────────┘ │  │  │ 5m ago    │ ● jane      │ UPDATE  │ /roles/2   │ ✓ OK  │  │   │
│            │  │  │ 12m ago   │ ● alex      │ DELETE  │ /users/3   │ ✓ OK  │  │   │
│ ┌────────┐ │  │  │ 1h ago    │ ● system    │ LOGIN   │ /auth      │ ✗ FAIL│  │   │
│ │ 🛡️     │ │  │  │ 2h ago    │ ● maya      │ EXPORT  │ /audit     │ ✓ OK  │  │   │
│ │ Access │ │  │  └─────────────────────────────────────────────────────────┘  │   │
│ └────────┘ │  │                                                               │   │
│            │  │                              [View All Logs →]                │   │
│ ┌────────┐ │  └───────────────────────────────────────────────────────────────┘   │
│ │ 📋     │ │                                                                       │
│ │ Audit  │ │  ┌─── QUICK ACTIONS (Zone C) ────────────────────────────────────┐   │
│ └────────┘ │  │                                                               │   │
│            │  │  [+ Add User]    [+ Create Role]    [📤 Export Logs]    [⚙️]  │   │
│ ┌────────┐ │  │                                                               │   │
│ │ ⚙️     │ │  └───────────────────────────────────────────────────────────────┘   │
│ │Settings│ │                                                                       │
│ └────────┘ │                                                                       │
│            │                                                                       │
│ ──────── │                                                                       │
│ ┌────────┐ │                                                                       │
│ │ 🤖 AI  │ │                                                                       │
│ └────────┘ │                                                                       │
└────────────┴───────────────────────────────────────────────────────────────────────┘
```

---

## Compact Mode Variation

```
┌──────────────────────────────────────────────────────────────────────────────────┐
│ [≡] NexusOS  [🔍 ⌘K]  [● Compact]  [🌙]  [🔔]  [👤]                             │
├──────┬───────────────────────────────────────────────────────────────────────────┤
│      │  ┌─────┐ ┌─────┐ ┌─────┐ ┌─────┐                                         │
│ 72px │  │ 245 │ │  12 │ │1847 │ │  3  │      ← Minimal KPI (no icons, sparklines)│
│      │  │Users│ │Roles│ │Acts │ │Alert│                                         │
│ [📊] │  └─────┘ └─────┘ └─────┘ └─────┘                                         │
│ [👥] │                                                                           │
│ [🔐] │  ┌─────────────────────────────────────────────────────────────────────┐ │
│ [🛡️] │  │Time    │User    │Action│Resource    │IP           │Status│Details│ │
│ [📋] │  │────────│────────│──────│────────────│─────────────│──────│───────│ │
│ [⚙️] │  │2m      │john    │CREATE│/users/5    │192.168.1.1  │✓     │[...]  │ │
│      │  │5m      │jane    │UPDATE│/roles/2    │192.168.1.2  │✓     │[...]  │ │
│ ──── │  │12m     │alex    │DELETE│/users/3    │192.168.1.1  │✓     │[...]  │ │
│ [🤖] │  │1h      │system  │LOGIN │/auth       │10.0.0.1     │✗     │[...]  │ │
│      │  │2h      │maya    │EXPORT│/audit      │192.168.1.5  │✓     │[...]  │ │
│      │  │2h      │dev     │READ  │/config     │192.168.1.3  │✓     │[...]  │ │
└──────┴──│3h      │admin   │UPDATE│/settings   │192.168.1.1  │✓     │[...]  │─┘
          └─────────────────────────────────────────────────────────────────────┘
```

---

## KPI Card Specifications

### Comfort Mode Card

```
┌────────────────────────────────────┐
│                                    │
│     ┌───┐                          │
│     │ 👥 │   ← Icon 48px, circle bg│
│     └───┘                          │
│                                    │
│      245                           │  ← Value: 32px Bold
│    Total Users                     │  ← Label: 14px Muted
│                                    │
│    ▲ +12% from last week          │  ← Trend: 12px + color
│                                    │
└────────────────────────────────────┘

Size: ~200px width
Shadow: shadow-lg
Border: none (light) / 1px slate-800 (dark)
Radius: 16px
Padding: 24px
```

### Compact Mode Card

```
┌──────────────────────┐
│  245    ▁▂▃▂▄▃▅     │  ← Value 20px + Sparkline
│  Users   +12%       │  ← Label 12px + Trend
└──────────────────────┘

Size: ~150px width
Shadow: none
Border: 1px slate-200
Radius: 4px
Padding: 12px
```

---

## Quick Actions Bar

```
Comfort Mode:
┌────────────────────────────────────────────────────────────────────────┐
│                                                                        │
│  [👤 + Add User]    [🛡️ + Create Role]    [📤 Export Logs]    [⚙️]    │
│                                                                        │
│  Button: 44px height, rounded-xl, with icon + text                     │
└────────────────────────────────────────────────────────────────────────┘

Compact Mode:
┌──────────────────────────────────────────────────────────────────────────┐
│ [+User] [+Role] [Export] [⚙️]                                           │
│ Button: 32px height, rounded-md, icon only or short text                │
└──────────────────────────────────────────────────────────────────────────┘
```

---

## Responsive Breakpoints

| Breakpoint     | KPI Layout | Grid          | Sidebar     |
| :------------- | :--------- | :------------ | :---------- |
| XL (1440+)     | 4 columns  | Full featured | Expanded    |
| LG (1024-1439) | 4 columns  | Full          | Collapsible |
| MD (768-1023)  | 2×2 grid   | Simplified    | Drawer      |
| SM (<768)      | Stacked    | Card view     | Bottom nav  |

---

## Mobile Layout (< 768px)

```
┌──────────────────────────┐
│ NexusOS        [🔔] [👤] │
├──────────────────────────┤
│  ┌──────┐  ┌──────┐     │
│  │ 245  │  │  12  │     │
│  │Users │  │Roles │     │
│  └──────┘  └──────┘     │
│                          │
│  ┌──────┐  ┌──────┐     │
│  │ 1847 │  │  3   │     │
│  │ Acts │  │Alerts│     │
│  └──────┘  └──────┘     │
│                          │
│  Recent Activity         │
│  ┌──────────────────────┐│
│  │ ● john - CREATE      ││
│  │   /users/5 - 2m ago  ││
│  ├──────────────────────┤│
│  │ ● jane - UPDATE      ││
│  │   /roles/2 - 5m ago  ││
│  └──────────────────────┘│
│                          │
│ [+ Add User] [+ Role]   │
│                          │
├──────────────────────────┤
│ [📊] [👥] [🔐] [📋] [⚙️]│
└──────────────────────────┘
     ↑ Bottom Navigation
```

---

## Interaction Notes

1. **KPI Cards** - Click to navigate to full list view
2. **Hyper-Grid rows** - Click to view detail, hover shows actions
3. **Quick Actions** - Primary buttons open modals
4. **AI Chat icon** - Click to open dockable chat panel
5. **Density Toggle** - Instant switch with smooth 300ms transition
