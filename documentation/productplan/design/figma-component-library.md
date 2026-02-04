# NexusOS Figma Component Library Guide

**Version:** 1.0
**Last Updated:** January 19, 2026
**Target Audience:** UI Designers, Product Designers, Developers
**Figma Plugin Requirements:** Token Studio, Figma Variables, Auto Layout

---

## 1. Library Architecture

### 1.1 File Structure

Create a Figma library with the following page structure:

```
📁 NexusOS Design System
├── 📄 Cover
├── 📄 Getting Started
├── 📄 Design Tokens
│   ├── 🎨 Colors
│   ├── 📐 Spacing
│   ├── 🔘 Radius
│   ├── 🔤 Typography
│   └── 🌑 Shadows
├── 📄 Atoms
│   ├── Button
│   ├── Input
│   ├── Badge
│   ├── Avatar
│   ├── Checkbox
│   ├── Radio
│   ├── Switch
│   ├── Icon
│   └── Divider
├── 📄 Molecules
│   ├── Form Field
│   ├── Search Bar
│   ├── Toast
│   ├── Dropdown
│   ├── Tab
│   └── Pagination
├── 📄 Organisms
│   ├── Hyper-Grid (Data Table)
│   ├── AI Chat Widget
│   ├── Metric Card
│   ├── Sidebar
│   ├── Navbar
│   └── Modal
├── 📄 Templates
│   ├── Dashboard Layout
│   ├── Auth Layout
│   ├── List Page Layout
│   └── Detail Page Layout
└── 📄 Example Screens
    ├── Dashboard (Comfort)
    ├── Dashboard (Compact)
    ├── User List
    ├── Permission Matrix
    ├── Audit Logs
    └── Login
```

---

## 2. Figma Variables Setup

### 2.1 Variable Collections

Create these Variable Collections with Modes:

#### Collection: "Theme"

| Mode      | Description                |
| :-------- | :------------------------- |
| **Light** | Light mode colors          |
| **Dark**  | Dark mode (Eclipse) colors |

#### Collection: "Density"

| Mode        | Description                  |
| :---------- | :--------------------------- |
| **Comfort** | SaaS-friendly spacing/sizing |
| **Compact** | Enterprise data-dense mode   |

### 2.2 Color Variables (Theme Collection)

```
📁 Semantic Colors
├── background
├── surface
├── surface-hover
├── foreground
├── muted-foreground
├── border
├── border-strong
├── primary
├── primary-foreground
├── secondary
├── secondary-foreground
├── accent
├── info
├── info-subtle
├── success
├── success-subtle
├── warning
├── warning-subtle
├── danger
├── danger-subtle
```

**Value Mapping:**

| Variable               | Light Mode | Dark Mode |
| :--------------------- | :--------- | :-------- |
| `background`           | #FFFFFF    | #020617   |
| `surface`              | #F8FAFC    | #0F172A   |
| `surface-hover`        | #F1F5F9    | #1E293B   |
| `foreground`           | #0F172A    | #F8FAFC   |
| `muted-foreground`     | #64748B    | #94A3B8   |
| `border`               | #E2E8F0    | #1E293B   |
| `primary`              | #6366F1    | #818CF8   |
| `primary-foreground`   | #FFFFFF    | #0F172A   |
| `secondary`            | #14B8A6    | #2DD4BF   |
| `secondary-foreground` | #FFFFFF    | #0F172A   |
| `accent`               | #8B5CF6    | #A78BFA   |
| `info`                 | #3B82F6    | #60A5FA   |
| `success`              | #10B981    | #34D399   |
| `warning`              | #F59E0B    | #FBBF24   |
| `danger`               | #DC2626    | #EF4444   |

### 2.3 Spacing Variables (Density Collection)

| Variable             | Comfort | Compact |
| :------------------- | :------ | :------ |
| `layout-padding`     | 32px    | 16px    |
| `card-padding`       | 24px    | 12px    |
| `component-gap`      | 16px    | 8px     |
| `input-padding-y`    | 10px    | 4px     |
| `input-padding-x`    | 16px    | 12px    |
| `table-cell-padding` | 16px    | 6px     |

### 2.4 Sizing Variables (Density Collection)

| Variable           | Comfort | Compact |
| :----------------- | :------ | :------ |
| `button-height`    | 44px    | 32px    |
| `input-height`     | 44px    | 32px    |
| `table-row-height` | 64px    | 36px    |
| `icon-size`        | 20px    | 16px    |
| `sidebar-width`    | 280px   | 72px    |
| `navbar-height`    | 80px    | 56px    |

### 2.5 Radius Variables (Density Collection)

| Variable    | Comfort | Compact |
| :---------- | :------ | :------ |
| `radius-sm` | 6px     | 2px     |
| `radius-md` | 8px     | 4px     |
| `radius-lg` | 12px    | 6px     |
| `radius-xl` | 16px    | 8px     |

---

## 3. Component Specifications

### 3.1 Button Component

**Variants:**

| Property   | Values                                        |
| :--------- | :-------------------------------------------- |
| `variant`  | primary, secondary, ghost, destructive, magic |
| `size`     | sm, md, lg                                    |
| `state`    | default, hover, pressed, disabled, loading    |
| `iconOnly` | true, false                                   |

**Auto Layout Settings:**

- Padding: `$spacing/input-padding-y` × `$spacing/button-padding-x`
- Gap: `$spacing/2` (between icon and text)
- Height: Fixed to `$sizing/button-height`
- Corner Radius: `$radius/radius-md`

**Variant-Specific Styles:**

| Variant     | Fill                     | Stroke    | Text                    |
| :---------- | :----------------------- | :-------- | :---------------------- |
| primary     | `$primary`               | none      | `$primary-foreground`   |
| secondary   | `$secondary`             | none      | `$secondary-foreground` |
| outline     | transparent              | `$border` | `$foreground`           |
| ghost       | transparent              | none      | `$foreground`           |
| destructive | `$danger`                | none      | white                   |
| magic       | gradient (Indigo→Violet) | gradient  | white                   |

### 3.2 Input Component

**Variants:**

| Property          | Values                              |
| :---------------- | :---------------------------------- |
| `state`           | default, focus, error, disabled     |
| `type`            | text, password, email, number, date |
| `hasLeadingIcon`  | true, false                         |
| `hasTrailingIcon` | true, false                         |
| `hasAiButton`     | true, false                         |

**Auto Layout Settings:**

- Height: `$sizing/input-height`
- Padding: `$spacing/input-padding-y` × `$spacing/input-padding-x`
- Corner Radius: `$radius/radius-md`
- Border: 1px `$border`

**State Colors:**

| State    | Border            | Background               |
| :------- | :---------------- | :----------------------- |
| default  | `$border`         | `$background`            |
| focus    | `$primary` + ring | `$background`            |
| error    | `$danger`         | `$background`            |
| disabled | `$border`         | `$surface` (opacity 50%) |

### 3.3 Badge Component

**Variants:**

| Property  | Values                                           |
| :-------- | :----------------------------------------------- |
| `variant` | neutral, success, warning, danger, info, primary |
| `style`   | solid, subtle, outline                           |
| `size`    | sm, md                                           |

**Subtle Style Formula:**

| Variant | Background        | Text       |
| :------ | :---------------- | :--------- |
| success | `$success` at 10% | `$success` |
| warning | `$warning` at 10% | `$warning` |
| danger  | `$danger` at 10%  | `$danger`  |
| primary | `$primary` at 10% | `$primary` |

### 3.4 Data Table (Hyper-Grid)

**Component Structure:**

```
📦 HyperGrid
├── 📦 Toolbar
│   ├── Search Input
│   ├── Filter Button
│   ├── Columns Button
│   ├── Density Toggle
│   └── Primary Action Button
├── 📦 Table Header
│   └── 📦 Header Cell (repeating)
│       ├── Label
│       ├── Sort Icon
│       └── Resize Handle
├── 📦 Table Body
│   └── 📦 Table Row (repeating)
│       ├── Checkbox Cell
│       ├── Data Cells
│       └── Actions Cell
├── 📦 Bulk Actions Bar (conditional)
│   ├── Selected Count
│   └── Action Buttons
└── 📦 Pagination Footer
    ├── Rows per Page
    ├── Results Summary
    └── Page Navigation
```

**Density Variations:**

| Element      | Comfort         | Compact   |
| :----------- | :-------------- | :-------- |
| Row Height   | 64px            | 36px      |
| Cell Padding | 16px            | 6px       |
| Font Size    | 14px            | 13px      |
| Border       | horizontal only | full grid |
| Header       | normal case     | uppercase |

### 3.5 Sidebar Component

**Variants:**

| Property | Values              |
| :------- | :------------------ |
| `state`  | expanded, collapsed |
| `theme`  | light, dark         |

**Structure (Expanded):**

```
📦 Sidebar (280px)
├── 📦 Logo Area
│   ├── Logo
│   └── Product Name
├── 📦 Search
├── 📦 Nav Group
│   ├── Group Label
│   └── 📦 Nav Item (repeating)
│       ├── Icon
│       ├── Label
│       ├── Badge (optional)
│       └── Chevron (if has children)
├── 📦 Divider
├── 📦 Nav Group (Secondary)
└── 📦 Footer
    ├── Collapse Toggle
    └── AI Chat Button
```

**Structure (Collapsed/Rail - 72px):**

```
📦 Sidebar Rail
├── 📦 Logo Icon
├── 📦 Nav Item (icon only)
│   └── Tooltip on hover
└── 📦 Footer Icons
```

---

## 4. Creating Adaptive Components

### 4.1 Mode-Switching Setup

To make components switch between Comfort/Compact modes:

1. **Select the component frame**
2. **Apply density variables to all number properties:**
   - Replace `44px` with `$sizing/button-height`
   - Replace `16px` padding with `$spacing/component-gap`
   - Replace `12px` radius with `$radius/radius-lg`

3. **Create variants for each mode** (if visual changes are not just numerical)

### 4.2 Theme-Switching Setup

For Light/Dark mode:

1. **Apply color variables to all color properties:**
   - Replace `#FFFFFF` with `$background`
   - Replace `#0F172A` with `$foreground`
   - Replace border colors with `$border`

2. **Shadow behavior:**
   - Light mode: Apply `$shadow/md`
   - Dark mode: Remove shadow, add `border: 1px $border`

### 4.3 Interactive States

Use Figma's **Interactive Components** for:

- Button hover, pressed, loading states
- Input focus states
- Checkbox/Radio toggle
- Dropdown open/close
- Sidebar expand/collapse

---

## 5. Design-to-Dev Handoff

### 5.1 Token Export with Token Studio

**Setup:**

1. Install Token Studio plugin
2. Connect to your design-tokens.json repository
3. Sync Variables ↔ Tokens

**Export Format:**

```json
{
  "color": {
    "semantic": {
      "primary": { "$value": "{color.indigo.600}" }
    }
  }
}
```

### 5.2 Component Documentation

Each component page should include:

1. **Overview** - What the component is for
2. **Variants Table** - All property combinations
3. **Anatomy** - Labeled breakdown of parts
4. **States** - All interactive states
5. **Spacing Specs** - Padding, margins, gaps
6. **Usage Guidelines** - Do's and Don'ts
7. **Accessibility Notes** - ARIA, keyboard nav

### 5.3 Developer Handoff Checklist

For each component, provide:

- [ ] Figma component with all variants
- [ ] Design tokens mapped to CSS variables
- [ ] State transition specifications
- [ ] Responsive behavior notes
- [ ] Accessibility requirements
- [ ] Code snippet (if available)

---

## 6. Quality Checklist

### 6.1 Component Quality

- [ ] Uses Variables (not hardcoded values)
- [ ] Auto Layout enabled
- [ ] All variants created
- [ ] All states created (hover, focus, disabled, etc.)
- [ ] Properly named layers
- [ ] Description filled in
- [ ] Constraints set correctly

### 6.2 Library Quality

- [ ] All components published to library
- [ ] Version history documented
- [ ] Change log maintained
- [ ] Thumbnail/cover image set
- [ ] Getting Started page complete

### 6.3 Accessibility Quality

- [ ] Touch targets ≥ 44px (Comfort) / 32px (Compact)
- [ ] Color contrast verified (4.5:1 minimum)
- [ ] Focus states designed
- [ ] Error states include non-color indicators

---

## 7. Version Control

### 7.1 Semantic Versioning

| Version           | When to Update                         |
| :---------------- | :------------------------------------- |
| **Major (X.0.0)** | Breaking changes, new design direction |
| **Minor (1.X.0)** | New components, new variants           |
| **Patch (1.0.X)** | Bug fixes, small tweaks                |

### 7.2 Branch Strategy

| Branch          | Purpose                      |
| :-------------- | :--------------------------- |
| **Main**        | Production-ready components  |
| **Development** | Work in progress             |
| **Feature/**    | Individual component updates |

### 7.3 Change Log Format

```markdown
## [1.1.0] - 2026-01-25

### Added

- New DatePicker component
- Loading skeleton variants for all cards

### Changed

- Updated primary color from indigo-600 to indigo-500
- Increased button padding in Compact mode

### Fixed

- Badge text alignment issue in Dark mode
- Sidebar tooltip z-index conflict
```

---

## 8. Collaboration Guidelines

### 8.1 For Designers

1. **Never detach instances** - Always use library components
2. **Use Variables for overrides** - Don't hardcode colors/spacing
3. **Follow naming conventions** - `Component/Variant/State`
4. **Document changes** - Add descriptions to modified components

### 8.2 For Developers

1. **Reference Figma for specs** - Don't guess values
2. **Use design-tokens.json** - Import tokens, don't hardcode
3. **Report discrepancies** - Flag any spec issues
4. **Validate implementation** - Compare side-by-side with Figma

### 8.3 Communication

- Weekly design sync with development team
- Figma comments for component feedback
- Slack channel for quick questions
- Design review before major releases

---

_Figma Component Library Guide for Nexus Design System v1.0_
_Enabling seamless designer-developer collaboration_
