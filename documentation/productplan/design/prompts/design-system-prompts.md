# NexusOS Design System - Gemini Image Generation Prompts

**Purpose:** Ordered prompts to generate complete Design System documentation visuals  
**AI Target:** Google Gemini 3 / Imagen  
**Generate Order:** Follow sequence 1 → 10 for complete design system documentation

---

## 💎 MASTER STYLE SEED (PENTING AGAR KONSISTEN)

**Copy block ini dan paste SETIAP KALI sebelum paste prompt nomor 1-10.**
Ini memastikan gaya visual (Camera, Lighting, Vibe) selalu seragam 100%.

```text
Global Style Parameters:
- STYLE: Professional UI Documentation, "Apple Human Interface Guidelines" aesthetic, High Fidelity Vector.
- CAMERA: Orthographic Top-Down View (90° angle), Flat Lay. No perspective distortion.
- LIGHTING: Soft diffused studio lighting, purely white light (5000K), no colored shadows.
- QUALITY: 4K resolution, razor sharp edges, sub-pixel perfect precision.
- BACKGROUND: Pure white (#FFFFFF) for Light Mode sections, Deep Slate (#020617) for Dark Mode sections.
- OBJECTS: Floating UI elements with subtle "Shadow-MD" elevation.
- CONSTRAINT: NO 3D rotation, NO tilted angles, NO blurry depth-of-field. Keep everything flat and readable.
```

---

## 📋 GENERATION SEQUENCE

| Order | Category            | Description                             |
| ----- | ------------------- | --------------------------------------- |
| 1     | Color Palette       | Complete color system with all variants |
| 2     | Typography Scale    | All text sizes and headings             |
| 3     | Spacing & Grid      | Spacing tokens and grid system          |
| 4     | Border Radius       | All radius variants                     |
| 5     | Shadows & Elevation | Shadow scale light/dark                 |
| 6     | Buttons             | All button variants and states          |
| 7     | Form Elements       | Inputs, checkboxes, toggles             |
| 8     | Badges & Status     | All semantic color badges               |
| 9     | Cards & Containers  | Card variants light/dark                |
| 10    | Dividers & Lines    | Divider styles                          |

---

## 🎨 PROMPT 1: COLOR PALETTE (Complete)

```
Generate a professional Design System color palette documentation image.

LAYOUT: Clean design documentation style, white canvas, organized sections

SECTION A - PRIMARY COLORS:
Show color swatches in a horizontal row with labels:
- Indigo 50: #EEF2FF
- Indigo 100: #E0E7FF
- Indigo 200: #C7D2FE
- Indigo 300: #A5B4FC
- Indigo 400: #818CF8 (Dark Mode Primary)
- Indigo 500: #6366F1 ⭐ (Light Mode Primary)
- Indigo 600: #4F46E5
- Indigo 700: #4338CA
- Indigo 800: #3730A3
- Indigo 900: #312E81

SECTION B - SECONDARY COLORS:
Teal palette swatches:
- Teal 300: #5EEAD4
- Teal 400: #2DD4BF (Dark Mode Secondary)
- Teal 500: #14B8A6 ⭐ (Light Mode Secondary)
- Teal 600: #0D9488

SECTION C - ACCENT COLORS:
Violet palette swatches:
- Violet 400: #A78BFA (Dark Mode Accent)
- Violet 500: #8B5CF6 ⭐ (Light Mode Accent)
- Violet 600: #7C3AED

SECTION D - NEUTRAL COLORS (Slate):
Full slate palette:
- Slate 50: #F8FAFC (Surface Light)
- Slate 100: #F1F5F9
- Slate 200: #E2E8F0 (Border Light)
- Slate 300: #CBD5E1
- Slate 400: #94A3B8 (Muted Text Dark)
- Slate 500: #64748B (Muted Text Light)
- Slate 600: #475569
- Slate 700: #334155
- Slate 800: #1E293B (Border Dark)
- Slate 900: #0F172A (Surface Dark)
- Slate 950: #020617 (Background Dark)

SECTION E - SEMANTIC COLORS:
Row of semantic color pairs (Light / Dark):
- Success: #10B981 / #34D399 (Emerald)
- Warning: #F59E0B / #FBBF24 (Amber)
- Error: #DC2626 / #EF4444 (Red)
- Info: #3B82F6 / #60A5FA (Blue)

Each swatch should be 80x80px with hex code label below.
Title: "NexusOS Nebula Palette"
Clean, professional design documentation style.
Image size: 1920x1080
```

---

## 🔤 PROMPT 2: TYPOGRAPHY SCALE

```
Generate a professional Design System typography documentation image.

LAYOUT: Split view - Left: Light Mode (White Bg), Right: Dark Mode (Slate-950 Bg)

HEADER: "NexusOS Typography Scale"

SECTION A - DISPLAY & HEADINGS (Show on both backgrounds):
- Display: 36px / Bold / Line-height 1.2
- H1: 24px / SemiBold / Line-height 1.2
- H2: 20px / SemiBold / Line-height 1.3
- H3: 18px / Medium / Line-height 1.4

SECTION B - BODY TEXT (Comfort Mode):
- Body Large: 16px / Regular / Line-height 1.6
- Body: 14px / Regular / Line-height 1.6
- Small: 13px / Regular / Line-height 1.5
- Caption: 12px / Medium / Line-height 1.4
- Tiny: 11px / Medium / Line-height 1.3

SECTION C - BODY TEXT (Compact Mode):
- Body: 13px / Regular / Line-height 1.3
- Small: 12px / Regular / Line-height 1.2
- Caption: 11px / Medium / Line-height 1.1
- Light Mode: Slate-900 (Headings), Slate-600 (Body)
- Dark Mode: Slate-50 (Headings), Slate-400 (Body)

SECTION D - MONOSPACE (Geist Mono):
- Code: 13px / Regular — "const data = getData();"
- Data: 14px / Medium — "1,234,567.89"

SECTION E - FONT WEIGHTS:
Show "Aa" in each weight:
- Regular (400)
- Medium (500)
- SemiBold (600)
- Bold (700)

Each style shows: Style name, Sample text "The quick brown fox"
Clean design documentation style.
Image size: 1920x1080
```

---

## 📐 PROMPT 3: SPACING & GRID SYSTEM

```
Generate a professional Design System spacing documentation image.

LAYOUT: White canvas with dark mode inset section

HEADER: "NexusOS Spacing System"

SECTION A - BASE SPACING SCALE (Light Mode):
Show horizontal bars with pixel labels (Indigo filled):
- space-1 (4px) to space-16 (64px)

SECTION B - UI CONTEXT (Split View):
1. LIGHT CARD (Comfort):
   - 32px padding
   - 16px gap between elements
   - White bg, Shadow-md

2. DARK CARD (Compact):
   - 16px padding
   - 8px gap between elements
   - Slate-900 bg, 1px Slate-800 border

both should have Visual boxes showing:
- Layout Padding: 32px (large outer box)
- Card Padding: 24px (medium box)
- Component Gap: 16px (space between elements)
- Input Padding: 10px × 16px (vertical × horizontal)
- Table Cell: 16px

SECTION C - COMPACT MODE SPACING:
Same boxes but tighter:
- Layout Padding: 16px
- Card Padding: 12px
- Component Gap: 8px
- Input Padding: 4px × 12px
- Table Cell: 6px

SECTION D - GRID SYSTEM:
Show 12-column grid with gutters:
- Desktop: 12 columns, 24px gutter
- Tablet: 8 columns, 16px gutter
- Mobile: 4 columns, 16px gutter

SECTION E - GRID:
Visual overlay showing 12-column grid (Indigo translucent columns).

Clean design documentation style.
Image size: 1920x1080
```

---

## 🔘 PROMPT 4: BORDER RADIUS

```
Generate a professional Design System border radius documentation image.

LAYOUT: Split view - Left: Comfort (Light), Right: Compact (Dark)

HEADER: "NexusOS Border Radius"

SCTION A - LEFT SIDE - COMFORT (Rounded/SaaS):
Show visuals with Soft Radius:
- Buttons: 8px (radius-md)
- Cards: 16px (radius-xl)
- Badges: 9999px (Pill)
- Style: White shapes on Slate-100 bg

RIGHT SIDE - COMPACT (Sharp/Enterprise):
Show visuals with Tight Radius:
- Buttons: 4px (radius-sm)
- Cards: 6px (radius-lg)
- Badges: 4px (Rectangle)
- Style: Slate-900 shapes on Slate-950 bg

SECTION B - COMFORT MODE (SaaS):
Show rounded rectangles demonstrating each radius:
- radius-sm: 6px (small button, badge)
- radius-md: 8px (inputs, buttons)
- radius-lg: 12px (cards, panels)
- radius-xl: 16px (large cards, modals)
- radius-full: 9999px (pills, avatars)

SECTION C - COMPACT MODE (Enterprise):
Show same rectangles with sharper corners:
- radius-sm: 2px
- radius-md: 4px
- radius-lg: 6px
- radius-xl: 8px
- radius-full: 9999px

SECTION D - USAGE EXAMPLES:
Side-by-side comparison:
- Button Comfort (8px) vs Button Compact (2px)
- Card Comfort (16px) vs Card Compact (4px)
- Badge Comfort (9999px pill) vs Badge Compact (4px rectangle)
- Input Comfort (8px) vs Input Compact (4px)

Use indigo (#6366F1) fill for the shapes.
Each shape labeled with pixel value.
Clean design documentation style.
Image size: 1440x900
```

---

## 🌑 PROMPT 5: SHADOWS & ELEVATION

```
Generate a professional Design System shadow documentation image.

LAYOUT: Split view - Left: Light Mode, Right: Dark Mode

HEADER: "NexusOS Shadow System"

LEFT SIDE - LIGHT MODE:
Show white cards on light gray (#F1F5F9) background with different shadows:
- None: No shadow (flat)
- XS: 0 1px 2px rgba(15,23,42,0.05) — very subtle
- SM: 0 1px 3px rgba(15,23,42,0.08) — small elements
- MD: 0 4px 6px rgba(15,23,42,0.08) — default cards ⭐
- LG: 0 10px 15px rgba(15,23,42,0.08) — dropdowns
- XL: 0 20px 25px rgba(15,23,42,0.08) — modals

RIGHT SIDE - DARK MODE:
Show slate-900 (#0F172A) cards on slate-950 (#020617) background:
- No shadows in dark mode
- Instead: 1px border using Slate-800 (#1E293B)
- Inner glow: ring-1 ring-white/5% (subtle white inner ring)

SECTION B - ELEVATION LAYERS:
Stack of cards showing z-index layers:
- Base (z-0): Page content
- Dropdown (z-10): Menus
- Sticky (z-20): Headers
- Fixed (z-30): Sidebars
- Overlay (z-40): Backdrop
- Modal (z-50): Dialogs
- Popover (z-60): Tooltips
- Toast (z-70): Notifications

Clean design documentation style.
Image size: 1920x1080
```

---

## 🔲 PROMPT 6: BUTTON COMPONENT (All Variants)

```
Generate a professional Design System button documentation image.

LAYOUT: White canvas with organized grid. Have Dark mode and Light mode

HEADER: "NexusOS Buttons"

S
SECTION A - VARIANTS (Default State):
Row of buttons showing each variant:
- Primary: Solid indigo (#6366F1), white text
- Secondary: Solid teal (#14B8A6), white text
- Outline: Transparent, slate border, dark text
- Ghost: Transparent, no border, dark text
- Destructive: Solid red (#DC2626), white text
- Magic: Gradient border (indigo→violet), sparkle icon ✨

SECTION B - SIZES:
Show Primary button in 3 sizes:
- Small (SM): 32px height, 12px padding
- Medium (MD): 36px height, 16px padding
- Large (LG): 44px height, 20px padding

SECTION C - STATES:
Show Primary button in all states:
- Default
- Hover (slightly lighter)
- Pressed/Active (slightly darker)
- Focus (2px indigo ring)
- Disabled (50% opacity)
- Loading (spinner icon replacing text)

SECTION D - ICON VARIANTS:
- Icon Left + Text
- Text + Icon Right
- Icon Only (square button)

SECTION E - COMFORT vs COMPACT:
Side-by-side:
- Comfort: 44px height, 8px radius
- Compact: 32px height, 2px radius

SECTION F - LIGHT MODE:
Background: White
Buttons:
- Primary: Solid Indigo (#6366F1)
- Secondary: Solid Teal (#14B8A6)
- Outline: White toggle with Slate border
- Destructive: Red (#DC2626)
- Magic: Gradient border

SECTION H - DARK MODE:
Background: Slate-950 (#020617)
Buttons:
- Primary: Lighter Indigo (#818CF8)
- Secondary: Teal (#2DD4BF)
- Outline: Transparent with Slate-700 border, White text
- Ghost: Hover State with Slate-800 bg

Show Default, Hover, and Disabled states for all.

Clean design documentation style.
Image size: 1920x1080

```

---

## 📝 PROMPT 7: FORM ELEMENTS

```
Generate a professional Design System form elements documentation image.

LAYOUT: White canvas with organized sections

HEADER: "NexusOS Form Elements"

SECTION A - TEXT INPUTS:
Show input fields with labels:
- Default: White bg, 1px slate-200 border
- Focus: 2px indigo border with blue glow ring
- Error: Red border, red helper text "This field is required"
- Disabled: Gray bg, 50% opacity
- With Icon Left (email icon)
- With AI Button (sparkle icon right side)

SECTION B - INPUT SIZES:
- Comfort: 44px height, 16px padding
- Compact: 32px height, 12px padding

SECTION C - CHECKBOXES:
- Unchecked: White bg, slate border
- Checked: Indigo bg, white checkmark
- Indeterminate: Indigo bg, minus sign
- Disabled: Gray, 50% opacity

SECTION D - RADIO BUTTONS:
- Unselected: White bg, slate border
- Selected: Indigo filled dot
- Disabled: Gray

SECTION E - TOGGLE SWITCHES:
- Off: Slate-200 track, white thumb
- On: Indigo track, white thumb
- Disabled: Gray

SECTION F - SELECT DROPDOWN:
- Closed state with chevron
- Open state with options highlighted

SECTION G - LIGHT MODE:
- Input Default: White bg, Slate-200 border
- Input Focus: Indigo ring, blue glow
- Checkbox Checked: Indigo bg
- Toggle On: Indigo track

SECTION H - DARK MODE:
Background: Slate-950 (#020617)
- Input Default: Slate-900 bg, Slate-800 border
- Input Focus: Indigo-500 border, no glow
- Checkbox Checked: Indigo-500 bg
- Toggle On: Indigo-500 track

Show Inputs, Checkboxes, Radios, Toggles in both modes.

Clean design documentation style.
Image size: 1920x1080
```

---

## 🏷️ PROMPT 8: BADGES & STATUS INDICATORS

```
Generate a professional Design System badges documentation image.

LAYOUT: White canvas, organized grid, standardizing Light vs Dark mode.

HEADER: "NexusOS Badges & Status"

SECTION A - SEMANTIC BADGES (Subtle Style):
Row of badges with 10% opacity background:
- Success: Emerald bg, "Active" + checkmark
- Warning: Amber bg, "Pending" + clock
- Error: Red bg, "Failed" + x-circle
- Info: Blue bg, "Processing" + info icon
- Neutral: Slate bg, "Draft"
- Primary: Indigo bg, "New"
- Secondary: Teal bg, "Updated"

SECTION B - STYLE VARIANTS:
Show Success badge in 3 styles:
- Subtle: 10% opacity bg, colored text (recommended)
- Solid: Full color bg, white text
- Outline: Transparent bg, colored border

SECTION C - SHAPE VARIANTS:
- Pill (Comfort): border-radius 9999px
- Rectangle (Compact): border-radius 4px

SECTION D - SIZES:
- Small: 20px height, 12px font
- Medium: 24px height, 13px font

SECTION E - STATUS DOTS:
Simple colored dots for inline status:
- 🟢 Online (Emerald)
- 🟡 Away (Amber)
- 🔴 Offline (Red)
- 🔵 Busy (Blue)

SECTION F - ACTION BADGES (Dark Mode):
Same badges on dark (#020617) background with brighter color - DARK MODE (Subtle Style):
Background: Slate-950
- Success: Emerald-900/50 bg, Emerald-300 text
- Warning: Amber-900/50 bg, Amber-300 text
- Error: Red-900/50 bg, Red-300 text
- Info: Blue-900/50 bg, Blue-300 text

Clean design documentation style.
Image size: 1440x900
```

---

## 🃏 PROMPT 9: CARDS & CONTAINERS

```
Generate a professional Design System cards documentation image.

LAYOUT: Split view - Top: Light Mode, Bottom: Dark Mode

HEADER: "NexusOS Cards & Panels"

LIGHT MODE SECTION:
Show cards on light gray (#F8FAFC) background:

Card Type A - Default:
- White background
- Shadow-md (0 4px 6px rgba(15,23,42,0.08))
- No border OR 1px slate-100 border
- 24px padding (Comfort) / 12px (Compact)
- 16px radius (Comfort) / 4px (Compact)

Card Type B - KPI Metric Card:
- Large number "1,234"
- Label "Total Users"
- Trend arrow "+12%"
- Icon in pastel circle

Card Type C - Elevated (Popover/Dropdown):
- Shadow-lg
- Appears floating above content

DARK MODE SECTION:
Show cards on slate-950 (#020617) background:

Card Type A - Default:
- Slate-900 (#0F172A) background
- NO shadow
- 1px border slate-800 (#1E293B)
- Inner glow: ring-1 ring-white/5%

Card Type B - KPI Metric Card (Dark):
- Same content, adapted colors
- Slightly brighter text

Card Type C - Interactive:
- Hover state with slate-800 background

Clean design documentation style.
Image size: 1920x1080
```

---

## ➖ PROMPT 10: DIVIDERS & LINES

```
Generate a professional Design System dividers documentation image.

LAYOUT: White canvas with organized sections

HEADER: "NexusOS Dividers & Lines"

SECTION A - HORIZONTAL DIVIDERS:
Show different divider styles:
- Default: 1px solid slate-200 (#E2E8F0)
- Strong: 1px solid slate-300 (#CBD5E1)
- Subtle: 1px solid slate-100 (#F1F5F9)
- Dashed: 1px dashed slate-200
- Branded: 2px solid indigo-500

SECTION B - VERTICAL DIVIDERS:
Same styles but vertical orientation:
- In toolbar between button groups
- In sidebar between nav sections

SECTION C - DARK MODE DIVIDERS:
On dark background (#020617):
- Default: 1px solid slate-800 (#1E293B)
- Strong: 1px solid slate-700 (#334155)
- Subtle: 1px solid slate-900 (#0F172A)

SECTION D - DIVIDER WITH LABEL:
- Horizontal line with centered text "OR"
- Horizontal line with left-aligned text "Section"

SECTION E - CARD INTERNAL DIVIDERS:
- Show card with content sections separated by dividers
- Header divider
- Footer divider

SECTION F - TABLE BORDERS:
- Comfort Mode: Horizontal borders only
- Compact Mode: Full grid (horizontal + vertical)

Clean design documentation style.
Image size: 1440x900
```

---

## 🎬 BONUS PROMPT: ANIMATION & TRANSITIONS

```
Generate a professional Design System animation documentation image.

LAYOUT: White canvas with organized sections

HEADER: "NexusOS Motion & Animation"

SECTION A - DURATION SCALE:
Show timeline bars:
- Instant: 0ms (no animation)
- Fast: 100ms (micro-interactions)
- Normal: 200ms (default transitions) ⭐
- Slow: 300ms (complex animations)
- Slower: 500ms (page transitions)

SECTION B - EASING CURVES:
Show bezier curve graphs:
- Default: cubic-bezier(0.4, 0, 0.2, 1) - smooth
- Ease In: cubic-bezier(0.4, 0, 1, 1) - accelerate
- Ease Out: cubic-bezier(0, 0, 0.2, 1) - decelerate
- Spring: cubic-bezier(0.175, 0.885, 0.32, 1.275) - bounce

SECTION C - COMMON TRANSITIONS:
Visual examples:
- Hover: Color change (200ms)
- Focus: Ring appear (100ms)
- Dropdown: Slide down (200ms)
- Modal: Fade + Scale (300ms)
- Sidebar: Slide (300ms)
- Toast: Slide in (200ms)

SECTION D - LOADING STATES:
- Skeleton: Shimmer gradient animation
- Spinner: Rotating circle
- AI Processing: Pulse glow effect
- Progress Bar: Fill animation

SECTION E - REDUCED MOTION:
Note: "Respect prefers-reduced-motion: reduce"
- Replace animations with instant transitions
- Keep essential feedback only

Clean design documentation style.
Image size: 1440x900
```

---

## 📖 HOW TO USE

### Step 1: Copy Master Context

First, copy this context before any prompt:

```
DESIGN SYSTEM: NexusOS Nexus Design System v1.0
STYLE: Professional design documentation / style guide
AESTHETIC: Clean, minimal, Figma-style documentation
OUTPUT: High-quality image for design reference
```

### Step 2: Generate in Order

Run prompts 1-10 in sequence for complete documentation.

### Step 3: Combine Results

Use generated images to:

- Create Figma design system file
- Build Storybook documentation
- Share with development team
- Use as visual reference

---

_Design System Generation Prompts for NexusOS v1.0_
_Optimized for Google Gemini 3 / Imagen_
