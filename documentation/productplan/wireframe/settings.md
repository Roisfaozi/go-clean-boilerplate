# Settings & Preferences Wireframe

## Overview

User settings page for profile, preferences, density mode, theme, and API keys.

---

## Settings Page Layout

```
┌────────────────────────────────────────────────────────────────────────────────────┐
│ NAVBAR                                                                             │
├────────────┬───────────────────────────────────────────────────────────────────────┤
│            │                                                                       │
│  SIDEBAR   │  ┌─── PAGE HEADER ───────────────────────────────────────────────┐   │
│            │  │                                                               │   │
│            │  │  ⚙️ Settings                                                  │   │
│  ● Settings│  │  Manage your account and preferences                         │   │
│            │  │                                                               │   │
│            │  └───────────────────────────────────────────────────────────────┘   │
│            │                                                                       │
│            │  ┌─── SETTINGS TABS ─────────────────────────────────────────────┐   │
│            │  │                                                               │   │
│            │  │  [Profile]  [Preferences]  [Security]  [API Keys]  [Billing]  │   │
│            │  │      ●                                                        │   │
│            │  └───────────────────────────────────────────────────────────────┘   │
│            │                                                                       │
│            │  ┌─── CONTENT AREA ──────────────────────────────────────────────┐   │
│            │  │                                                               │   │
│            │  │                    (Tab Content)                              │   │
│            │  │                                                               │   │
│            │  └───────────────────────────────────────────────────────────────┘   │
│            │                                                                       │
└────────────┴───────────────────────────────────────────────────────────────────────┘
```

---

## Tab 1: Profile

```
┌─── Profile Settings ───────────────────────────────────────────────────────────────┐
│                                                                                    │
│  ┌─── Profile Picture ─────────────────────────────────────────────────────────┐  │
│  │                                                                             │  │
│  │     ┌───────────┐                                                           │  │
│  │     │           │                                                           │  │
│  │     │    JD     │      John Doe                                            │  │
│  │     │           │      john@example.com                                    │  │
│  │     └───────────┘      Member since Jan 2026                               │  │
│  │         80px                                                                │  │
│  │                                                                             │  │
│  │     [📷 Change Photo]  [🗑️ Remove]                                          │  │
│  │                                                                             │  │
│  └─────────────────────────────────────────────────────────────────────────────┘  │
│                                                                                    │
│  ┌─── Personal Information ────────────────────────────────────────────────────┐  │
│  │                                                                             │  │
│  │  Full Name                          Email                                   │  │
│  │  ┌─────────────────────────┐       ┌─────────────────────────┐             │  │
│  │  │ John Doe                │       │ john@example.com    🔒  │             │  │
│  │  └─────────────────────────┘       └─────────────────────────┘             │  │
│  │                                    ℹ️ Contact admin to change email         │  │
│  │                                                                             │  │
│  │  Phone (optional)                   Timezone                                │  │
│  │  ┌─────────────────────────┐       ┌─────────────────────────┐             │  │
│  │  │ +62 812 3456 7890       │       │ Asia/Jakarta (UTC+7) ▼  │             │  │
│  │  └─────────────────────────┘       └─────────────────────────┘             │  │
│  │                                                                             │  │
│  │  Bio (optional)                                                             │  │
│  │  ┌───────────────────────────────────────────────────────────────────────┐ │  │
│  │  │ System administrator at Example Corp.                                 │ │  │
│  │  │                                                                       │ │  │
│  │  └───────────────────────────────────────────────────────────────────────┘ │  │
│  │                                                                             │  │
│  │                                                         [Save Changes]     │  │
│  │                                                                             │  │
│  └─────────────────────────────────────────────────────────────────────────────┘  │
│                                                                                    │
│  ┌─── Danger Zone ─────────────────────────────────────────────────────────────┐  │
│  │                                                                             │  │
│  │  Delete Account                                                             │  │
│  │  Permanently delete your account and all associated data.                  │  │
│  │                                                                             │  │
│  │  [🗑️ Delete My Account]                                                     │  │
│  │                                                                             │  │
│  └─────────────────────────────────────────────────────────────────────────────┘  │
│                                                                                    │
└────────────────────────────────────────────────────────────────────────────────────┘
```

---

## Tab 2: Preferences

```
┌─── Preferences ────────────────────────────────────────────────────────────────────┐
│                                                                                    │
│  ┌─── Appearance ──────────────────────────────────────────────────────────────┐  │
│  │                                                                             │  │
│  │  Theme                                                                      │  │
│  │  ┌───────────────────────────────────────────────────────────────────────┐ │  │
│  │  │                                                                       │ │  │
│  │  │  ┌─────────┐    ┌─────────┐    ┌─────────┐                           │ │  │
│  │  │  │  ☀️     │    │  🌙     │    │  💻     │                           │ │  │
│  │  │  │  Light  │    │  Dark   │    │  System │                           │ │  │
│  │  │  │   ●     │    │         │    │         │                           │ │  │
│  │  │  └─────────┘    └─────────┘    └─────────┘                           │ │  │
│  │  │                                                                       │ │  │
│  │  └───────────────────────────────────────────────────────────────────────┘ │  │
│  │                                                                             │  │
│  │  Density Mode (Chameleon Engine)                                            │  │
│  │  ┌───────────────────────────────────────────────────────────────────────┐ │  │
│  │  │                                                                       │ │  │
│  │  │  ┌───────────────────┐    ┌───────────────────┐                      │ │  │
│  │  │  │                   │    │                   │                      │ │  │
│  │  │  │  ◐ Comfort        │    │  ◐ Compact        │                      │ │  │
│  │  │  │                   │    │                   │                      │ │  │
│  │  │  │  Spacious layout  │    │  Dense data view  │                      │ │  │
│  │  │  │  Larger fonts     │    │  Smaller fonts    │                      │ │  │
│  │  │  │  More whitespace  │    │  Less whitespace  │                      │ │  │
│  │  │  │       ●           │    │                   │                      │ │  │
│  │  │  └───────────────────┘    └───────────────────┘                      │ │  │
│  │  │                                                                       │ │  │
│  │  │  💡 Tip: Use Ctrl+D to quickly toggle density mode                   │ │  │
│  │  │                                                                       │ │  │
│  │  └───────────────────────────────────────────────────────────────────────┘ │  │
│  │                                                                             │  │
│  │  Sidebar                                                                    │  │
│  │  ○ Expanded (280px)  ● Collapsed (72px)  ○ Auto (responsive)              │  │
│  │                                                                             │  │
│  └─────────────────────────────────────────────────────────────────────────────┘  │
│                                                                                    │
│  ┌─── Notifications ───────────────────────────────────────────────────────────┐  │
│  │                                                                             │  │
│  │  Email Notifications                                                        │  │
│  │                                                                             │  │
│  │  ☑ Security alerts (login from new device, password changes)              │  │
│  │  ☑ Weekly activity summary                                                 │  │
│  │  ☐ Product updates and announcements                                       │  │
│  │  ☐ Tips and tutorials                                                      │  │
│  │                                                                             │  │
│  │  In-App Notifications                                                       │  │
│  │                                                                             │  │
│  │  ☑ Show desktop notifications                        [Test 🔔]            │  │
│  │  ☑ Play sound for alerts                                                  │  │
│  │                                                                             │  │
│  └─────────────────────────────────────────────────────────────────────────────┘  │
│                                                                                    │
│  ┌─── Language & Region ───────────────────────────────────────────────────────┐  │
│  │                                                                             │  │
│  │  Language                           Date Format                             │  │
│  │  ┌─────────────────────────┐       ┌─────────────────────────┐             │  │
│  │  │ English (US)          ▼│       │ DD/MM/YYYY           ▼  │             │  │
│  │  └─────────────────────────┘       └─────────────────────────┘             │  │
│  │                                                                             │  │
│  │  Number Format                      First day of week                       │  │
│  │  ┌─────────────────────────┐       ┌─────────────────────────┐             │  │
│  │  │ 1,234,567.89          ▼│       │ Monday               ▼  │             │  │
│  │  └─────────────────────────┘       └─────────────────────────┘             │  │
│  │                                                                             │  │
│  └─────────────────────────────────────────────────────────────────────────────┘  │
│                                                                                    │
│                                                                   [Save Changes]  │
│                                                                                    │
└────────────────────────────────────────────────────────────────────────────────────┘
```

---

## Tab 3: Security

```
┌─── Security Settings ──────────────────────────────────────────────────────────────┐
│                                                                                    │
│  ┌─── Password ────────────────────────────────────────────────────────────────┐  │
│  │                                                                             │  │
│  │  Password                                                                   │  │
│  │  Last changed 30 days ago                                                   │  │
│  │                                                                             │  │
│  │  [🔐 Change Password]                                                       │  │
│  │                                                                             │  │
│  └─────────────────────────────────────────────────────────────────────────────┘  │
│                                                                                    │
│  ┌─── Two-Factor Authentication ───────────────────────────────────────────────┐  │
│  │                                                                             │  │
│  │  Status: 🟢 Enabled                                                         │  │
│  │                                                                             │  │
│  │  Authenticator App: Google Authenticator                                    │  │
│  │  Added: Jan 15, 2026                                                        │  │
│  │                                                                             │  │
│  │  [🔄 Regenerate Backup Codes]  [🗑️ Disable 2FA]                             │  │
│  │                                                                             │  │
│  └─────────────────────────────────────────────────────────────────────────────┘  │
│                                                                                    │
│  ┌─── Active Sessions ─────────────────────────────────────────────────────────┐  │
│  │                                                                             │  │
│  │  ● This device                                                              │  │
│  │    Chrome on Windows • Jakarta, Indonesia                                   │  │
│  │    Active now                                                    [Current] │  │
│  │                                                                             │  │
│  │  ○ iPhone 15 Pro                                                           │  │
│  │    Safari on iOS • Jakarta, Indonesia                                       │  │
│  │    Last active 2 hours ago                                      [Revoke]   │  │
│  │                                                                             │  │
│  │  ○ MacBook Pro                                                              │  │
│  │    Chrome on macOS • Singapore                                              │  │
│  │    Last active 3 days ago                                       [Revoke]   │  │
│  │                                                                             │  │
│  │                                            [Revoke All Other Sessions]     │  │
│  │                                                                             │  │
│  └─────────────────────────────────────────────────────────────────────────────┘  │
│                                                                                    │
│  ┌─── Login History ───────────────────────────────────────────────────────────┐  │
│  │                                                                             │  │
│  │  Jan 19, 2026 14:32  •  Chrome on Windows  •  Jakarta, ID  •  ✓ Success   │  │
│  │  Jan 19, 2026 09:15  •  Safari on iOS      •  Jakarta, ID  •  ✓ Success   │  │
│  │  Jan 18, 2026 22:45  •  Chrome on macOS    •  Singapore    •  ✓ Success   │  │
│  │  Jan 18, 2026 18:30  •  Unknown            •  10.0.0.1     •  ✗ Failed    │  │
│  │                                                                             │  │
│  │                                                  [View Full History →]     │  │
│  │                                                                             │  │
│  └─────────────────────────────────────────────────────────────────────────────┘  │
│                                                                                    │
└────────────────────────────────────────────────────────────────────────────────────┘
```

---

## Tab 4: API Keys

```
┌─── API Keys ───────────────────────────────────────────────────────────────────────┐
│                                                                                    │
│  ┌─── Info Banner ─────────────────────────────────────────────────────────────┐  │
│  │  ℹ️ API keys allow you to authenticate with the NexusOS API.                 │  │
│  │     Keep your keys secure and never share them publicly.                    │  │
│  │     [📖 View API Documentation]                                              │  │
│  └─────────────────────────────────────────────────────────────────────────────┘  │
│                                                                                    │
│  ┌─── Your API Keys ───────────────────────────────────────────────────────────┐  │
│  │                                                                             │  │
│  │  ┌─────────────────────────────────────────────────────────────────────┐   │  │
│  │  │                                                                     │   │  │
│  │  │  Production Key                                        🟢 Active    │   │  │
│  │  │                                                                     │   │  │
│  │  │  Key:  nxos_prod_sk_●●●●●●●●●●●●●●●●●●●●●●●●abc123        [📋] [👁] │   │  │
│  │  │                                                                     │   │  │
│  │  │  Created: Jan 10, 2026                                              │   │  │
│  │  │  Last used: 2 hours ago                                             │   │  │
│  │  │  Permissions: Read, Write                                           │   │  │
│  │  │                                                                     │   │  │
│  │  │  [⚙️ Edit Permissions]  [🔄 Regenerate]  [🗑️ Revoke]                │   │  │
│  │  │                                                                     │   │  │
│  │  └─────────────────────────────────────────────────────────────────────┘   │  │
│  │                                                                             │  │
│  │  ┌─────────────────────────────────────────────────────────────────────┐   │  │
│  │  │                                                                     │   │  │
│  │  │  Development Key                                       🟡 Limited   │   │  │
│  │  │                                                                     │   │  │
│  │  │  Key:  nxos_dev_sk_●●●●●●●●●●●●●●●●●●●●●●●●def456         [📋] [👁] │   │  │
│  │  │                                                                     │   │  │
│  │  │  Created: Jan 5, 2026                                               │   │  │
│  │  │  Last used: 5 days ago                                              │   │  │
│  │  │  Permissions: Read only                                             │   │  │
│  │  │                                                                     │   │  │
│  │  │  [⚙️ Edit Permissions]  [🔄 Regenerate]  [🗑️ Revoke]                │   │  │
│  │  │                                                                     │   │  │
│  │  └─────────────────────────────────────────────────────────────────────┘   │  │
│  │                                                                             │  │
│  │                                                    [+ Create New API Key]  │  │
│  │                                                                             │  │
│  └─────────────────────────────────────────────────────────────────────────────┘  │
│                                                                                    │
└────────────────────────────────────────────────────────────────────────────────────┘
```

---

## Create API Key Modal

```
┌─── Create New API Key ─────────────────────────────────────────┐
│                                                          [✕]   │
│                                                                │
│  Key Name *                                                    │
│  ┌──────────────────────────────────────────────────────────┐ │
│  │ My Integration Key                                       │ │
│  └──────────────────────────────────────────────────────────┘ │
│                                                                │
│  Environment                                                   │
│  ○ Production (Full access)                                   │
│  ● Development (Rate limited, sandbox data)                   │
│                                                                │
│  Permissions                                                   │
│  ┌──────────────────────────────────────────────────────────┐ │
│  │ ☑ Users                  ☑ Read  ☑ Write               │ │
│  │ ☑ Roles                  ☑ Read  ☐ Write               │ │
│  │ ☐ Audit Logs             ☐ Read  ☐ Write               │ │
│  │ ☑ Config                 ☑ Read  ☐ Write               │ │
│  └──────────────────────────────────────────────────────────┘ │
│                                                                │
│  Expiration                                                    │
│  ○ Never expires                                              │
│  ● Expires in: [90 days ▼]                                    │
│                                                                │
│  IP Whitelist (optional)                                       │
│  ┌──────────────────────────────────────────────────────────┐ │
│  │ 192.168.1.0/24, 10.0.0.1                                 │ │
│  └──────────────────────────────────────────────────────────┘ │
│  ℹ️ Leave empty to allow all IPs                               │
│                                                                │
│  ┌──────────────────────────────────────────────────────────┐ │
│  │  [Cancel]                               [Create Key]     │ │
│  └──────────────────────────────────────────────────────────┘ │
│                                                                │
└────────────────────────────────────────────────────────────────┘
```

---

## Key Created Success Modal

```
┌─── API Key Created ────────────────────────────────────────────┐
│                                                                │
│                        ✓                                       │
│                                                                │
│              Your API key has been created!                    │
│                                                                │
│  ⚠️ Make sure to copy your key now.                            │
│     You won't be able to see it again.                        │
│                                                                │
│  ┌──────────────────────────────────────────────────────────┐ │
│  │                                                          │ │
│  │  nxos_dev_sk_a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6q7r8s9t0   │ │
│  │                                                          │ │
│  │                                              [📋 Copy]   │ │
│  └──────────────────────────────────────────────────────────┘ │
│                                                                │
│                                               [Done]           │
│                                                                │
└────────────────────────────────────────────────────────────────┘
```

---

## Density Mode Variations

### Comfort vs Compact Comparison

| Element             | Comfort Mode | Compact Mode |
| :------------------ | :----------- | :----------- |
| **Section Padding** | 24px         | 16px         |
| **Label Position**  | Above inputs | Inline left  |
| **Input Height**    | 44px         | 36px         |
| **Card Spacing**    | 24px gap     | 12px gap     |
| **Font Size**       | 14px         | 13px         |
| **Avatar Size**     | 80px         | 48px         |

### Compact Mode Profile Tab

```
┌─── Profile ─────────────────────────────────────────────────────────────────┐
│                                                                             │
│  ┌──────┐  Name  [John Doe____________]   Email [john@example.com 🔒]      │
│  │  JD  │  Phone [+62 812 3456 7890___]   TZ    [Asia/Jakarta (UTC+7) ▼]   │
│  └──────┘                                                                   │
│   48px     Bio   [System administrator at Example Corp.________________]   │
│                                                                             │
│  [📷 Photo] [🗑️]                                           [Save Changes]  │
│                                                                             │
│  ───────────────────────────────────────────────────────────────────────── │
│  ⚠️ Danger: [🗑️ Delete Account]                                             │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘

Key Differences:
- Inline labels (Name, Email, Phone, TZ, Bio)
- Smaller avatar (48px vs 80px)
- Single condensed section
- 36px input height
```

### Compact Mode Security Tab

```
┌─── Security ────────────────────────────────────────────────────────────────┐
│                                                                             │
│  Password: Changed 30 days ago  [🔐 Change]     2FA: 🟢  [🔄] [🗑️]         │
│                                                                             │
│  Active Sessions (3):                                                       │
│  ├─ ● Chrome/Windows • Jakarta • Active now                    [Current]  │
│  ├─ ○ Safari/iOS • Jakarta • 2h ago                            [Revoke]   │
│  └─ ○ Chrome/macOS • Singapore • 3d ago                        [Revoke]   │
│                                                             [Revoke All]   │
│                                                                             │
│  Recent Logins: [View Full History →]                                       │
│  • Jan 19 14:32 Chrome/Win Jakarta ✓ • Jan 19 09:15 Safari/iOS Jakarta ✓  │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘

Key Differences:
- Sessions in tree list format
- Inline status/actions
- Compact login history (2 per line)
```

### Compact Mode API Keys Tab

```
┌─── API Keys ────────────────────────────────────────────────────────────────┐
│                                                                             │
│  Production 🟢 │ nxos_prod_●●●●●●abc123  │ RW │ 2h ago  │ [📋][⚙️][🔄][🗑️] │
│  Development 🟡 │ nxos_dev_●●●●●●def456  │ R  │ 5d ago  │ [📋][⚙️][🔄][🗑️] │
│                                                                             │
│                                        [+ Create New Key]  [📖 API Docs]   │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘

Key Differences:
- Single line per key
- Abbreviated columns (RW = Read+Write, R = Read only)
- Compact icon buttons
```
