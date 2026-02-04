# Empty States & Loading Patterns Wireframe

## Overview

Comprehensive empty state designs for all screens, following NexusOS design language with helpful guidance and clear CTAs.

---

## Design Principles for Empty States

| Principle              | Implementation                                |
| :--------------------- | :-------------------------------------------- |
| **Helpful, not empty** | Always explain why it's empty and what to do  |
| **Visual interest**    | Use illustrations or icons, not just text     |
| **Clear CTA**          | One primary action to resolve the empty state |
| **Contextual**         | Message relates to the specific feature       |
| **Density-aware**      | Illustrations scale with Comfort/Compact mode |

---

## Dashboard Empty States

### First-Time User (Onboarding)

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                                                                             │
│                                                                             │
│                           ┌─────────────────────┐                           │
│                           │                     │                           │
│                           │    🚀               │                           │
│                           │   ╱  ╲              │  ← Illustration:          │
│                           │  ╱    ╲             │    Rocket launch          │
│                           │ ╱______╲            │    (Lottie animation)     │
│                           │    ▓▓                │                           │
│                           │   ▓▓▓▓               │                           │
│                           │                     │                           │
│                           └─────────────────────┘                           │
│                                                                             │
│                       Welcome to NexusOS! 🎉                                │
│                                                                             │
│              Your admin dashboard is ready to be configured.                │
│              Let's set up your first user and role.                        │
│                                                                             │
│                       [🚀 Start Setup Wizard]                               │
│                                                                             │
│                         Skip for now →                                      │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### No Recent Activity

```
┌─────────────────────────────────────────────────────────────────────────────┐
│  📊 Recent Activity                                                         │
│  ─────────────────────────────────────────────────────────────────────────  │
│                                                                             │
│                           ┌─────────────────────┐                           │
│                           │                     │                           │
│                           │      📭            │                           │
│                           │     ╱  ╲           │  ← Empty inbox icon       │
│                           │    ╱    ╲          │                           │
│                           │   ╱______╲         │                           │
│                           │                     │                           │
│                           └─────────────────────┘                           │
│                                                                             │
│                      No activity recorded yet                               │
│                                                                             │
│           Activity will appear here as users interact                       │
│                    with your application.                                   │
│                                                                             │
│                      [View All Audit Logs]                                  │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## User Management Empty States

### No Users

```
┌─────────────────────────────────────────────────────────────────────────────┐
│  👥 Users                                            [+ Add User]           │
│  ─────────────────────────────────────────────────────────────────────────  │
│                                                                             │
│  [🔍 Search users...]                                                       │
│                                                                             │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │                                                                     │   │
│  │                                                                     │   │
│  │                     ┌───────────────────┐                           │   │
│  │                     │      👥           │                           │   │
│  │                     │    ╭─────╮        │                           │   │
│  │                     │    │     │        │                           │   │
│  │                     │    │  +  │        │  ← Group icon with +      │   │
│  │                     │    │     │        │                           │   │
│  │                     │    ╰─────╯        │                           │   │
│  │                     └───────────────────┘                           │   │
│  │                                                                     │   │
│  │                     No users yet                                    │   │
│  │                                                                     │   │
│  │           Add your first user to start managing access.            │   │
│  │                                                                     │   │
│  │                 [👤 Add Your First User]                            │   │
│  │                                                                     │   │
│  │          Or invite team members via email:                          │   │
│  │          [📧 Send Invitations]                                      │   │
│  │                                                                     │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### No Search Results

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                                                                             │
│                     ┌───────────────────┐                                   │
│                     │       🔍          │                                   │
│                     │      ╱  ╲         │                                   │
│                     │     ╱ ✕  ╲        │  ← Magnifying glass with X        │
│                     │    ╱──────╲       │                                   │
│                     └───────────────────┘                                   │
│                                                                             │
│                 No results for "john smith"                                 │
│                                                                             │
│                      Try adjusting your search:                             │
│                                                                             │
│                 • Check for typos or spelling                               │
│                 • Try using fewer or different keywords                     │
│                 • Remove some filters                                       │
│                                                                             │
│                        [Clear Search]                                       │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Filter Returns No Results

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                                                                             │
│  Active Filters:                                                            │
│  ┌─────────┐ ┌──────────────┐ ┌─────────────────┐                          │
│  │Role:Admin✕│ │Status:Active✕│ │Created: Today ✕│    [Clear All]          │
│  └─────────┘ └──────────────┘ └─────────────────┘                          │
│                                                                             │
│                     ┌───────────────────┐                                   │
│                     │       🔧          │                                   │
│                     │     ╱ │ ╲         │  ← Filter funnel empty            │
│                     │    ╱  │  ╲        │                                   │
│                     │   ╱   │   ╲       │                                   │
│                     │  ╱    ▼    ╲      │                                   │
│                     └───────────────────┘                                   │
│                                                                             │
│              No users match your current filters                            │
│                                                                             │
│          3 filters applied • 47 total users in system                      │
│                                                                             │
│    [Remove "Created: Today" filter]    [Clear All Filters]                  │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Permission Matrix Empty States

### No Roles Defined

```
┌─────────────────────────────────────────────────────────────────────────────┐
│  🛡️ Access Control                                   [+ Add Role]           │
│  ─────────────────────────────────────────────────────────────────────────  │
│                                                                             │
│                                                                             │
│                     ┌───────────────────┐                                   │
│                     │       🛡️          │                                   │
│                     │       ╱╲          │                                   │
│                     │      ╱  ╲         │  ← Shield with lock               │
│                     │     ╱ 🔒 ╲        │                                   │
│                     │    ╱──────╲       │                                   │
│                     └───────────────────┘                                   │
│                                                                             │
│                  No roles configured yet                                    │
│                                                                             │
│        Roles define what users can do in your application.                 │
│        Start with common roles like Admin, Editor, and Viewer.             │
│                                                                             │
│                    [🛡️ Create First Role]                                   │
│                                                                             │
│               Or use our template:                                          │
│        [ Import Default Roles (Admin, Editor, Viewer) ]                    │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### No Resources Defined

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                                                                             │
│  ┌─────────┬─────────────────────────────────────────────────────────────┐ │
│  │   Role  │                    No Resources                             │ │
│  ├─────────┼─────────────────────────────────────────────────────────────┤ │
│  │ admin   │                                                             │ │
│  │ editor  │       ┌───────────────────┐                                 │ │
│  │ viewer  │       │       📁          │                                 │ │
│  │         │       │     ─────         │  ← Folder with dotted outline   │ │
│  │         │       │    │ · · │        │                                 │ │
│  │         │       │     ─────         │                                 │ │
│  │         │       └───────────────────┘                                 │ │
│  │         │                                                             │ │
│  │         │       Add resources to configure permissions                │ │
│  │         │                                                             │ │
│  │         │                [+ Add Resource]                             │ │
│  └─────────┴─────────────────────────────────────────────────────────────┘ │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Audit Logs Empty States

### No Logs Yet

```
┌─────────────────────────────────────────────────────────────────────────────┐
│  📋 Audit Logs                                                              │
│  ─────────────────────────────────────────────────────────────────────────  │
│                                                                             │
│                     ┌───────────────────┐                                   │
│                     │       📋          │                                   │
│                     │      ────         │                                   │
│                     │     │    │        │  ← Clipboard, empty               │
│                     │     │    │        │                                   │
│                     │     │    │        │                                   │
│                     │      ────         │                                   │
│                     └───────────────────┘                                   │
│                                                                             │
│                   No activity logged yet                                    │
│                                                                             │
│         Audit logs will automatically record all user actions              │
│         once activity begins in your application.                          │
│                                                                             │
│         Logs include: User actions, API calls, login attempts,             │
│         permission changes, and system events.                             │
│                                                                             │
│                   [📖 Learn About Audit Logs]                               │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### No Logs in Date Range

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                                                                             │
│  Date Range: Dec 1, 2025 — Dec 15, 2025                                    │
│                                                                             │
│                     ┌───────────────────┐                                   │
│                     │       📅          │                                   │
│                     │     ╱────╲        │  ← Calendar with X                │
│                     │    │  ✕  │        │                                   │
│                     │     ╲────╱        │                                   │
│                     └───────────────────┘                                   │
│                                                                             │
│             No activity found in this date range                           │
│                                                                             │
│                 Try selecting a different period:                           │
│                                                                             │
│      [Today]  [Last 7 Days]  [Last 30 Days]  [All Time]                    │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Settings Empty States

### No API Keys

```
┌─────────────────────────────────────────────────────────────────────────────┐
│  🔑 API Keys                                        [+ Create API Key]      │
│  ─────────────────────────────────────────────────────────────────────────  │
│                                                                             │
│                     ┌───────────────────┐                                   │
│                     │       🔑          │                                   │
│                     │      ╭───╮        │                                   │
│                     │      │   │        │  ← Key icon                       │
│                     │      │ + │        │                                   │
│                     │      ╰───╯        │                                   │
│                     └───────────────────┘                                   │
│                                                                             │
│                   No API keys created                                       │
│                                                                             │
│       API keys allow external applications to interact with                │
│       your NexusOS instance programmatically.                              │
│                                                                             │
│                   [🔑 Generate Your First API Key]                          │
│                                                                             │
│               📖 Read the API Documentation →                               │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### No Active Sessions (Other Devices)

```
┌─────────────────────────────────────────────────────────────────────────────┐
│  💻 Active Sessions                                                         │
│  ─────────────────────────────────────────────────────────────────────────  │
│                                                                             │
│  ● This device (current)                                                   │
│    Chrome on Windows • Jakarta, Indonesia                                   │
│    Active now                                             [Current]        │
│                                                                             │
│  ────────────────────────────────────────────────────────────────────────  │
│                                                                             │
│                     ┌───────────────────┐                                   │
│                     │       ✓           │                                   │
│                     │      ╭─╮          │  ← Checkmark, secure              │
│                     │      │✓│          │                                   │
│                     │      ╰─╯          │                                   │
│                     └───────────────────┘                                   │
│                                                                             │
│              You're only logged in on this device                          │
│                                                                             │
│         This is good! You have no other active sessions.                   │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## AI Chat Empty State

```
┌─────────────────────────────────────────┐
│ 🤖 NexusAI                        ─  ✕ │
├─────────────────────────────────────────┤
│                                         │
│           ┌─────────────────┐           │
│           │       🤖        │           │
│           │     ╭─────╮     │           │
│           │     │ ✨  │     │  ← AI bot │
│           │     ╰─────╯     │    with   │
│           │    ╱       ╲    │    sparkle│
│           └─────────────────┘           │
│                                         │
│     Hi! I'm NexusAI, your admin        │
│     assistant.                          │
│                                         │
│     Try asking me:                      │
│                                         │
│   "Show users who logged in today"     │
│   "How do I create a new role?"        │
│   "Explain the permission matrix"      │
│                                         │
├─────────────────────────────────────────┤
│  ┌─────────────────────────────────┐   │
│  │ Ask NexusAI anything...    [→]  │   │
│  └─────────────────────────────────┘   │
└─────────────────────────────────────────┘
```

---

## Loading Skeleton States

### Table Loading

```
┌─────────────────────────────────────────────────────────────────────────────┐
│ Timestamp    │ User        │ Action │ Resource    │ Status                  │
│──────────────│─────────────│────────│─────────────│─────────────────────────│
│ ▓▓▓▓▓▓▓▓░░░░ │ ● ▓▓▓▓▓▓░░ │ ▓▓▓▓░░ │ ▓▓▓▓▓▓▓▓░░ │ ▓▓▓░░                   │
│ ▓▓▓▓▓▓▓░░░░░ │ ● ▓▓▓▓▓░░░ │ ▓▓▓▓▓░ │ ▓▓▓▓▓▓░░░░ │ ▓▓▓░░                   │
│ ▓▓▓▓▓▓▓▓▓░░░ │ ● ▓▓▓▓░░░░ │ ▓▓▓░░░ │ ▓▓▓▓▓▓▓░░░ │ ▓▓▓░░                   │
│ ▓▓▓▓▓▓░░░░░░ │ ● ▓▓▓▓▓▓▓░ │ ▓▓▓▓▓░ │ ▓▓▓▓▓░░░░░ │ ▓▓▓░░                   │
│ ▓▓▓▓▓▓▓▓░░░░ │ ● ▓▓▓▓▓░░░ │ ▓▓▓▓░░ │ ▓▓▓▓▓▓▓▓▓░ │ ▓▓▓░░                   │
└─────────────────────────────────────────────────────────────────────────────┘

Animation: Shimmer gradient from left to right
           bg-gradient from slate-200 to slate-100 to slate-200
           1.5s infinite
```

### Card Loading

```
┌─────────────────────────────────────────┐
│                                         │
│   ▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓░░░░░░░░░░░░░░░░░    │  ← Shimmer animation
│                                         │
│   ▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓░░░░░░░░░░░░░░░  │
│   ▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓░░░░░░░░░░░░░░░░░░  │
│                                         │
│   ▓▓▓▓▓▓▓▓▓░░░░░░░                      │
│                                         │
└─────────────────────────────────────────┘
```

### KPI Card Loading

```
┌─────────────────────┐ ┌─────────────────────┐
│ ▓▓▓▓▓▓▓▓░░░░        │ │ ▓▓▓▓▓▓▓▓▓░░░        │
│                     │ │                     │
│ ▓▓▓▓▓▓▓▓▓▓▓▓░░░░░░  │ │ ▓▓▓▓▓▓▓▓▓▓▓░░░░░░░  │
│                     │ │                     │
│ ▓▓▓▓▓░░░  ▓▓░░      │ │ ▓▓▓▓▓▓░░  ▓▓░░      │
└─────────────────────┘ └─────────────────────┘
```

---

## Error States

### Connection Error

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                                                                             │
│                     ┌───────────────────┐                                   │
│                     │       ⚠️          │                                   │
│                     │      ╱ ╲          │  ← Warning triangle               │
│                     │     ╱   ╲         │                                   │
│                     │    ╱  !  ╲        │                                   │
│                     │   ╱───────╲       │                                   │
│                     └───────────────────┘                                   │
│                                                                             │
│                  Unable to load data                                        │
│                                                                             │
│       We couldn't connect to the server. This might be a                   │
│       temporary issue. Please try again.                                   │
│                                                                             │
│                       [🔄 Try Again]                                        │
│                                                                             │
│              If the problem persists, contact support.                      │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Access Denied

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                                                                             │
│                     ┌───────────────────┐                                   │
│                     │       🔒          │                                   │
│                     │      ╭───╮        │  ← Padlock, closed                │
│                     │      │   │        │                                   │
│                     │      ╰───╯        │                                   │
│                     │     ╱─────╲       │                                   │
│                     └───────────────────┘                                   │
│                                                                             │
│                     Access Denied                                           │
│                                                                             │
│       You don't have permission to view this page.                         │
│       Your current role: Editor                                            │
│                                                                             │
│       Required permission: admin:read                                       │
│                                                                             │
│                     [← Go Back]                                             │
│                                                                             │
│       Need access? Contact your administrator.                             │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Compact Mode Variations

In Compact mode, empty states are smaller:

```
Comfort Mode:                           Compact Mode:
┌─────────────────────────────────┐    ┌─────────────────────────────────┐
│                                 │    │      📭 No users yet            │
│         ┌───────────────┐       │    │                                 │
│         │      📭       │       │    │  Add your first user to start. │
│         │    ╱    ╲     │       │    │                                 │
│         │   ╱      ╲    │       │    │  [+ Add User]                   │
│         │  ╱        ╲   │       │    └─────────────────────────────────┘
│         └───────────────┘       │
│                                 │    ↑ Illustration replaced with
│      No users yet               │      inline emoji icon
│                                 │    ↑ Text condensed
│  Add your first user to start  │    ↑ Less vertical padding
│  managing access control.       │
│                                 │
│      [👤 Add Your First User]   │
│                                 │
└─────────────────────────────────┘
```
