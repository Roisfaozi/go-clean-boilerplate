# Audit Logs Wireframe (Hyper-Grid)

## Overview

Enterprise-grade activity monitoring with advanced filtering and export capabilities.

---

## Audit Logs Page (Full View)

```
┌────────────────────────────────────────────────────────────────────────────────────┐
│ NAVBAR                                                                             │
├────────────┬───────────────────────────────────────────────────────────────────────┤
│            │                                                                       │
│  SIDEBAR   │  ┌─── PAGE HEADER ───────────────────────────────────────────────┐   │
│            │  │                                                               │   │
│            │  │  📋 Audit Logs                               [📊 Analytics]  │   │
│  ● Audit   │  │  Monitor all system activity and user actions                │   │
│            │  │                                                               │   │
│            │  └───────────────────────────────────────────────────────────────┘   │
│            │                                                                       │
│            │  ┌─── FILTER BAR ────────────────────────────────────────────────┐   │
│            │  │                                                               │   │
│            │  │  [🔍 Search logs...]                                          │   │
│            │  │                                                               │   │
│            │  │  [📅 Date Range ▼]  [👤 User ▼]  [📊 Action ▼]  [🔗 Resource ▼] │  │
│            │  │                                                               │   │
│            │  │  Active Filters:                                              │   │
│            │  │  ┌─────────┐ ┌──────────────┐ ┌─────────────────┐            │   │
│            │  │  │Today  ✕│ │Action:CREATE✕│ │Status:Failed ✕ │            │   │
│            │  │  └─────────┘ └──────────────┘ └─────────────────┘            │   │
│            │  │                                                [Clear All]   │   │
│            │  └───────────────────────────────────────────────────────────────┘   │
│            │                                                                       │
│            │  ┌─── TOOLBAR ───────────────────────────────────────────────────┐   │
│            │  │                                                               │   │
│            │  │  [Columns ▼]   [◐ Comfort ● Compact]        [📤 Export ▼]    │   │
│            │  │                                                               │   │
│            │  └───────────────────────────────────────────────────────────────┘   │
│            │                                                                       │
│            │  ┌─── HYPER-GRID ────────────────────────────────────────────────┐   │
│            │  │                                                               │   │
│            │  │ Timestamp    │ User        │ Action │ Resource    │ IP      │ S│   │
│            │  │──────────────│─────────────│────────│─────────────│─────────│──│   │
│            │  │ 2 min ago    │ ● john      │ CREATE │ /users/5    │192.168.1│✓ │   │
│            │  │ 5 min ago    │ ● jane      │ UPDATE │ /roles/2    │192.168.1│✓ │   │
│            │  │ 12 min ago   │ ● alex      │ DELETE │ /users/3    │192.168.1│✓ │   │
│            │  │ 1 hour ago   │ ● system    │ LOGIN  │ /auth       │10.0.0.1 │✗ │   │
│            │  │ 2 hours ago  │ ● maya      │ EXPORT │ /audit      │192.168.1│✓ │   │
│            │  │ 3 hours ago  │ ● dev       │ READ   │ /config     │192.168.1│✓ │   │
│            │  │ 5 hours ago  │ ● admin     │ UPDATE │ /settings   │192.168.1│✓ │   │
│            │  │ Yesterday    │ ● john      │ LOGIN  │ /auth       │192.168.1│✓ │   │
│            │  │                                                               │   │
│            │  └───────────────────────────────────────────────────────────────┘   │
│            │                                                                       │
│            │  ┌─── PAGINATION ────────────────────────────────────────────────┐   │
│            │  │  Rows: [25 ▼]   Showing 1-25 of 12,847    [←] 1 2 3...514 [→]│   │
│            │  └───────────────────────────────────────────────────────────────┘   │
│            │                                                                       │
└────────────┴───────────────────────────────────────────────────────────────────────┘
```

---

## Compact Mode (Data Dense)

```
┌──────┬────────────────────────────────────────────────────────────────────────────┐
│ 72px │  📋 Audit Logs                          [📊] [📤 Export]                   │
│      │                                                                            │
│ [📊] │  [🔍 Search...] [📅 Today ▼] [👤 All ▼] [📊 All ▼] [• Compact]            │
│ [👥] │                                                                            │
│ [🔐] │  ┌────────────────────────────────────────────────────────────────────┐   │
│ [🛡️] │  │Time   │User   │Action│Resource   │IP         │Agent       │Status│   │
│ [📋] │  │───────│───────│──────│───────────│───────────│────────────│──────│   │
│ [⚙️] │  │2m     │john   │CREATE│/users/5   │192.168.1.1│Chrome/120  │ ✓    │   │
│      │  │5m     │jane   │UPDATE│/roles/2   │192.168.1.2│Firefox/119 │ ✓    │   │
│      │  │12m    │alex   │DELETE│/users/3   │192.168.1.1│Safari/17   │ ✓    │   │
│      │  │1h     │system │LOGIN │/auth      │10.0.0.1   │API Client  │ ✗    │   │
│      │  │2h     │maya   │EXPORT│/audit     │192.168.1.5│Chrome/120  │ ✓    │   │
│      │  │3h     │dev    │READ  │/config    │192.168.1.3│Postman     │ ✓    │   │
│      │  │5h     │admin  │UPDATE│/settings  │192.168.1.1│Chrome/120  │ ✓    │   │
│      │  │8h     │john   │CREATE│/roles/3   │192.168.1.1│Chrome/120  │ ✓    │   │
│      │  │12h    │jane   │LOGIN │/auth      │192.168.1.2│Firefox/119 │ ✓    │   │
│      │  │1d     │system │BACKUP│/db        │127.0.0.1  │Cron Job    │ ✓    │   │
└──────┴──│1d     │maya   │DELETE│/cache     │192.168.1.5│Script      │ ✓    │───┘
          └────────────────────────────────────────────────────────────────────┘
```

---

## Action Badge Colors

| Action | Badge Style                       | Meaning            |
| :----- | :-------------------------------- | :----------------- |
| CREATE | `bg-emerald-100 text-emerald-700` | New record created |
| READ   | `bg-slate-100 text-slate-600`     | Record viewed      |
| UPDATE | `bg-amber-100 text-amber-700`     | Record modified    |
| DELETE | `bg-red-100 text-red-700`         | Record removed     |
| LOGIN  | `bg-indigo-100 text-indigo-700`   | Auth attempt       |
| LOGOUT | `bg-slate-100 text-slate-600`     | Session ended      |
| EXPORT | `bg-violet-100 text-violet-700`   | Data exported      |
| IMPORT | `bg-blue-100 text-blue-700`       | Data imported      |

---

## Date Range Picker

```
┌─── Date Range ─────────────────────────────────────┐
│                                                    │
│  Quick Select:                                     │
│  ┌────────────────────────────────────────────┐   │
│  │ ○ Today                                    │   │
│  │ ○ Yesterday                                │   │
│  │ ○ Last 7 days                              │   │
│  │ ○ Last 30 days                             │   │
│  │ ○ This month                               │   │
│  │ ○ Last month                               │   │
│  │ ● Custom range                             │   │
│  └────────────────────────────────────────────┘   │
│                                                    │
│  From:                    To:                      │
│  ┌────────────────────┐  ┌────────────────────┐   │
│  │ Jan 15, 2026   [📅]│  │ Jan 19, 2026   [📅]│   │
│  └────────────────────┘  └────────────────────┘   │
│                                                    │
│       [Cancel]                    [Apply]          │
└────────────────────────────────────────────────────┘
```

---

## Export Options

```
┌─── Export Audit Logs ──────────────────────────────┐
│                                                    │
│  Export Format:                                    │
│  ○ CSV (Spreadsheet compatible)                   │
│  ● JSON (API/Development)                         │
│  ○ PDF Report (Formatted document)                │
│                                                    │
│  Date Range:                                       │
│  Jan 15, 2026 — Jan 19, 2026                      │
│                                                    │
│  Records to export: 1,234                          │
│                                                    │
│  Include columns:                                  │
│  ☑ Timestamp                                      │
│  ☑ User                                           │
│  ☑ Action                                         │
│  ☑ Resource                                       │
│  ☑ IP Address                                     │
│  ☑ User Agent                                     │
│  ☑ Status                                         │
│  ☑ Request Body                                   │
│  ☑ Response                                       │
│                                                    │
│  ⚠️ Large exports may take several minutes         │
│                                                    │
│       [Cancel]               [📤 Start Export]     │
└────────────────────────────────────────────────────┘
```

---

## Log Detail Modal

```
When clicking a row, show full details:

┌─── Log Entry Details ──────────────────────────────────────────┐
│                                                          [✕]   │
│                                                                │
│  ┌─── Overview ──────────────────────────────────────────────┐ │
│  │                                                           │ │
│  │  Action:     CREATE                    Status: ✓ Success  │ │
│  │  Timestamp:  Jan 19, 2026 14:32:15 UTC                    │ │
│  │  Duration:   45ms                                         │ │
│  │                                                           │ │
│  └───────────────────────────────────────────────────────────┘ │
│                                                                │
│  ┌─── Actor ─────────────────────────────────────────────────┐ │
│  │                                                           │ │
│  │  ● John Doe                                               │ │
│  │    john@example.com                                       │ │
│  │                                                           │ │
│  │  Role:       Admin                                        │ │
│  │  IP:         192.168.1.1                                  │ │
│  │  User Agent: Mozilla/5.0 Chrome/120.0.0.0                 │ │
│  │  Location:   Jakarta, Indonesia (approximate)             │ │
│  │                                                           │ │
│  └───────────────────────────────────────────────────────────┘ │
│                                                                │
│  ┌─── Request ───────────────────────────────────────────────┐ │
│  │                                                           │ │
│  │  Method:   POST                                           │ │
│  │  Endpoint: /api/v1/users                                  │ │
│  │                                                           │ │
│  │  Body:                                                    │ │
│  │  ┌─────────────────────────────────────────────────────┐ │ │
│  │  │ {                                                   │ │ │
│  │  │   "name": "Maya Chen",                              │ │ │
│  │  │   "email": "maya@example.com",                      │ │ │
│  │  │   "role": "editor"                                  │ │ │
│  │  │ }                                                   │ │ │
│  │  └─────────────────────────────────────────────────────┘ │ │
│  │                                                           │ │
│  └───────────────────────────────────────────────────────────┘ │
│                                                                │
│  ┌─── Response ──────────────────────────────────────────────┐ │
│  │                                                           │ │
│  │  Status Code: 201 Created                                 │ │
│  │                                                           │ │
│  │  Body:                                                    │ │
│  │  ┌─────────────────────────────────────────────────────┐ │ │
│  │  │ {                                                   │ │ │
│  │  │   "id": 5,                                          │ │ │
│  │  │   "name": "Maya Chen",                              │ │ │
│  │  │   "email": "maya@example.com",                      │ │ │
│  │  │   "role": "editor",                                 │ │ │
│  │  │   "created_at": "2026-01-19T14:32:15Z"              │ │ │
│  │  │ }                                                   │ │ │
│  │  └─────────────────────────────────────────────────────┘ │ │
│  │                                                           │ │
│  └───────────────────────────────────────────────────────────┘ │
│                                                                │
│                                          [📋 Copy JSON]        │
│                                                                │
└────────────────────────────────────────────────────────────────┘
```

---

## Analytics View (Optional Tab)

```
┌─────────────────────────────────────────────────────────────────────────┐
│  📊 Audit Analytics                                                     │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  ┌─── Activity Over Time (7 days) ────────────────────────────────────┐ │
│  │                                                                    │ │
│  │  1200 ─┤                              ┌──────────────────────────┐│ │
│  │        │                              │ ─ Total Actions          ││ │
│  │   800 ─┤        ╭─╮                   │ ─ Successful             ││ │
│  │        │   ╭───╮│ │╭──╮              │ ─ Failed                 ││ │
│  │   400 ─┤  ╭╯   ╰╯ ╰╯  ╰╮             └──────────────────────────┘│ │
│  │        │ ╭╯             ╰─────────╮                               │ │
│  │     0 ─┴─────────────────────────────                             │ │
│  │        Mon  Tue  Wed  Thu  Fri  Sat  Sun                          │ │
│  │                                                                    │ │
│  └────────────────────────────────────────────────────────────────────┘ │
│                                                                         │
│  ┌─── By Action ──────────┐  ┌─── By User ────────────────────────────┐│
│  │                        │  │                                        ││
│  │  CREATE  ████████ 45%  │  │  1. john@example.com     456 actions   ││
│  │  READ    ██████   30%  │  │  2. jane@example.com     234 actions   ││
│  │  UPDATE  ████     15%  │  │  3. system               189 actions   ││
│  │  DELETE  ██        7%  │  │  4. alex@example.com      67 actions   ││
│  │  OTHER   █         3%  │  │  5. maya@example.com      45 actions   ││
│  │                        │  │                                        ││
│  └────────────────────────┘  └────────────────────────────────────────┘│
│                                                                         │
│  ┌─── Failed Actions (Alerts) ────────────────────────────────────────┐│
│  │                                                                    ││
│  │  ⚠️ 3 failed login attempts from IP 10.0.0.1 in last hour          ││
│  │  ⚠️ 1 unauthorized access attempt to /admin/config                 ││
│  │                                                                    ││
│  └────────────────────────────────────────────────────────────────────┘│
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## Zebra Striping (Dark Mode)

```
Dark Mode Table:

│ Time   │ User   │ Action │ Resource    │ Status │
│────────│────────│────────│─────────────│────────│  ← Header: bg-slate-900
│ 2m     │ john   │ CREATE │ /users/5    │ ✓      │  ← Row Even: transparent
│ 5m     │ jane   │ UPDATE │ /roles/2    │ ✓      │  ← Row Odd: bg-slate-900/50
│ 12m    │ alex   │ DELETE │ /users/3    │ ✓      │  ← Row Even: transparent
│ 1h     │ system │ LOGIN  │ /auth       │ ✗      │  ← Row Odd: bg-slate-900/50
│ 2h     │ maya   │ EXPORT │ /audit      │ ✓      │  ← Row Even: transparent

Hover: bg-indigo-500/10 (subtle indigo tint)
Selected: bg-indigo-500/20 + left border 2px indigo
```

---

## Real-Time Streaming Mode

Live audit log streaming for real-time monitoring and security operations.

### Streaming Indicator (Toolbar)

```
Normal Mode (Paginated):
┌─────────────────────────────────────────────────────────────────────────────┐
│  [Columns ▼]   [◐ Comfort ● Compact]   [📡 Live: Off]      [📤 Export ▼]  │
│                                              ↑                              │
│                                    Toggle to enable streaming               │
└─────────────────────────────────────────────────────────────────────────────┘

Streaming Mode (Live):
┌─────────────────────────────────────────────────────────────────────────────┐
│  [Columns ▼]   [◐ Comfort ● Compact]   [📡 Live ◉]  12/s   [📤 Export ▼]  │
│                                            ↑        ↑                       │
│                                         Pulsing    Events per second        │
│                                         green dot  (live rate counter)      │
└─────────────────────────────────────────────────────────────────────────────┘
```

### New Event Animation

```
When new event arrives in streaming mode:

┌──────────────────────────────────────────────────────────────────────────┐
│ Timestamp    │ User        │ Action │ Resource    │ IP        │ Status  │
│──────────────│─────────────│────────│─────────────│───────────│─────────│
│ just now     │ ● maya      │ CREATE │ /users/7    │192.168.1.5│ ✓       │ ← NEW
│ 2 sec ago    │ ● john      │ UPDATE │ /roles/2    │192.168.1.1│ ✓       │ ← fade-in
│ 5 sec ago    │ ● system    │ LOGIN  │ /auth       │10.0.0.1   │ ✗       │
│ 12 sec ago   │ ● alex      │ DELETE │ /users/3    │192.168.1.1│ ✓       │
│ 30 sec ago   │ ● jane      │ EXPORT │ /audit      │192.168.1.2│ ✓       │

Animation Sequence:
1. New row slides in from top with bg-indigo-500/20 highlight
2. Highlight fades to transparent over 2 seconds
3. "just now" timestamp auto-updates (2 sec ago, 5 sec ago...)
4. Older rows scroll down (max 100 visible in streaming mode)
```

### Streaming Controls

```
┌─── Live Streaming Settings ─────────────────────────────────────┐
│                                                                 │
│  Streaming Status: 🟢 Connected                                 │
│  Events received: 1,234 (this session)                          │
│  Buffer size: 100 rows (newest visible)                         │
│                                                                 │
│  ─────────────────────────────────────────────────────────────  │
│                                                                 │
│  Filter Streaming Events:                                       │
│  ☑ Show all events                                             │
│  ☐ Only show errors/failures                                   │
│  ☐ Only show specific actions: [________]                      │
│                                                                 │
│  Auto-Pause Streaming:                                          │
│  ○ Never                                                       │
│  ● After 5 minutes of inactivity                               │
│  ○ After 1000 events                                           │
│                                                                 │
│  Sound Notification:                                            │
│  ☐ Play sound on new errors                                    │
│  ☐ Play sound on security events (LOGIN failures)              │
│                                                                 │
│                                    [Done]                       │
└─────────────────────────────────────────────────────────────────┘
```

### Connection States

```
Connected:
│  [📡 Live ◉]  45/s   │   ← Green pulsing dot, showing events/sec

Reconnecting:
│  [📡 Reconnecting...] │   ← Yellow, animated dots

Disconnected:
│  [📡 Offline ⚠️]  [Retry] │   ← Red with retry button

Paused:
│  [📡 Paused ⏸️]  [Resume] │   ← Gray with resume button
```

### Security Alert Highlight

```
When a security-relevant event arrives (e.g., failed login):

┌──────────────────────────────────────────────────────────────────────────┐
│ ⚠️ just now  │ ● unknown  │ LOGIN  │ /auth       │10.0.0.1   │ ⚠️ FAILED│
└──────────────────────────────────────────────────────────────────────────┘
   ↑                                                                     ↑
Border left                                                        Alert icon
4px red-500                                                        + red badge

Toast Notification (optional):
┌─────────────────────────────────────────────────────────────┐
│ ⚠️ Security Alert                                     [✕]   │
│    Failed login attempt from IP 10.0.0.1                    │
│    [View Details]  [Block IP]                               │
└─────────────────────────────────────────────────────────────┘
```
