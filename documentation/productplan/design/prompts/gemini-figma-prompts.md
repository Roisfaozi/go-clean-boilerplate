# 🚀 Ready-to-Use Prompts (Gemini 3 & Figma)

**Instruksi:**

1. **Copy** blok "MASTER CONTEXT" dulu dan paste ke chat Gemini.
2. **Copy** blok "TASK PROMPT" sesuai kebutuhan kamu (UI, Wireframe, atau Code).
3. Untuk **Figma**, gunakan prompt "Visual Generation" di plugin AI atau gunakan hasil dari Gemini sebagai referensi.

---

## 1️⃣ MASTER CONTEXT (PASTE FIRST)

```text
SYSTEM ROLE: Senior Product Designer & Design Systems Architect
PROJECT: NexusOS (Enterprise SaaS Platform)

DESIGN SYSTEM SPECIFICATIONS ("Nebula"):
- Philosophy: "Fluid Density" (Adapts between Comfort/SaaS and Compact/Enterprise modes).
- PRIMARY Color: Indigo (#6366F1 Light / #818CF8 Dark).
- SECONDARY Color: Teal (#14B8A6 Light / #2DD4BF Dark).
- ACCENT: Violet (#8B5CF6).
- NEUTRALS: Slate (#0F172A Surface / #F8FAFC Bg).
- SEMANTIC: Info (Blue #3B82F6), Success (Emerald #10B981), Warning (Amber #F59E0B), Danger (Red #DC2626).
- TYPOGRAPHY: Geist Sans (UI), Geist Mono (Code/Data).
- RADIUS: 8px (Comfort) / 4px (Compact).
- SPACING: 4pt scale (16px component gap).
- SHADOWS: Soft colored shadows (Light) / Inner glow borders (Dark).

CORE COMPONENTS:
- HyperGrid (Data Table): High density, resizable, sticky headers.
- Sidebar: Collapsible (280px -> 72px rail).
- Cards: Minimal border, clean padding.

OUTPUT RULES:
- When asked for UI: Describe structure, spacing, and colors explicitly.
- When asked for Code: Use React, Tailwind CSS v4, Lucide Icons.
- Always verify: Accessibility (WCAG AA) and Contrast.
```

---

## 2️⃣ UI GENERATION PROMPTS (For Visuals)

### 🎨 Dashboard UI

```text
TASK: Generate a High-Fidelity Dashboard UI Design
SCREEN: Admin Overview Dashboard
MODE: Comfort Mode (Light Theme)
REQUIREMENTS:
1. Header: "Dashboard", Breadcrumbs, Global Search, Profile Avatar.
2. KPI Cards (Row 1): 4 Cards (Total Users, Revenue, Active Sessions, Pending Alerts). Use Gradient icons (Indigo/Teal/Violet).
3. Main Chart (Row 2): Large Line Chart "Revenue Trend" (Indigo primary line, Teal secondary line).
4. Recent Activity (Row 3 Left): List of 5 recent logins with status badges (Emerald/Amber).
5. Quick Actions (Row 3 Right): Grid of 4 buttons (Add User, Generate Report, Settings, View Logs).
STYLE: Modern SaaS, clean white background #FFFFFF, Slate-50 surface #F8FAFC. Soft shadows.
OUTPUT: Detailed visual description of layout and specific Tailwind color classes for each element.
```

### 📱 Mobile Responsive View

```text
TASK: Generate Mobile Responsive View
SCREEN: User List Data Table
CONSTRAINT: Mobile Viewport (375px width)
ADAPTATION:
- "HyperGrid" table transforms into "Card List Layout".
- Sidebar becomes "Burger Menu" (Slide-over).
- Bulk Actions move to "Bottom Action Sheet".
DETAILS:
- Each user row becomes a Card with Avatar left, Name/Email stacked, Status badge top-right.
- Action menu (3 dots) on bottom-right of card.
STYLE: Compact, touch-friendly targets (min 44px).
OUTPUT: Visual hierarchy description and component stacking order.
```

### 🌑 Dark Mode Variation

```text
TASK: Convert UI to Dark Mode (Eclipse Theme)
TARGET: Settings Page
CHANGES:
- Background: Change White to Slate-950 (#020617).
- Surface: Change Slate-50 to Slate-900 (#0F172A).
- Borders: Slate-800 (#1E293B).
- Text: Slate-900 -> Slate-50 (#F8FAFC).
- Primary Action: Indigo-600 -> Indigo-500 (#6366F1).
- Secondary: Teal-500 -> Teal-400 (#2DD4BF).
- Shadows: REMOVE all shadows. ADD 1px border + faint inner white glow (5%).
OUTPUT: List of specific color token swaps for the page elements.
```

---

## 3️⃣ CODING PROMPTS (For Dev)

### 💻 React Component (Tailwind v4)

```text
TASK: Create React Component
COMPONENT: StatusBadge
PROPS:
- variant: 'success' | 'warning' | 'error' | 'info' | 'neutral'
- style: 'solid' | 'subtle' | 'outline'
- density: 'comfort' | 'compact'
SPECS:
- Text: Geist Sans, Medium weight.
- Radius: Full pill (9999px) for Comfort, Rounded (4px) for Compact.
- Colors:
  - Success: Bg emerald-100 text emerald-700 (Subtle)
  - Warning: Bg amber-100 text amber-700 (Subtle)
  - Error: Bg red-100 text red-700 (Subtle)
  - Info: Bg blue-100 text blue-700 (Subtle)
CODE: Return clean React Functional Component with Typescript interfaces and Tailwind classes.
```

### 📄 Full Page Layout

```text
TASK: Implement Profile Settings Page
STACK: Next.js 16, Tailwind, Lucide React
LAYOUT:
1. PageHeader: Title "Profile Settings", Subtitle "Manage your account info".
2. Tabs: [General, Security, Notifications, API Keys].
3. Form (General Tab):
   - Avatar Upload (Circle, 80px).
   - Grid (2 Col): First Name, Last Name.
   - Full Width: Email (Disabled).
   - Textarea: Bio.
   - Buttons: "Save Changes" (Primary Indigo), "Cancel" (Ghost).
DENSITY: Comfort Mode (Spacious).
CODE: Single file react component `ProfilePage.tsx` with responsive grid classes.
```

---

## 4️⃣ FIGMA / IMAGE GEN PROMPTS

### 🖌️ Design System Asset Generation

_Use this in Image Generation tools (Gemini/Imagen/Midjourney)_

```text
Design System Component Set, "NexusOS", Clean Professional UI Style.
CANVAS: Split view (Light Mode vs Dark Mode).
CONTENT: Organized grid of UI buttons.
ROWS (Show for both Light/Dark):
1. Primary Buttons: Solid Indigo Blue (#6366F1/Light, #818CF8/Dark).
2. Secondary Buttons: Solid Teal (#14B8A6/Light, #2DD4BF/Dark).
3. Outline Buttons: Slate border (Light), Slate-700 border (Dark).
4. Destructive Buttons: Red (#DC2626).
COLUMNS: Default, Hover, Active, Disabled states.
STYLE: High fidelity, vector style, consistent spacing, Figma documentation aesthetic.
```
