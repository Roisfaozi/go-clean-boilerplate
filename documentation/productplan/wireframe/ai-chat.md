# AI Chat Wireframe (Command Center)

## Overview

Dockable AI assistant that provides context-aware assistance throughout the application.

---

## Display Modes

### Mode 1: Floating Widget (Default)

```
                                        ┌─────────────────────────────────────────┐
                                        │ 🤖 NexusAI                        ─  ✕ │
                                        ├─────────────────────────────────────────┤
                                        │                                         │
                                        │  ┌─────────────────────────────────┐   │
                                        │  │ Hi! I'm NexusAI, your admin     │   │
                                        │  │ assistant. How can I help you?  │   │
                                        │  └─────────────────────────────────┘   │
                                        │                            AI Bubble   │
                                        │                                         │
                                        │  ┌─────────────────────────────────┐   │
                                        │  │ Show me users who logged in     │   │
                                        │  │ from a new IP this week         │   │
                                        │  └─────────────────────────────────┘   │
                                        │                          User Bubble   │
                                        │                                         │
                                        │  ┌─────────────────────────────────┐   │
                                        │  │ 🤖 ● ● ● Thinking...            │   │
                                        │  │   ✨ Shimmer animation          │   │
                                        │  └─────────────────────────────────┘   │
                                        │                                         │
                                        ├─────────────────────────────────────────┤
                                        │  ┌─────────────────────────────────┐   │
                                        │  │ Ask NexusAI anything...    [📎][→]│   │
                                        │  └─────────────────────────────────┘   │
                                        └─────────────────────────────────────────┘
                                                            ↑
                                              Position: fixed bottom-6 right-6
                                              Size: 380px wide × 500px tall
```

### Mode 2: Split View (Co-Pilot Mode)

```
┌────────────────────────────────────────────────────────────────────────────────────┐
│ NAVBAR                                                        [◐] [🌙] [👤]       │
├────────────┬───────────────────────────────────────────┬───────────────────────────┤
│            │                                           │                           │
│  SIDEBAR   │         MAIN CONTENT (70%)               │    AI PANEL (30%)         │
│            │                                           │                           │
│            │  ┌─────────────────────────────────────┐ │  ┌───────────────────────┐ │
│            │  │                                     │ │  │ 🤖 NexusAI       [⊡][✕]│ │
│            │  │                                     │ │  ├───────────────────────┤ │
│            │  │      Current Page Content           │ │  │                       │ │
│            │  │      (Dashboard / Users / etc)      │ │  │ Analyzing the current │ │
│            │  │                                     │ │  │ Audit Logs data...    │ │
│            │  │                                     │ │  │                       │ │
│            │  │                                     │ │  │ I found 3 suspicious  │ │
│            │  │                                     │ │  │ login attempts:       │ │
│            │  │                                     │ │  │                       │ │
│            │  │                                     │ │  │ • IP: 10.0.0.1        │ │
│            │  │                                     │ │  │   5 failed logins     │ │
│            │  │                                     │ │  │ • IP: 192.168.50.1    │ │
│            │  │                                     │ │  │   3 failed logins     │ │
│            │  │                                     │ │  │                       │ │
│            │  │                                     │ │  │ [Show Details]        │ │
│            │  │                                     │ │  │ [Block These IPs]     │ │
│            │  │                                     │ │  │                       │ │
│            │  │                                     │ │  ├───────────────────────┤ │
│            │  │                                     │ │  │ [Ask follow-up... ][→]│ │
│            │  └─────────────────────────────────────┘ │  └───────────────────────┘ │
│            │                                           │                           │
└────────────┴───────────────────────────────────────────┴───────────────────────────┘

Transition: grid-template-columns animates from "1fr" to "1fr 350px" (300ms ease)
Trigger: Click "Dock" button [⊡] in floating widget or keyboard shortcut Ctrl+/
```

---

## Widget Header States

```
Idle State:
┌─────────────────────────────────────────┐
│ 🤖 NexusAI              [⊡] [─] [✕]    │  ⊡ = Dock to side
└─────────────────────────────────────────┘     ─ = Minimize
                                                ✕ = Close

Processing State:
┌─────────────────────────────────────────┐
│ 🤖 NexusAI  ● Thinking...  [⊡] [─] [✕] │  ● = Pulsing indigo dot
└─────────────────────────────────────────┘

Streaming State:
┌─────────────────────────────────────────┐
│ 🤖 NexusAI  ◉ Writing...  [⊡] [─] [✕]  │  ◉ = Animated streaming
└─────────────────────────────────────────┘

Error State:
┌─────────────────────────────────────────┐
│ 🤖 NexusAI  ⚠️ Offline     [⊡] [─] [✕] │  Try reconnecting
└─────────────────────────────────────────┘
```

---

## Chat Bubbles

### User Bubble

```
                            ┌─────────────────────────────────┐
                            │ Show me all users with the      │
                            │ "admin" role                    │
                            └─────────────────────────────────┘
                                  ↑
                      Align: Right
                      Background: bg-slate-100 (light) / bg-slate-800 (dark)
                      Radius: rounded-lg rounded-tr-sm
                      Max-width: 85%
```

### AI Bubble

```
┌─────────────────────────────────────────────────────────────┐
│ 🤖                                                         │
│                                                             │
│ I found **5 users** with the admin role:                   │
│                                                             │
│ | Name       | Email              | Last Active |          │
│ |------------|--------------------| ------------|          │
│ | John Doe   | john@example.com   | 2h ago      |          │
│ | Jane Smith | jane@example.com   | 5h ago      |          │
│ | ...        | ...                | ...         |          │
│                                                             │
│ Would you like me to:                                       │
│ • [View all admins in Users page]                          │
│ • [Export this list]                                        │
│ • [Check their recent activity]                            │
│                                                             │
└─────────────────────────────────────────────────────────────┘
     ↑
Align: Left
Background: bg-white border border-indigo-100 (light)
            bg-slate-900 border border-indigo-500/20 (dark)
Radius: rounded-lg rounded-tl-sm
Supports: Full Markdown rendering (tables, code, lists)
Action Buttons: Clickable quick actions
```

### AI Thinking State

```
┌─────────────────────────────────────────────────────────────┐
│ 🤖                                                         │
│                                                             │
│   ╭──────────────────────────────────────────────────────╮ │
│   │  ✨ ✨ ✨ ✨ ✨ ✨ ✨ ✨ ✨ ✨ ✨ ✨ ✨ ✨ ✨ ✨          │ │
│   │                 Shimmer animation                    │ │
│   │                 (violet gradient)                    │ │
│   ╰──────────────────────────────────────────────────────╯ │
│                                                             │
└─────────────────────────────────────────────────────────────┘

Animation: background gradient moves left to right
Colors: from-violet-500/5 via-fuchsia-500/10 to-violet-500/5
Duration: 1.5s infinite
```

---

## Input Area

```
┌───────────────────────────────────────────────────────────────┐
│                                                               │
│  ┌───────────────────────────────────────────────────────┐   │
│  │                                                       │   │
│  │ Ask NexusAI anything...                           [📎]│   │
│  │                                                       │   │
│  │                                             [↵ Send] │   │
│  └───────────────────────────────────────────────────────┘   │
│                                                               │
│  [📊 Analyze this page] [❓ Help] [⌨️ Shortcuts]              │
│        ↑ Quick action chips                                   │
└───────────────────────────────────────────────────────────────┘

Input Features:
- Auto-grow textarea (1-5 lines)
- 📎 = Attach context (current page data, selected rows, filters)
- Enter = Send | Shift+Enter = New line
- Quick action chips for common tasks
```

---

## Context Attachment Modal

```
When clicking 📎 (Attach Context):

┌─── Attach Context ────────────────────────────────────────────┐
│                                                         [✕]   │
│                                                               │
│  Include in conversation:                                     │
│                                                               │
│  ☑ Current page data (Audit Logs)                            │
│      → 1,234 filtered records                                │
│                                                               │
│  ☐ Selected rows                                             │
│      → No rows selected                                      │
│                                                               │
│  ☑ Active filters                                            │
│      → Date: Today, Action: CREATE, Status: Failed           │
│                                                               │
│  ☐ Current user context                                      │
│      → john@example.com (Admin role)                         │
│                                                               │
│  ┌─────────────────────────────────────────────────────────┐ │
│  │ Preview:                                                 │ │
│  │ "Analyzing 1,234 audit log entries filtered by:         │ │
│  │  Today's CREATE actions with Failed status..."          │ │
│  └─────────────────────────────────────────────────────────┘ │
│                                                               │
│                                      [Cancel]  [Attach]       │
│                                                               │
└───────────────────────────────────────────────────────────────┘
```

---

## Example Conversations

### Scenario 1: Data Analysis

```
User: "Why are there so many failed logins today?"

AI: "I analyzed today's audit logs and found **23 failed login attempts**.

Here's the breakdown:

| Reason          | Count | % of Total |
|-----------------|-------|------------|
| Wrong password  | 15    | 65%        |
| Account locked  | 5     | 22%        |
| 2FA failed      | 3     | 13%        |

⚠️ **Alert**: 12 attempts came from IP `10.0.0.1` which is not in your
trusted IP list.

[View Full Report] [Block Suspicious IP] [Notify Security Team]"
```

### Scenario 2: Quick Actions

```
User: "Add a new user with editor role"

AI: "I'll help you create a new user. Please provide:

**Required Information:**
- Full Name: [____________]
- Email: [____________]
- ☑ Role: Editor (pre-selected)
- ☐ Send welcome email

[Create User] [Cancel]

Or I can open the full user creation form for you:
[Open User Form →]"
```

### Scenario 3: Help & Guidance

```
User: "How do I export audit logs?"

AI: "To export audit logs:

1. Go to **Audit Logs** page
2. Apply any filters you need (date, user, action)
3. Click **📤 Export** button in the toolbar
4. Choose format: CSV, JSON, or PDF Report
5. Select columns to include
6. Click **Start Export**

💡 **Tip**: For large exports (>10,000 records), we'll email you the
download link when ready.

[Go to Audit Logs] [Watch Tutorial Video]"
```

---

## Minimized State (Fab Button)

```
When minimized, show as floating action button:

                                                          ┌───────┐
                                                          │  🤖   │
                                                          │  (1)  │  ← Unread count badge
                                                          └───────┘
                                                              ↑
                                            Position: fixed bottom-6 right-6
                                            Size: 56px circle
                                            Shadow: shadow-lg
                                            Hover: scale-110 + pulse glow
```

---

## Keyboard Shortcuts

| Shortcut             | Action                      |
| :------------------- | :-------------------------- |
| `Ctrl/⌘ + /`         | Toggle AI Chat              |
| `Ctrl/⌘ + Shift + /` | Toggle Dock mode            |
| `Escape`             | Close/Minimize              |
| `Enter`              | Send message                |
| `Shift + Enter`      | New line in input           |
| `Ctrl/⌘ + K`         | Focus input (from anywhere) |

---

## Mobile Behavior

```
On mobile (< 768px), AI Chat becomes full-screen sheet:

┌──────────────────────────┐
│ 🤖 NexusAI         [✕]  │
├──────────────────────────┤
│                          │
│  Chat history            │
│  scrollable area         │
│                          │
│                          │
│                          │
│                          │
│                          │
│                          │
├──────────────────────────┤
│ [Ask anything...]   [→]  │
│                          │
│ [📊 Page] [❓] [⌨️]      │
└──────────────────────────┘

Trigger: Bottom nav AI icon or FAB
Animation: Slide up from bottom
```

---

## Density Mode Variations

### Comfort vs Compact Comparison

| Element            | Comfort Mode | Compact Mode |
| :----------------- | :----------- | :----------- |
| **Widget Width**   | 380px        | 320px        |
| **Widget Height**  | 500px        | 400px        |
| **Bubble Padding** | 16px         | 10px         |
| **Font Size**      | 14px         | 13px         |
| **Input Height**   | 48px         | 36px         |
| **Quick Chips**    | Full text    | Icons only   |
| **Split Panel**    | 350px        | 280px        |

### Compact Mode Floating Widget

```
                                ┌───────────────────────────────────┐
                                │ 🤖 NexusAI           [⊡] [─] [✕] │
                                ├───────────────────────────────────┤
                                │                                   │
                                │ ┌───────────────────────────────┐│
                                ││ Hi! How can I help you today? ││  ← Smaller bubbles
                                │ └───────────────────────────────┘│
                                │                                   │
                                │            ┌────────────────────┐│
                                │            │ Show failed logins ││
                                │            └────────────────────┘│
                                │                                   │
                                │ ┌───────────────────────────────┐│
                                ││ Found 23 failed attempts...   ││
                                │ └───────────────────────────────┘│
                                │                                   │
                                ├───────────────────────────────────┤
                                │ [Ask anything...         ] [📎][→]│
                                │ [📊] [❓] [⌨️]                     │  ← Icon-only chips
                                └───────────────────────────────────┘
                                           320px × 400px

Key Differences:
- Smaller widget dimensions
- Condensed bubble padding
- Icon-only quick action chips (📊 = Analyze, ❓ = Help, ⌨️ = Shortcuts)
- Single line input area
```

### Compact Mode Split View

```
┌────────────────────────────────────────────────────────────────────────────────────┐
│ NAVBAR                                                       [◐ Compact] [🌙] [👤] │
├────────────┬────────────────────────────────────────────────┬──────────────────────┤
│            │                                                │                      │
│  SIDEBAR   │           MAIN CONTENT (75%)                  │   AI PANEL (25%)     │
│            │                                                │       280px          │
│            │  ┌──────────────────────────────────────────┐ │  ┌──────────────────┐│
│            │  │                                          │ │  │🤖 NexusAI  [⊡][✕]││
│            │  │                                          │ │  ├──────────────────┤│
│            │  │                                          │ │  │Found 3 issues:   ││
│            │  │            Page Content                  │ │  │• IP: 10.0.0.1    ││
│            │  │                                          │ │  │• IP: 192.168.1.1 ││
│            │  │                                          │ │  │[Details][Block]  ││
│            │  │                                          │ │  ├──────────────────┤│
│            │  │                                          │ │  │[Ask...    ][📎][→]││
│            │  └──────────────────────────────────────────┘ │  └──────────────────┘│
│            │                                                │                      │
└────────────┴────────────────────────────────────────────────┴──────────────────────┘

Key Differences:
- Narrower panel (280px vs 350px)
- Compact bubbles and buttons
- More screen real estate for main content
```
