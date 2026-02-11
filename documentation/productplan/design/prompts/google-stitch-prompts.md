# NexusOS for Google Stitch (Labs) - Prompts

**Purpose:** Optimized prompts for [Google Stitch](https://labs.google/stitch) to generate Design Systems, Components, and UI.
**Workflow:** Follow this sequence design consistency: **Foundation -> Components -> Screens**.

---

## 🎨 1. FOUNDATION: VISUAL LANGUAGE

_Run these prompts sequentially to establish the Design System logic in Stitch._

### 1.1 Color System (Light & Dark)

**Prompt:**

```text
Create a "Color System" documentation page.
Layout: Split view with Light Mode (Left) and Dark Mode (Right).

CONTENT:
1. Primary Palette (Indigo):
   - Show swatches 50-950.
   - Light Mode Primary: #6366F1 (Indigo-600).
   - Dark Mode Primary: #818CF8 (Indigo-400).
2. Secondary Palette (Teal):
   - Light Mode Secondary: #14B8A6 (Teal-500).
   - Dark Mode Secondary: #2DD4BF (Teal-400).
3. Semantic Colors (Badges/Alerts):
   - Success (Emerald), Warning (Amber), Error (Red), Info (Blue).
   - Show how these look on White bg vs Slate-950 bg.

REQUIREMENT: Generate tailwind code showing how consistent variable names map to different hex codes in .dark class.
```

### 1.2 Typography Scale (Responsive)

**Prompt:**

```text
Create a "Typography Scale" page using Geist Font family.
Grid Layout: Show scale on White Paper (Light) and Dark Slate Paper (Dark).

SCALE DEFINITIONS:
1. Display: 36px (Bold).
2. Headings: H1 (24px SemiBold), H2 (20px), H3 (18px).
3. Body:
   - Body Large: 16px (Relaxed reading).
   - Body Default: 14px (Standard UI).
   - Caption: 12px (Muted text).
4. Code: Geist Mono 13px.

DENSITY CONTEXT:
- Show comparison: "Comfort Mode" (Line-height 1.6) vs "Compact Mode" (Line-height 1.3).
```

### 1.3 Spacing & Radius (Compact vs Comfort)

**Prompt:**

```text
Create a "Spacing & Radius" specification card.
Layout: Side-by-Side comparison of "SaaS Comfort" vs "Enterprise Compact".

LEFT SIDE (Comfort/SaaS):
- Radius: 8px (Rounded-md/lg).
- Padding: p-4 to p-6 (Spacious).
- Gap: gap-4.
- Input Height: 44px.
- Visual Vibe: Friendly, approachable.

RIGHT SIDE (Compact/Enterprise):
- Radius: 4px (Rounded-sm).
- Padding: p-2 to p-3 (Dense).
- Gap: gap-2.
- Input Height: 32px.
- Visual Vibe: Precise, data-dense.

Output: Visual examples of a "Card Component" rendered in both density modes.
```

### 1.4 Shadows & Depth (Elevation)

**Prompt:**

```text
Create an "Elevation & Depth" guide.
Context: How depth works in Light vs Dark mode.

LIGHT MODE SECTION:
- Use Shadows (shadow-sm, shadow-md, shadow-lg, shadow-xl).
- Card style: White bg with soft gray shadow.

DARK MODE SECTION:
- NO SHADOWS.
- Use Borders instead (border-slate-800).
- Use Inner Glow (ring-1 ring-white/5).
- Card style: Slate-900 bg on Slate-950 surface.
```

---

## 🧩 2. COMPONENT LIBRARY (Variants)

Generate these individually to build up your asset library.

### 2.1 Buttons & Actions

**Prompt:**

```text
Create a "Button Component Set" showing all interaction states across Comfort and Compact density modes.

Design System Context:
- Use `@design-tokens` for all color, spacing, and radius values.

Grid Layout:
- Rows: Primary (@brand-primary), Secondary (@brand-secondary), Outline (@border-base), Ghost (Transparent), Destructive (@status-error).
- Columns: Default, Hover, Active/Pressed, Disabled, Loading (Spinner).

Density Specifications:
1. COMFORT (Standard):
   - Height: 40px.
   - Radius: @radius-xl (12px).
   - Padding: @spacing-comfort (p-4).
2. COMPACT (Enterprise):
   - Height: 32px.
   - Radius: @radius-sm (4px).
   - Padding: @spacing-dense (p-2).

Shared Specifications:
- Typography: @font-weight-medium.
- Focus: @ring-indigo-300 with @ring-offset.
```

### 2.2 Form Inputs & Controls

**Prompt:**

```text
Create a "Form Elements Showcase" demonstrating input states across Comfort and Compact density modes.

Design System Context:
- Use `@design-tokens` for all colors, spacing, and radius values.

Components to include:
1. Text Input: Show Default, Focus, Error (with helper text), and Disabled states.
2. Select Dropdown: Custom chevron icon, @brand-primary focus ring.
3. Checkbox & Radio: @brand-primary accent color.
4. Toggle Switch: State On/Off.

Density Specifications:
1. COMFORT (SaaS):
   - Input Height: 44px.
   - Radius: @radius-lg.
   - Padding: @spacing-comfort.
2. COMPACT (Enterprise):
   - Input Height: 32px.
   - Radius: @radius-sm.
   - Padding: @spacing-dense.
   - Typography: @font-size-sm.

Visual Style:
- Light Mode: @bg-white, @border-slate-300.
- Dark Mode: @bg-slate-900, @border-slate-700.
- Focus: @ring-2 @ring-indigo-500.
```

### 2.3 Status Badges & Alerts

**Prompt:**

```text
Create a set of "Semantic Badges" and "Alert Banners" demonstrating Success, Warning, Error, and Info states across Comfort and Compact density modes.

Design System Context:
- Use `@design-tokens` for all semantic colors, spacing, and radius values.

Components to include:
1. Semantic Badges: Pill-shaped, subtle background (10% opacity), bold colored text.
2. Alert Banners: Full-width container, thick left border (4px), leading icon, and descriptive text.

Density Specifications:
1. COMFORT (SaaS):
   - Badge: @spacing-comfort-x @spacing-comfort-y, @radius-full.
   - Alert: @spacing-comfort padding, @font-size-base.
2. COMPACT (Enterprise):
   - Badge: @spacing-dense-x @spacing-dense-y, @radius-sm.
   - Alert: @spacing-dense padding, @font-size-sm.

Visual Style:
- Success: @color-emerald-500 (Text/Border), @bg-emerald-500 (10% opacity).
- Warning: @color-amber-500 (Text/Border), @bg-amber-500 (10% opacity).
- Error: @color-red-500 (Text/Border), @bg-red-500 (10% opacity).
- Info: @color-blue-500 (Text/Border), @bg-blue-500 (10% opacity).
```

---

## 🖥️ 3. UI GENERATION (Screens)

Now combine standard components into full pages.

### 3.1 Admin Dashboard (Hybrid Layout)

**Prompt:**

```text
Create a professional Admin Dashboard for "NexusOS" demonstrating both Comfort and Compact density modes.

Design System Context:
- Use `@design-tokens` for all colors, spacing, and radius values.

Structure:
1. Left Sidebar (Fixed, 280px, @bg-slate-900):
   - Logo "NexusOS" at top.
   - Nav items: Dashboard, Users, Roles, Settings using `@font-weight-medium`.
   - Active user profile at bottom.
2. Top Navbar (@bg-white, @border-base):
   - Global Search input (Command+K style) with `@radius-md`.
   - Right side: Notification bell, Theme toggle, User avatar.
3. Main Content Area (@bg-slate-50):
   - Page Title "Overview".
   - Top Row: 4 KPI Cards (Users, Revenue, Active Sessions, Alerts). Each card has an icon, big number, and trend indicator.
   - Middle Row: Data Grid showing "Recent Activity" (User, Action, Time, Status badge).
   - Bottom Row: Two charts (Line chart for Traffic, Bar chart for Signups).

Density Specifications:
1. COMFORT (SaaS):
   - Layout Padding: @spacing-comfort (p-6).
   - Card Radius: @radius-xl.
   - Component Height: 40px.
   - Typography: @font-size-base.
2. COMPACT (Enterprise):
   - Layout Padding: @spacing-dense (p-4).
   - Card Radius: @radius-sm.
   - Component Height: 32px.
   - Typography: @font-size-sm.

Visual Style:
- Primary Color: @brand-primary (Indigo-600).
- Status Badges: Use semantic colors (@status-success, @status-error) with 10% opacity backgrounds.
- Borders: @border-base (Slate-200).
- Focus States: @ring-2 @ring-indigo-500.
```

### 3.2 Login Screen (Authentication)

**Prompt:**

```text
Create a modern Split-Screen Login page using `@design-tokens`.

Structure:
1. Left Side (50% width, `@bg-slate-900`):
   - Abstract geometric patterns.
   - Testimonial quote in `@text-white` at the bottom.
2. Right Side (50% width, `@bg-white`):
   - Centered Login Form.
   - Header: "Welcome back" using `@font-weight-bold`.
   - Inputs: Email (with icon), Password (with eye toggle) using `@border-base`.
   - "Forgot Password" link using `@brand-primary`.
   - Primary Button: "Sign In", full width, `@brand-primary`, `@text-white`.
   - Secondary Button: "Sign in with Google", outline style using `@border-base`.
   - Footer: "Don't have an account? Sign up" with link in `@brand-primary`.

Density Specifications:
1. COMFORT:
   - Form Padding: `@spacing-comfort`.
   - Component Height: 44px.
   - Typography: `@font-size-base`.
   - Radius: `@radius-xl`.
2. COMPACT:
   - Form Padding: `@spacing-dense`.
   - Component Height: 36px.
   - Typography: `@font-size-sm`.
   - Radius: `@radius-sm`.
```

### 3.3 User Management Grid (Data Heavy)

**Prompt:**

```text
Create a detailed "User Management" data table screen for "NexusOS" demonstrating both Comfort and Compact density modes across Light and Dark themes.

Design System Context:
- Use `@design-tokens` for all colors, spacing, and radius values.

Structure:
1. Header:
   - Title "Users" (@font-weight-bold).
   - Action Bar: Search input (left), "Filter" & "Export" (outline buttons), "Add User" (Primary Indigo button).
2. Data Table:
   - Columns: Checkbox, User (Avatar + Name + Email), Role (Badge), Status (Badge), Last Login, Actions (Icon).
   - Rows: 5 rows of realistic data.
   - Interactive: Row hover state, checkbox selection.
3. Pagination Footer:
   - "Showing 1-10 of 500 results".
   - Navigation: Previous/Next buttons and page numbers.

Density Specifications:
1. COMFORT (SaaS):
   - Cell Padding: @spacing-comfort-y @spacing-comfort-x.
   - Row Height: 64px.
   - Typography: @font-size-base.
   - Radius: @radius-lg.
2. COMPACT (Enterprise):
   - Cell Padding: @spacing-dense-y @spacing-dense-x.
   - Row Height: 40px.
   - Typography: @font-size-sm.
   - Radius: @radius-sm.

Visual Style:
- Light Mode: @bg-white, @border-slate-200, @text-slate-900.
- Dark Mode: @bg-slate-900, @border-slate-700, @text-slate-100.
- Status Badges (10% opacity bg):
  - Active: @color-emerald-500.
  - Inactive: @color-slate-400.
  - Admin Role: @color-purple-500.
  - Member Role: @color-blue-500.
```
