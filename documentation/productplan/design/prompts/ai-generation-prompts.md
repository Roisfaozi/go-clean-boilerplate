# NexusOS Design Generation Prompts for Gemini 3

**Purpose:** AI prompts for generating mockups, wireframes, and design system visuals  
**Target AI:** Gemini 3, Gemini 2.0 Flash, or similar image generation models  
**Design System:** Nexus Design System v1.0 ("Fluid Density")

---

## 📋 Quick Reference: Design Context

Before using prompts, ensure the AI understands this context:

```
DESIGN SYSTEM CONTEXT:
- Name: NexusOS Admin Dashboard
- Philosophy: "Fluid Density" - dual-mode (Comfort for SaaS, Compact for Enterprise)
- Colors: Nebula Palette - Primary: Indigo (#6366F1), Background: White/Slate-950
- Typography: Geist Sans (UI), Geist Mono (data/code)
- Tech Stack: Next.js 16, Tailwind v4, Shadcn UI
- Style: Modern SaaS aesthetic with Enterprise data-density option
```

---

## 🎨 SECTION 1: DESIGN SYSTEM COMPONENT PROMPTS

### 1.1 Button Component Set

```
Generate a professional UI design showing a button component library for a modern SaaS admin dashboard.

SPECIFICATIONS:
- Layout: Split view (Light Mode vs Dark Mode)
- Style: Modern, clean, Shadcn UI inspired

BUTTON VARIANTS (Show in both Light/Dark contexts):
1. Primary: Indigo (#6366F1) / Dark Mode Lighter Indigo (#818CF8)
2. Secondary: Teal (#14B8A6) / Dark Mode Teal (#2DD4BF)
3. Ghost: Transparent with dark text / Dark Mode light text
4. Destructive: Red background works on both

BUTTON VARIANTS (show all in a grid):
1. Primary: Solid indigo (#6366F1) background, white text, 8px rounded corners
2. Secondary: White background with slate-200 border, dark text
3. Ghost: Transparent with dark text, subtle hover state
4. Destructive: Red (#DC2626) background, white text
5. Magic/AI: Gradient border (indigo to violet), sparkle icon

SIZES (show each variant in 3 sizes):
- Large: 44px height, 20px horizontal padding
- Medium: 36px height, 16px horizontal padding
- Small: 32px height, 12px horizontal padding

STATES (show for primary button):
- Default
- Hover (slightly brighter)
- Pressed (slightly darker)
- Disabled (50% opacity)
- Loading (with spinner icon)

Include button with icon-left, icon-right, and icon-only variants.
Typography: Sans-serif, 14px, medium weight.
Show clear contrast between Light Mode (White bg) and Dark Mode (Slate-950 bg) sections.
No device frames, clean component showcase on canvas.
```

### 1.2 Input Field Component Set

```
Generate a professional UI design showing input field components (Light & Dark Mode).

SPECIFICATIONS:
- Layout: Side-by-side comparison
- Left: Light Mode (Gray #F8FAFC bg)
- Right: Dark Mode (Slate #020617 bg)

INPUT TYPES (show in a vertical list):
1. Text Input - standard single line
2. Password Input - with show/hide eye icon
3. Email Input - with mail icon left
4. Number Input - with increment/decrement arrows
5. Search Input - with magnifying glass icon and keyboard shortcut badge (⌘K)
6. Textarea - multi-line, auto-grow appearance

STATES (show for text input):
- Default: 1px slate-200 border, white background
- Focus: 2px indigo (#6366F1) border with soft blue glow ring
- Error: 1px red border, red error icon, red helper text below
- Disabled: Gray background, 50% opacity
- With AI Button: Small sparkle icon button inside right side

ANATOMY:
- Label above input (13px, medium weight, dark gray)
- Input field (44px height, 16px horizontal padding, 8px rounded)
- Helper text below (12px, slate-500)
- Error message below (12px, red)


VISUAL STYLE:
- Light Mode: White input bg, Slate-200 border, Indigo focus ring
- Dark Mode: Slate-900 input bg, Slate-800 border, Indigo-500 focus border (no glow)

Show label stacking differences:
- Vertical (SaaS)
- Horizontal (Enterprise)

Typography: Sans-serif. No device frames.
```

### 1.3 Badge/Status Chip Component Set

```
Generate a professional UI design showing badge/status chip components for a modern admin dashboard.

SPECIFICATIONS:
- Style: Modern, minimal badges
- Background: White design canvas

SEMANTIC VARIANTS (show all):
1. Success: Emerald green (#10B981) - checkmark icon, "Active" text
2. Warning: Amber (#F59E0B) - alert icon, "Pending" text
3. Error: Red (#EF4444) - x-circle icon, "Failed" text
4. Info: Blue (#3B82F6) - info icon, "Processing" text
5. Neutral: Slate gray - no icon, "Draft" text
6. Primary: Indigo (#6366F1) - "New" text

STYLE VARIANTS (show for Success):
1. Subtle: 10% opacity background, colored text (preferred)
2. Solid: Full color background, white text
3. Outline: Transparent with colored border

SHAPES (show for all variants):
1. Pill: Fully rounded ends (border-radius: 9999px) - for SaaS mode
2. Rounded Rectangle: 4px border-radius - for Enterprise mode

SIZE:
- Height: 24px
- Padding: 8px horizontal
- Font: 12px, medium weight
- Icon: 14px

Show badges both standalone and inline with text.
No device frames.
```

### 1.4 Data Table Row States

```
Generate a professional UI design showing data table row states for an enterprise admin dashboard.

SPECIFICATIONS:
- Style: Clean, Excel-like data grid
- Background: White

TABLE STRUCTURE:
- 6 columns: Checkbox | Avatar+Name | Email | Role (badge) | Status (badge) | Actions (...)
- Sample data rows with realistic names and emails

ROW STATES (show each):
1. Default: White background, 1px bottom border slate-200
2. Hover: Very light blue tint (#EEF2FF), cursor pointer
3. Selected: Light indigo tint with 2px left indigo border
4. Zebra Striping (dark mode): Alternating transparent and slate-900/50%

HEADER ROW:
- Background: Slate-50 (#F8FAFC)
- Text: Uppercase, 12px, bold, slate-500 color
- Sort indicators: Up/down chevron icons

COMPACT VS COMFORT MODE:
Show same table in two versions:
1. Comfort: 64px row height, 14px font, horizontal borders only
2. Compact: 36px row height, 13px font, full grid lines visible

Actions column: Show "..." menu icon that appears on hover.
No device frames.
```

### 1.5 Card Component Variants

```
Generate a professional UI design showing card components for a SaaS admin dashboard.

SPECIFICATIONS:
- Background: Light gray canvas

CARD TYPES:

1. KPI METRIC CARD (SaaS Style):
- Size: 280px wide, 140px tall
- White background, soft shadow (shadow-md)
- No border
- Content: Large icon (48px) in pastel circle, big number (36px, bold), label (14px, gray), trend arrow with percentage
- Border radius: 16px

2. KPI METRIC CARD (Enterprise Style):
- Size: 220px wide, 100px tall
- Slate-50 background, 1px slate-200 border, NO shadow
- Content: Small/no icon, medium number (24px), inline sparkline chart (mini line graph)
- Border radius: 4px

3. USER PROFILE CARD:
- Avatar (64px circle), Name (18px bold), Email (14px gray), Role badge
- Two action buttons: Edit, View

4. ROLE CARD:
- Icon + Role name, description, member count badge
- Permission summary as small badges

DARK MODE CARD:
- Background: Slate-900 (#0F172A)
- Border: 1px slate-800
- No shadow
- Subtle inner white glow ring (ring-white/5%)

Show cards in a grid layout. No device frames.
```

### 1.6 Navigation Sidebar

```
Generate a professional UI design showing sidebar navigation (Light & Dark Variants).

SPECIFICATIONS:
- Two versions side by side:

EXPANDED SIDEBAR (280px width):
- Top: Logo + "NexusOS" text
- Search bar with ⌘K shortcut badge
- Navigation groups with section labels ("MAIN", "MANAGEMENT", "SETTINGS")
- Nav items: Icon (20px) + Label + Optional badge (for counts)
- Active state: Indigo-50 background, 3px left indigo border, indigo icon
- Hover state: Slate-100 background
- Sub-menu accordion: Expanded with child items indented
- Bottom: Collapse toggle button, AI Chat button with sparkle icon

COLLAPSED SIDEBAR (72px width - Rail):
- Logo icon only
- Navigation icons only (centered)
- Tooltip appearing on hover showing label
- Active state: Indigo background on icon
- Bottom: Collapse toggle, AI icon

VARIANT A: LIGHT SIDEBAR
- Background: White
- Active Item: Indigo-50 bg + Indigo text
- Text: Slate-700

VARIANT B: DARK SIDEBAR
- Background: Slate-900 (#0F172A)
- Active Item: Indigo-500/10 bg + Indigo-400 text
- Text: Slate-400

NAVIGATION ITEMS:
1. Dashboard (Home icon)
2. Users (Users icon) - with "12" badge
3. Roles (Shield icon)
4. Access Control (Lock icon)
5. Audit Logs (FileText icon)
6. Settings (Cog icon)

Colors: White background, slate-800 text, indigo active states.
Typography: 14px for labels, 12px for section headers.
Show both light mode and dark mode. No device frames.

Show both Expanded (280px) and Collapsed (72px) versions for context.
Clean, vector style.
```

---

## 🖼️ SECTION 2: SCREEN MOCKUP PROMPTS

### 2.1 Dashboard Screen (Comfort Mode)

```
Generate a professional UI mockup of an admin dashboard for a SaaS application.

LAYOUT:
- Left: Expanded sidebar (280px) with navigation
- Top: Navbar (80px height) with search bar, density toggle, theme toggle, user avatar
- Main content area with padding (32px)

CONTENT ZONES:
Zone A (Top): 4 KPI metric cards in a row
- Total Users: 1,234 with +12% trend arrow (green)
- Active Roles: 8
- Today's Actions: 156 with live pulse indicator
- Failed Logins: 3 with warning indicator (amber)

Zone B (Middle): Recent Audit Logs table
- 5 rows preview with columns: Time, User (avatar+name), Action (badge), Resource, Status
- "View All" link at bottom right

Zone C (Bottom): Quick action buttons
- "+ Add User" (primary button)
- "+ Create Role" (secondary button)
- "Export Logs" (secondary with download icon)
- Settings (ghost button with cog icon)

STYLE:
- Clean, modern SaaS aesthetic
- White background, soft shadows on cards
- Indigo (#6366F1) primary color
- Geist-like sans-serif typography
- Rounded corners (12px on cards)

Desktop screen (1440x900). Show as clean UI, no browser chrome or device frame.
```

### 2.2 Dashboard Screen (Compact Mode - Enterprise)

```
Generate a professional UI mockup of an admin dashboard optimized for enterprise data density.

LAYOUT:
- Left: Collapsed sidebar rail (72px) with icons only
- Top: Slim navbar (56px height)
- Main content with tight padding (16px)

CONTENT:
Zone A: 4 compact KPI cards in a row
- Smaller text, no shadows, border only, sparkline charts inline

Zone B: Dense data table showing 10+ rows
- Compact row height (36px)
- Full grid lines visible
- Uppercase headers
- 13px font size

Zone C: Horizontal toolbar with action buttons (32px height)

STYLE:
- Enterprise data-dense aesthetic
- Sharp corners (4px radius)
- High contrast borders
- No shadows
- Tight spacing between elements
- More rows visible on screen

Desktop screen (1920x1080). Clean UI, no device frame.
```

### 2.3 User Management Screen

```
Generate a professional UI mockup of a user management list page for an admin dashboard.

LAYOUT:
- Sidebar with "Users" item highlighted
- Page header: "User Management" title, "+ Add User" primary button

TOOLBAR:
- Search input with placeholder "Search users..."
- Filter dropdown button
- Columns visibility dropdown
- Density toggle (comfort/compact switch)

DATA TABLE (Hyper-Grid):
- Columns: Checkbox | Avatar+Name | Email | Role (badge) | Status (badge) | Created Date | Actions
- 8 rows of sample user data
- First column (checkbox) sticky on horizontal scroll
- Header row sticky at top
- One row showing hover state with visible "..." actions menu

PAGINATION FOOTER:
- "Showing 1-20 of 1,234 users"
- Rows per page dropdown: 10/20/50/100
- Page navigation: < 1 2 3 ... 62 >

Sample users with diverse names, roles (Admin, Editor, Viewer), and statuses (Active, Inactive, Pending).
Desktop screen, light mode, modern SaaS style. No device frame.
```

### 2.4 Permission Matrix Screen

```
Generate a professional UI mockup of a permission matrix grid for an RBAC admin dashboard.

LAYOUT:
- Sidebar with "Access Control" highlighted
- Page header: "Permission Matrix" title, "+ Add Role" and "+ Add Resource" buttons

MATRIX GRID:
- Columns: Role name | /users | /roles | /audit | /content | /api/*
- Rows: superadmin, admin, editor, viewer
- Cell content: 4 small squares representing C-R-U-D (Create, Read, Update, Delete)
  - Filled square = permission enabled (indigo color)
  - Empty/gray square = permission disabled

INTERACTION HINTS:
- One row showing hover highlight
- One cell showing clicked state with popup menu for toggling individual permissions

LEGEND:
- Bottom of matrix: "C = Create, R = Read, U = Update, D = Delete"
- Filled vs empty visual explanation

SLIDE-OVER PANEL (shown partially):
- Right side panel showing "Role: admin" details
- Member list with avatars
- Permission toggles organized by resource

Visual style: Clean grid layout, clear visual hierarchy, indigo primary color.
Desktop screen, modern enterprise UI style. No device frame.
```

### 2.5 Audit Logs Screen

```
Generate a professional UI mockup of an audit logs page for security monitoring.

LAYOUT:
- Sidebar with "Audit Logs" highlighted
- Page header: "Audit Logs" title with export dropdown button

FILTER BAR:
- Search input: "Search logs..."
- Date range picker with calendar icon
- User filter dropdown
- Action type filter dropdown
- Clear filters link

DATA TABLE:
- Columns: Timestamp | User (avatar+name) | Action | Resource | IP Address | Status
- 10 rows of sample audit data
- Action badges color-coded:
  - CREATE: Emerald green
  - READ: Slate gray
  - UPDATE: Amber
  - DELETE: Red
  - LOGIN: Indigo

STATUS BADGES:
- Success: Green checkmark + "OK"
- Failed: Red X + "FAILED"

TIMESTAMP FORMAT: Relative time (2m ago, 5m ago, 1h ago)

PAGINATION: Standard pagination footer

DENSITY TOGGLE: Shown in toolbar, currently set to "Compact" mode
(table showing compact styling with smaller rows and tighter spacing)

Desktop screen, dark mode (#020617 background), enterprise style. No device frame.
```

### 2.6 Login Page (Split Layout)

```
Generate a professional UI mockup of a login page for a SaaS admin dashboard.

LAYOUT: 50/50 split screen

LEFT PANEL (Functional):
- White/light background
- Top-left: NexusOS logo
- Center-aligned content:
  - Heading: "Welcome Back" (24px, bold)
  - Subheading: "Sign in to your account" (14px, gray)
  - Email input field with label
  - Password input field with label and show/hide toggle
  - "Remember me" checkbox
  - "Sign In" primary button (full width, 44px height)
  - "Forgot password?" link
  - Divider with "or continue with"
  - Social login buttons: Google, GitHub

RIGHT PANEL (Visual/Branding):
- Gradient background: Deep indigo to violet (#4F46E5 to #8B5CF6)
- Subtle noise texture overlay
- 3D abstract floating shapes or dashboard screenshot mockup (skewed/angled)
- Testimonial quote: "NexusOS transformed how we manage permissions" - Company logo
- "Trusted by 500+ companies" with small company logos

STYLE:
- Modern, premium SaaS aesthetic
- Soft shadows on form area
- Rounded inputs (8px)
- Primary button indigo

Screen size: 1440x900. Clean UI, no browser frame.
```

### 2.7 AI Chat Interface

```
Generate a professional UI mockup of an AI chat assistant interface for an admin dashboard.

LAYOUT: Full-screen chat view (replacing main content area)

STRUCTURE:
Header:
- "AI Assistant" title with sparkle icon
- Status indicator: green dot + "Online"
- Close (X) and Minimize buttons

Chat Body:
- Clean message thread layout
- User messages: Right-aligned, slate-100 background, rounded bubble
- AI messages: Left-aligned, subtle indigo-50 background with thin indigo border, rounded bubble
- AI avatar: Small sparkle/brain icon
- Timestamps below messages (12px, gray)

SAMPLE CONVERSATION:
User: "Show me users with failed login attempts this week"
AI: [Renders a small data table with 3 users, their emails, and attempt counts]
User: "Disable the top account"
AI: "✓ User john@example.com has been disabled. Would you like me to notify them via email?"

AI PROCESSING STATE:
- Show one AI message with shimmer/pulse animation
- Typing indicator dots or gradient shimmer

INPUT AREA:
- Multi-line textarea with auto-grow
- Left: Attachment/context button (paperclip icon)
- Right: Send button (indigo, arrow icon)
- Placeholder: "Ask AI anything about your system..."

DOCK MODE INDICATOR:
- Small toggle to switch between "Float" and "Split View" modes

Light mode, modern chat UI aesthetic. Desktop screen, no device frame.
```

---

## 📱 SECTION 3: MOBILE RESPONSIVE PROMPTS

### 3.1 Mobile Dashboard

```
Generate a professional mobile UI mockup of an admin dashboard app.

DEVICE: iPhone 15 Pro size (393x852)

LAYOUT:
- Top: Slim navbar with hamburger menu icon, "Dashboard" title, user avatar
- Main: Scrollable content area
- Bottom: Navigation bar with 5 icons (Home, Users, Access, Audit, More)

CONTENT:
- 2x2 grid of KPI cards (compact size)
- Recent activity section with 3 card-style log entries
- FAB (Floating Action Button) in bottom-right corner with + icon

STYLE:
- Bottom nav: 64px height, icons 24px, active state filled with label
- Cards: Touch-friendly (full-width, stacked)
- FAB: 56px diameter, indigo primary

Mobile optimized spacing and touch targets. Light mode. Show as device mockup with iPhone frame.
```

### 3.2 Mobile Data Table (Card View)

```
Generate a professional mobile UI mockup showing a user list in card format.

DEVICE: iPhone 15 Pro size

LAYOUT:
- Top navbar with back arrow, "Users" title, hamburger menu
- Search bar below navbar
- Scrollable list of user cards
- Bottom navigation bar

USER CARD STRUCTURE:
- Avatar (48px) on left
- Name (16px bold) and email (14px gray)
- Role badge and Status badge
- Created date (12px gray)
- Horizontal divider
- Action buttons: Edit, Assign Role, "..." menu

INTERACTIONS:
- Pull-to-refresh indicator at top
- Infinite scroll (no pagination)

Stack 4 user cards visible on screen. Light mode, modern mobile UI. Show with iPhone device frame.
```

---

## 🌙 SECTION 4: DARK MODE PROMPTS

### 4.1 Dashboard Dark Mode

```
Generate a professional UI mockup of an admin dashboard in dark mode.

COLORS (Eclipse palette):
- Page background: #020617 (Slate-950, deep blue-black)
- Card/Surface background: #0F172A (Slate-900)
- Borders: #1E293B (Slate-800)
- Primary text: #F8FAFC (Slate-50, off-white)
- Muted text: #94A3B8 (Slate-400)
- Primary accent: #818CF8 (Indigo-400, brighter for dark bg)

LAYOUT:
- Same dashboard layout as light mode
- Sidebar with dark background
- Navbar with dark background

KEY DIFFERENCES FROM LIGHT:
- Cards have 1px border instead of shadow
- Subtle inner glow on cards (ring-white/5%)
- Tables have zebra striping (alternating row backgrounds)
- Primary buttons use lighter indigo (#818CF8)
- Active navigation has indigo glow

Premium tech aesthetic, not just inverted colors. Desktop 1440x900. No device frame.
```

### 4.2 Data Table Dark Mode

```
Generate a professional UI mockup of a data table in dark mode for an enterprise dashboard.

COLORS:
- Background: #020617
- Table surface: #0F172A
- Borders: #1E293B
- Header: #0F172A with uppercase text
- Text: #F8FAFC

TABLE FEATURES:
- Zebra striping: Odd rows have subtle Slate-900/50% background
- Hover row: Indigo-500 at 10% opacity tint
- Selected row: Indigo-500 at 20% opacity + left border indicator
- Column sort icons in header

Show 8 rows of user data with various statuses.
Compact mode styling (36px row height, tight padding).
Dark mode optimized for prolonged data entry use.
Desktop screen, no device frame.
```

---

## 🔧 SECTION 5: PROMPT MODIFIERS

Add these to any prompt to customize output:

### Style Modifiers:

```
- "Minimalist design with maximum whitespace"
- "Enterprise-focused, data-dense layout"
- "Premium SaaS with glassmorphism effects"
- "Brutalist modern with bold typography"
```

### Technical Modifiers:

```
- "Include pixel-perfect spacing annotations"
- "Show component in all states: default, hover, focus, disabled"
- "Display responsive variants: desktop, tablet, mobile"
- "Include design token color codes as labels"
```

### Output Modifiers:

```
- "Clean UI only, no device frames or browser chrome"
- "Show as iPhone/Android device mockup"
- "Display as Figma-style component documentation"
- "Generate as side-by-side comparison: before/after or light/dark"
```

---

## 📝 USAGE TIPS

### For Best Results:

1. **Be Specific with Colors** - Always include hex codes
2. **Specify Dimensions** - Include sizes in pixels
3. **Reference Real Patterns** - Mention "Shadcn UI style" or "Linear app aesthetic"
4. **Describe Layout Zones** - Use terms like "header", "sidebar", "main content area"
5. **Include Sample Data** - Realistic names, emails, numbers
6. **State the Screen Size** - "1440x900 desktop" or "iPhone 15 size"
7. **Specify No Device Frames** - Unless you want them

### Prompt Structure:

```
1. What you're generating (component/screen/mockup)
2. Layout description
3. Content details
4. Visual style specifications
5. Colors and typography
6. Screen size and format
```

---

## 🎯 SECTION 6: WIREFRAME-SPECIFIC PROMPTS

### 6.1 Add/Edit User Modal

```
Generate a UI mockup of a user creation modal for an admin dashboard.

MODAL STRUCTURE:
- Size: 520px wide, centered on dimmed backdrop
- Header: "Add New User" title with X close button
- Form sections with dividers

FORM CONTENT:
Section 1 - Personal Information:
- Full Name (required): Text input
- Email Address (required): Email input
- Password (required): Password input with show/hide + "Generate" button

Section 2 - Access & Permissions:
- Assign Role (required): Dropdown selector
- Status: Radio buttons (Active, Pending, Inactive)
- Checkbox: "Send welcome email with login credentials"

FOOTER:
- Cancel button (secondary)
- "Create User" button (primary, indigo)

Style: Shadcn modal, 24px padding, 8px input radius.
Light mode, desktop. No device frame.
```

### 6.2 User Profile Slide-Over

```
Generate a UI mockup of a user profile slide-over panel.

POSITION: Right side panel, 420px wide, full height

STRUCTURE:
Header: User avatar (80px), name, email, status badge, close X button

Sections (collapsible accordions):
1. Account Information:
   - Created date
   - Last login (relative time)
   - IP Address
   - User Agent/Browser

2. Assigned Roles:
   - Role tags with X remove button
   - "+ Assign Role" button

3. Recent Activity:
   - 3-4 activity entries with icons
   - "View All Activity" link

Footer: "Delete User" danger button (left-aligned)

Slide-over with subtle shadow, white background.
Light mode. No device frame.
```

### 6.3 Role Cards Grid View

```
Generate a UI mockup of role management cards for RBAC system.

LAYOUT: 3-column grid of role cards

ROLE CARD DESIGN (each ~280px wide):
- Shield icon + Role name (e.g., "superadmin", "admin", "editor")
- Short description (14px, gray)
- Stats: "5 members" badge, "5 resources" badge
- "Edit →" link button

Include one "Create Role" card:
- Dashed border
- Large + icon
- "Add a new role" text

Show 4 role cards + 1 create card.
Modern card grid, 24px gap, rounded corners.
Light mode. No device frame.
```

### 6.4 Role Inheritance Tree

```
Generate a UI mockup of a role inheritance visualization tree.

STRUCTURE: Collapsible tree view panel

TREE CONTENT:
📁 superadmin (root)
    └── 📁 admin
        ├── 📁 editor
        │   └── 📁 author
        └── 📁 viewer
📁 api_client (standalone root)

Each node shows:
- Folder icon + role name
- Permission summary: 🔐 CRUD indicators
- Expand/collapse chevron
- (inherited) or (own permission) labels

LEGEND BOX:
C = Create, R = Read, U = Update, D = Delete
Permission source indicators explained

HOVER STATE: "Edit" and "+ Child Role" buttons appear

Tree lines connecting parent/child roles.
Enterprise UI style, compact. No device frame.
```

### 6.5 Permission Cell Detail Popup

```
Generate a UI mockup of a permission editing popup for RBAC matrix.

TRIGGER: Clicked cell in permission matrix grid

POPUP CONTENT (360px wide):
Header: "/users Permissions for: admin"

Permission toggles (checkbox + description each):
☑ Read - "View user list and profile details"
☑ Create - "Add new users to the system"
☑ Update - "Modify user profiles and settings"
☐ Delete - "Remove users from the system permanently"

Footer buttons:
- Cancel (secondary)
- Apply (primary)

Small dropdown popup with shadow, arrow pointing to cell.
Clean toggle UI. Light mode.
```

### 6.6 Audit Log Detail Modal

```
Generate a UI mockup of an audit log entry detail view.

MODAL SIZE: 600px wide

SECTIONS:
1. Overview Box:
   - Action: CREATE badge (green)
   - Status: ✓ Success badge
   - Timestamp: "Jan 19, 2026 14:32:15 UTC"
   - Duration: "45ms"

2. Actor Section:
   - Avatar + John Doe + john@example.com
   - Role: Admin
   - IP: 192.168.1.1
   - User Agent: Chrome/120.0.0.0
   - Location: Jakarta, Indonesia

3. Request Section:
   - Method: POST
   - Endpoint: /api/v1/users
   - Body: JSON code block with syntax highlighting

4. Response Section:
   - Status Code: 201 Created
   - Body: JSON code block

Footer: "Copy JSON" button

Technical detail modal, code blocks with dark background.
Light mode overall. No device frame.
```

### 6.7 Live Streaming Audit View

```
Generate a UI mockup of real-time audit log streaming interface.

TOOLBAR INDICATOR:
[📡 Live ◉] with green pulsing dot + "12/s" rate counter

TABLE FEATURES:
- New rows slide in from top with indigo highlight
- Highlight fades over 2 seconds
- Timestamps show "just now", "2 sec ago", etc.
- Maximum 100 rows visible

SECURITY ALERT ROW:
- Red left border (4px)
- ⚠️ Warning icon
- "FAILED" red badge
- Different background color

CONNECTION STATES (show as small badges):
- 🟢 Connected
- 🟡 Reconnecting...
- 🔴 Offline [Retry]
- ⏸️ Paused [Resume]

Real-time streaming UI, subtle animations indicated.
Dark mode preferred. No device frame.
```

### 6.8 Export Options Modal

```
Generate a UI mockup of an audit log export dialog.

MODAL CONTENT:
Title: "Export Audit Logs"

Format Selection (radio):
○ CSV (Spreadsheet compatible)
● JSON (API/Development)
○ PDF Report (Formatted document)

Date Range Display:
"Jan 15, 2026 — Jan 19, 2026"
"Records to export: 1,234"

Column Selection (checkboxes):
☑ Timestamp
☑ User
☑ Action
☑ Resource
☑ IP Address
☑ User Agent
☑ Status
☑ Request Body
☑ Response

Warning: "⚠️ Large exports may take several minutes"

Buttons: Cancel | "📤 Start Export"

Clean form modal, organized options.
Light mode. No device frame.
```

### 6.9 AI Chat Floating Widget

```
Generate a UI mockup of a floating AI chat widget.

POSITION: Fixed bottom-right, 380px wide × 500px tall

STRUCTURE:
Header:
- 🤖 NexusAI title
- Status: "● Thinking..." with pulse animation
- Dock, Minimize, Close buttons

Chat Area:
- AI greeting bubble (left-aligned, indigo-50 background)
- User message bubble (right-aligned, slate-100 background)
- AI response with markdown table rendered
- AI thinking state: shimmer/gradient animation box

Input Area:
- Multi-line textarea
- 📎 Attach button (paperclip)
- Send button (→ arrow, indigo)
- Placeholder: "Ask NexusAI anything..."

Quick Action Chips:
[📊 Analyze this page] [❓ Help] [⌨️ Shortcuts]

Floating widget with prominent shadow.
Light mode. Show as standalone widget, no page behind.
```

### 6.10 AI Chat Split/Docked View

```
Generate a UI mockup of dashboard with docked AI assistant panel.

LAYOUT: Main content 70% | AI Panel 30% (right side)

MAIN CONTENT AREA:
- Show any dashboard/table content (partial, blurred is OK)

AI PANEL (350px wide):
Header: 🤖 NexusAI [Undock] [X]
Body:
- AI analyzing current page context
- Results: "I found 3 suspicious login attempts..."
- Bullet points with findings
- Action buttons: [Show Details] [Block These IPs]
Footer: Input bar with send button

Shows co-pilot mode where AI is contextually aware.
Light mode, desktop. No device frame.
```

### 6.11 Empty States

```
Generate a UI mockup showing empty state designs for admin dashboard.

SHOW 4 EMPTY STATES in a 2x2 grid:

1. No Users Found:
   - Large user icon (64px, muted)
   - "No users found" heading
   - "Try adjusting your filters or add a new user"
   - [+ Add Your First User] primary button

2. No Audit Logs:
   - Document icon
   - "No activity recorded"
   - "System activity will appear here"

3. No Search Results:
   - Search icon with X
   - "No results for 'xyz'"
   - "Try different keywords"
   - [Clear Search] link

4. Empty Permission Matrix:
   - Grid icon
   - "No roles configured"
   - "Create your first role to get started"
   - [+ Create Role] button

Centered content, muted colors, helpful CTAs.
Light mode. No device frame.
```

### 6.12 Bulk Selection Action Bar

```
Generate a UI mockup of a bulk action floating bar for data tables.

CONTEXT: Table with rows selected

FLOATING BAR (appears at bottom of table):
- Left: "☑ 5 users selected" text
- Center: Action buttons
  - [Change Role ▼] dropdown
  - [Deactivate] button
  - [Delete] danger button
- Right: [✕] dismiss button

BAR STYLING:
- 64px height
- Floating with shadow-lg
- Slight rounded corners
- Background: white (light) or slate-800 (dark)
- Appears with slide-up animation

Show bar floating over bottom portion of table.
Light mode. No device frame.
```

### 6.13 Date Range Picker

```
Generate a UI mockup of a date range picker dropdown.

DROPDOWN STRUCTURE (400px wide):

Left Column - Quick Select:
○ Today
○ Yesterday
○ Last 7 days
○ Last 30 days
○ This month
○ Last month
● Custom range

Right Area - Custom Range:
From: [Jan 15, 2026] [📅]
To: [Jan 19, 2026] [📅]

(When custom selected, show dual calendar grid)

Footer:
- Cancel button
- Apply button

Modern date picker UI, indigo accent for selected dates.
Light mode. No device frame.
```

### 6.14 Command Menu / Global Search

```
Generate a UI mockup of a command palette / global search modal.

MODAL: Centered, 560px wide

SEARCH BAR:
- 🔍 icon left
- "Type a command or search..." placeholder
- ⌘K keyboard shortcut badge right

RESULTS GROUPED:
Recent:
- Dashboard (Clock icon)
- User: John Doe (User icon)

Quick Actions:
- + Create New User (Plus icon)
- + Create New Role (Shield icon)
- 📤 Export Audit Logs (Download icon)

Navigation:
- Go to Users (Arrow icon)
- Go to Roles (Arrow icon)
- Go to Audit Logs (Arrow icon)

Each item: Icon + label + optional keyboard shortcut right-aligned
Hover state: Slate-100 background

CMD+K modal style (like Raycast/Spotlight).
Light mode. No device frame.
```

### 6.15 Toast Notifications

```
Generate a UI mockup showing toast notification variants.

LAYOUT: Stack of 4 toasts in bottom-right corner

TOAST VARIANTS (each ~360px wide):
1. Success Toast:
   - ✓ Green checkmark icon
   - "User created successfully"
   - X dismiss button
   - Green left accent border

2. Error Toast:
   - ✗ Red X icon
   - "Failed to update role"
   - "Permission denied" secondary text
   - [Retry] action button
   - Red left accent border

3. Warning Toast:
   - ⚠ Amber warning icon
   - "Session expires in 5 minutes"
   - [Extend] action button
   - Amber left accent border

4. Info Toast:
   - ℹ Blue info icon
   - "Export completed"
   - [Download] action link
   - Blue left accent border

Show progress bar animation on one toast (auto-dismiss countdown).
Stacked with 8px gap. Light mode.
```

---

## 📊 SECTION 7: PROMPT COMBINATIONS

Combine these for complex outputs:

### Full Page + Component Detail

```
Generate two images:
1. Full User Management page mockup (1440x900)
2. Zoomed detail of Add User modal (component view)
```

### Light + Dark Comparison

```
Generate side-by-side comparison:
Left: Dashboard in Light Mode
Right: Same dashboard in Dark Mode ("Eclipse Theme")
```

### Desktop + Mobile Responsive

```
Generate responsive views:
1. Full desktop dashboard (1440x900)
2. Tablet view (768px width)
3. Mobile view (393px width, with bottom nav)
```

### State Progression

```
Generate a sequence showing:
1. Empty state (no users)
2. Loading state (skeleton)
3. Loaded state (5 users)
4. Error state (failed to load)
```

---

_Prompt Library for Nexus Design System v1.0_  
_Total: 35+ Ready-to-Use Prompts_  
_Optimized for Gemini 3, Gemini 2.0 Flash, and similar AI image generators_
