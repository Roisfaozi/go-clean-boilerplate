# NexusOS Structured AI Prompts (v2)

**Based on Framework:** [Designed for Humans: Gemini 3 Speed Up](https://designedforhumans.tech/blog/can-gemini-3-speed-up-ui-design-without-losing-quality)
**Optimization:** Gemini 3 / High-Reasoning Models
**Core Structure:** Role + Context + Task + Constraints + Output

---

## 1️⃣ WIREFRAMING (Prompt-to-Prototype Loop)

Use these for generating initial screens, flows, and layout ideas before high-fidelity design.

### A. New Feature Flow

```markdown
**Role:** Senior Product Designer for Enterprise SaaS
**Context:** NexusOS Admin Dashboard (Hybrid Comfort/Compact density).
**Task:** Create a wireframe flow for [Feature Name, e.g., "Bulk User Import"].
**Flow Steps:** [Step 1] → [Step 2] → [Step 3] -> [Step 4]
**Constraints:**

- Use "Fluid Density" layout principles (adaptable spacing).
- Must use standard NexusOS components: HyperGrid, Slide-over Panel, Toast.
- Accessibility: WCAG 2.1 AA compliant.
- Navigation: Sidebar (left), Header (top).
  **Ask:** Create 3 alternative wireframes for this flow with annotations.
  **Output:**

1. List of required screens.
2. Key components per screen.
3. Edge cases (e.g., error states, empty states).
4. Simple textual wireframe (ASCII or structure outline).
```

### B. Single Screen Layout

```markdown
**Role:** UI/UX Designer
**Context:** NexusOS [Module Name] - [Screen Name]
**Task:** Design the layout structure for the [Screen Name] screen.
**Content Zones:**

1. Header: Page title, breadcrumbs, primary action.
2. Main Content: [Describe main data/widget].
3. Sidebar/Panel: [Describe secondary info if any].
   **Constraints:**

- Density Mode: [Comfort (SaaS) / Compact (Enterprise)].
- Colors: Grayscale only (focus on layout).
- Spacing: Use NexusOS 4pt scale (16px, 24px, 32px gaps).
  **Output:**
- Detailed layout description using Tailwind grid/flex terminology.
- List of "Molecules" and "Organisms" needed from the design system.
- Annotation of user interactive zones.
```

---

## 2️⃣ DESIGN SYSTEM FOUNDATIONS (The Consistency Stack)

Use these to expand or verify the design system tokens and variables.

### A. Token Generation/Extension

```markdown
**Role:** Design System Architect
**Context:** NexusOS "Nebula" Design System.
**Task:** Propose a semantic token set for a new [Feature/Theme, e.g., "Data Visualization Palette"].
**Current Foundations:**

- Primary: Indigo (#6366F1)
- Secondary: Teal (#14B8A6)
- Neutrals: Slate (#0F172A - #F8FAFC)
  **Constraints:**
- Must align with existing "Fluid Density" system.
- Provide Light and Dark mode values.
- Check WCAG AA contrast ratios against Surface (#F8FAFC / #0F172A).
  **Output:**
- JSON format matching `design-tokens.json` schema.
- Brief rationale for color choices.
- Usage table (Token Name | Value | Usage Rule).
```

### B. Component Specification

```markdown
**Role:** Design System Lead
**Context:** NexusOS Component Library
**Task:** Create a detailed specification for the [Component Name] component.
**Requirements:**

- Variants: [List variants, e.g., Solid, Outline, Ghost].
- States: Default, Hover, Focused, Disabled, Loading.
- Sizes: Comfort (Standard) and Compact (Dense).
  **Constraints:**
- Use NexusOS Spacing variables (e.g., $spacing-4).
- Use Semantic Colors (e.g., $primary-500).
- Interaction: Define hover/focus transitions (200ms cubic-bezier).
  **Output:**
- Component Anatomy list (Icon, Label, Container).
- Design Tokens mapping table.
- Accessibility requirements (ARIA roles, keyboard nav).
```

---

## 3️⃣ UI TO CODE (Handover)

Use these to convert designs/specs into actual NexusOS code (React/Tailwind/Shadcn).

### A. Component Implementation

```markdown
**Role:** Senior Frontend Engineer (React/Next.js)
**Context:** NexusOS codebase (Next.js 16, Tailwind v4, Shadcn UI).
**Task:** Implement the [Component Name] based on this spec.
**Input Spec:**

- [Paste Component Spec or Description here]
  **Constraints:**
- Use `tailwind-merge` and `clsx` for class management.
- Support "Comfort" and "Compact" density via context or props.
- Use `lucide-react` for icons.
- Ensure Dark Mode support (use `dark:` variants).
- Typescript: Strict typing with Interfaces.
  **Output:**
- Single `.tsx` file code.
- Associated `interface` definitions.
- Usage example.
```

### B. Page Implementation

```markdown
**Role:** Frontend Developer
**Context:** NexusOS Admin Dashboard
**Task:** Build the [Page Name] using existing components.
**Layout Structure:**

- DashboardShell (Sidebar + Header)
- PageHeader (Title + Actions)
- PageContent (Grid layout)
  **Constraints:**
- Responsive: Mobile (Stack) -> Tablet (2 Col) -> Desktop (4 Col).
- Use `Grid` and `Flex` utilities.
- Implement "Loading" states using Skeleton components.
- Data fetching simulation (no real API calls).
  **Output:**
- Complete Page Component code.
```

---

## 4️⃣ VISUAL ASSETS (Illustration & Photo)

Use these to generate consistent imagery using Gemini/Imagen.

### A. Brand Illustration

```markdown
**Role:** Brand Illustrator
**Audience:** B2B Enterprise Users & SaaS Founders
**Task:** Create a feature illustration for [Feature Name, e.g., "AI Analytics"].
**Style Brief:**

- Style: Tech-Minimalist, Flat with subtle depth.
- Palette: Indigo (#6366F1), Teal (#14B8A6), Violet (#8B5CF6).
- Mood: Professional, Trustworthy, Innovative.
- Forms: Rounded corners, geometric shapes, thin line accents.
- Avoid: Cartoons, messy textures, 3D render style.
  **Output:**
- 3 Scene prompts describing the composition.
- Hex color palette usage.
- "Do and Don't" rules for this specific illustration.
```

---

## 5️⃣ ACCESSIBILITY AUDIT (Quality Check)

Use these to verify designs or code before shipping.

### A. Quick Spec Audit

```markdown
**Role:** Accessibility Specialist (WCAG 2.1/2.2 Expert)
**Context:** NexusOS Interface Design
**Task:** Audit this [Component/Screen] description/code for accessibility issues.
**Checklist:**

1. Color Contrast (Text vs Background) - Verify in both Light & Dark modes.
2. Keyboard Navigation & Focus States.
3. Screen Reader Labels (aria-label, alt text).
4. Touch Target Sizes (min 44px for Comfort mode).
5. Information Density & Readability.
   **Input:** [Paste Design Description or Code Snippet]
   **Output:**

- Table of Issues (Severity: High/Med/Low).
- Relevant WCAG Guideline references.
- Suggested fixes (Code or Design tweaks).
- Brief Compliance Summary (%).
```

---

## 💡 BEST PRACTICES FOR USING THESE PROMPTS

1.  **Define the "Job to be Done":** Don't just ask for a UI. Ask for a solution to a user problem.
2.  **Iterate:** Use the "Output" from one prompt (e.g., Wireframe) as the "Input" for the next (e.g., Component Spec).
3.  **Enforce Tokens:** Always reference specific NexusOS colors (Indigo/Teal) and spacing to prevent "hallucinated" styles.
4.  **Human Review:** Always audit the output using Prompt 5 (Accessibility) before finalizing.
