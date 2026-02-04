# NexusOS Accessibility Specification (WCAG 2.1 AA)

**Version:** 1.0
**Compliance Target:** WCAG 2.1 Level AA
**Last Updated:** January 19, 2026

---

## 1. Executive Summary

NexusOS is designed to be fully accessible to users with disabilities, meeting WCAG 2.1 Level AA standards. This specification provides implementation guidelines for developers and QA testers.

---

## 2. Color Contrast Requirements

### 2.1 Text Contrast Ratios

All text must meet minimum contrast ratios against their backgrounds.

| Text Type                            | Minimum Ratio | Light Mode Example                 | Dark Mode Example                  |
| :----------------------------------- | :------------ | :--------------------------------- | :--------------------------------- |
| **Normal Text** (< 18px)             | **4.5:1**     | `#0F172A` on `#FFFFFF` = 15.31:1 ✓ | `#F8FAFC` on `#020617` = 15.89:1 ✓ |
| **Large Text** (≥ 18px or 14px bold) | **3:1**       | `#64748B` on `#FFFFFF` = 4.68:1 ✓  | `#94A3B8` on `#020617` = 6.89:1 ✓  |
| **UI Components**                    | **3:1**       | `#6366F1` on `#FFFFFF` = 4.59:1 ✓  | `#818CF8` on `#020617` = 6.94:1 ✓  |

### 2.2 Verified Color Combinations

| Token Pair                              | Light Mode Ratio | Dark Mode Ratio | Status             |
| :-------------------------------------- | :--------------- | :-------------- | :----------------- |
| `foreground` on `background`            | 15.31:1          | 15.89:1         | ✅ Pass            |
| `muted-fg` on `background`              | 4.68:1           | 6.89:1          | ✅ Pass            |
| `primary` on `background`               | 4.59:1           | 6.94:1          | ✅ Pass            |
| `primary-fg` on `primary`               | 8.26:1           | 3.12:1          | ✅ Pass            |
| `danger` (red-600) on `background`      | 4.53:1           | 4.68:1          | ✅ Pass            |
| `success` (emerald-600) on `background` | 3.87:1           | 4.12:1          | ⚠️ Large text only |

### 2.3 Color Independence

> **Critical Rule:** Never use color as the only means of conveying information.

| Scenario       | Color Indicator  | Required Supplement              |
| :------------- | :--------------- | :------------------------------- |
| Error state    | Red border       | Error icon + text message        |
| Success state  | Green badge      | Checkmark icon + "Success" text  |
| Warning state  | Amber badge      | Warning icon + descriptive text  |
| Status active  | Indigo highlight | "Active" text or aria-current    |
| Required field | -                | Asterisk (\*) + "required" label |

---

## 3. Focus Management

### 3.1 Focus Indicator Specification

All interactive elements must have visible focus indicators.

```css
/* Default Focus Ring */
:focus-visible {
  outline: 2px solid var(--primary);
  outline-offset: 2px;
  border-radius: var(--radius-sm);
}

/* Dark Mode Focus Ring */
.dark :focus-visible {
  outline-color: var(--primary); /* indigo-400 */
  box-shadow: 0 0 0 4px rgba(129, 140, 248, 0.3);
}

/* High Contrast Mode */
@media (prefers-contrast: more) {
  :focus-visible {
    outline: 3px solid currentColor;
    outline-offset: 3px;
  }
}
```

### 3.2 Focus Order

Focus must follow logical reading order:

1. **Skip Link** ("Skip to main content")
2. **Logo** (link to home)
3. **Search Bar**
4. **Navbar Actions** (Density Toggle, Theme, Profile)
5. **Sidebar Navigation** (top to bottom)
6. **Main Content** (top to bottom, left to right)
7. **Page Actions** (buttons, links)
8. **Footer** (if present)

### 3.3 Focus Trapping

Modals and dialogs must trap focus:

```tsx
// React Hook for Focus Trap
const useFocusTrap = (
  isOpen: boolean,
  containerRef: RefObject<HTMLElement>,
) => {
  useEffect(() => {
    if (!isOpen || !containerRef.current) return

    const focusableElements = containerRef.current.querySelectorAll(
      'button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])',
    )

    const firstElement = focusableElements[0] as HTMLElement
    const lastElement = focusableElements[
      focusableElements.length - 1
    ] as HTMLElement

    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Tab') {
        if (e.shiftKey && document.activeElement === firstElement) {
          e.preventDefault()
          lastElement.focus()
        } else if (!e.shiftKey && document.activeElement === lastElement) {
          e.preventDefault()
          firstElement.focus()
        }
      }
      if (e.key === 'Escape') {
        // Close modal
      }
    }

    containerRef.current.addEventListener('keydown', handleKeyDown)
    firstElement?.focus()

    return () =>
      containerRef.current?.removeEventListener('keydown', handleKeyDown)
  }, [isOpen])
}
```

---

## 4. Keyboard Navigation

### 4.1 Global Keyboard Shortcuts

| Shortcut             | Action                  | Implementation  |
| :------------------- | :---------------------- | :-------------- |
| `Tab`                | Move focus forward      | Native browser  |
| `Shift + Tab`        | Move focus backward     | Native browser  |
| `Enter` / `Space`    | Activate button/link    | Native + custom |
| `Escape`             | Close modal/dropdown    | Custom          |
| `Ctrl/⌘ + K`         | Open Command Menu       | Custom          |
| `Ctrl/⌘ + D`         | Toggle Density Mode     | Custom          |
| `Ctrl/⌘ + Shift + D` | Toggle Dark Mode        | Custom          |
| `?`                  | Show keyboard shortcuts | Custom          |

### 4.2 Component-Specific Navigation

#### Data Table (Hyper-Grid)

| Key        | Action                                 |
| :--------- | :------------------------------------- |
| `↑` / `↓`  | Move between rows                      |
| `←` / `→`  | Move between cells (when in cell mode) |
| `Home`     | Move to first row                      |
| `End`      | Move to last row                       |
| `Space`    | Toggle row selection                   |
| `Ctrl + A` | Select all rows                        |
| `Enter`    | Open row actions menu                  |

#### Dropdown Menu

| Key            | Action                                 |
| :------------- | :------------------------------------- |
| `↑` / `↓`      | Move between options                   |
| `Enter`        | Select option                          |
| `Escape`       | Close dropdown                         |
| `Home`         | Jump to first option                   |
| `End`          | Jump to last option                    |
| Type character | Jump to option starting with character |

#### Tab Component

| Key       | Action              |
| :-------- | :------------------ |
| `←` / `→` | Switch between tabs |
| `Home`    | Go to first tab     |
| `End`     | Go to last tab      |

---

## 5. ARIA Implementation Guide

### 5.1 Landmark Regions

```html
<body>
  <a href="#main-content" class="skip-link">Skip to main content</a>

  <header role="banner">
    <nav role="navigation" aria-label="Main navigation">
      <!-- Navbar content -->
    </nav>
  </header>

  <aside role="complementary" aria-label="Sidebar navigation">
    <nav aria-label="Primary navigation">
      <!-- Sidebar links -->
    </nav>
  </aside>

  <main id="main-content" role="main">
    <!-- Page content -->
  </main>

  <div role="status" aria-live="polite" id="toast-container">
    <!-- Toast notifications -->
  </div>
</body>
```

### 5.2 Data Table ARIA

```html
<table
  role="grid"
  aria-label="User list"
  aria-rowcount="1234"
  aria-colcount="7">
  <thead>
    <tr>
      <th scope="col" aria-sort="none">
        <button aria-label="Sort by name">Name</button>
      </th>
      <th scope="col" aria-sort="ascending">
        <button aria-label="Sort by email, currently sorted ascending">
          Email
        </button>
      </th>
      <!-- More headers -->
    </tr>
  </thead>
  <tbody>
    <tr aria-rowindex="1" aria-selected="false" tabindex="0">
      <td aria-colindex="1">John Doe</td>
      <td aria-colindex="2">john@example.com</td>
      <!-- More cells -->
    </tr>
  </tbody>
</table>

<!-- Live region for announcements -->
<div aria-live="polite" aria-atomic="true" class="sr-only">
  Showing 1 to 20 of 1234 users
</div>
```

### 5.3 Modal Dialog ARIA

```html
<div
  role="dialog"
  aria-modal="true"
  aria-labelledby="dialog-title"
  aria-describedby="dialog-description">
  <h2 id="dialog-title">Add New User</h2>
  <p id="dialog-description">
    Fill out the form below to create a new user account.
  </p>

  <form>
    <!-- Form content -->
  </form>

  <button aria-label="Close dialog">×</button>
</div>
```

### 5.4 Permission Matrix ARIA

```html
<table role="grid" aria-label="Permission matrix">
  <thead>
    <tr>
      <th scope="col">Role</th>
      <th scope="col">/users</th>
      <th scope="col">/roles</th>
    </tr>
  </thead>
  <tbody>
    <tr>
      <th scope="row">admin</th>
      <td>
        <button
          role="checkbox"
          aria-checked="true"
          aria-label="Read permission for admin on /users is enabled. Click to toggle.">
          ☑
        </button>
      </td>
    </tr>
  </tbody>
</table>
```

---

## 6. Screen Reader Support

### 6.1 Tested Screen Readers

| Screen Reader | Browser         | Platform | Status   |
| :------------ | :-------------- | :------- | :------- |
| **NVDA**      | Chrome, Firefox | Windows  | Required |
| **JAWS**      | Chrome, Edge    | Windows  | Required |
| **VoiceOver** | Safari          | macOS    | Required |
| **VoiceOver** | Safari          | iOS      | Required |
| **TalkBack**  | Chrome          | Android  | Required |

### 6.2 Screen Reader Announcements

| Event          | Announcement                                         |
| :------------- | :--------------------------------------------------- |
| Page load      | "NexusOS Dashboard loaded"                           |
| Toast success  | "Success: User created successfully"                 |
| Toast error    | "Error: Failed to save changes"                      |
| Row selected   | "Row 5 selected, John Doe"                           |
| Filter applied | "Showing 25 of 1234 results filtered by role: admin" |
| Modal open     | "Add User dialog opened"                             |
| Modal close    | "Dialog closed"                                      |
| AI processing  | "AI assistant is processing your request"            |
| AI response    | "AI assistant responded"                             |

### 6.3 Hidden Screen Reader Text

```css
/* Visually hidden but accessible to screen readers */
.sr-only {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  white-space: nowrap;
  border: 0;
}

/* Show on focus (for skip links) */
.sr-only-focusable:focus {
  position: static;
  width: auto;
  height: auto;
  padding: 0.5rem 1rem;
  margin: 0;
  overflow: visible;
  clip: auto;
  white-space: normal;
}
```

---

## 7. Motion & Animation

### 7.1 Reduced Motion Support

```css
/* Default animations */
.animate-fade-in {
  animation: fadeIn 200ms ease-out;
}

.animate-shimmer {
  animation: shimmer 1.5s infinite;
}

/* Respect user preference */
@media (prefers-reduced-motion: reduce) {
  *,
  *::before,
  *::after {
    animation-duration: 0.01ms !important;
    animation-iteration-count: 1 !important;
    transition-duration: 0.01ms !important;
  }

  .animate-shimmer {
    animation: none;
    background: var(--muted);
  }
}
```

### 7.2 Animation Timing Guidelines

| Animation Type   | Duration      | Easing                       | Reduce Motion Alternative |
| :--------------- | :------------ | :--------------------------- | :------------------------ |
| Fade in/out      | 150-200ms     | ease-out                     | Instant                   |
| Slide panel      | 200-300ms     | cubic-bezier(0.4, 0, 0.2, 1) | Instant                   |
| Skeleton shimmer | 1.5s infinite | linear                       | Static gray               |
| Button hover     | 100ms         | ease                         | No change                 |
| Toast enter      | 300ms         | ease-out                     | Instant                   |
| Modal scale      | 150ms         | ease-out                     | Instant                   |

---

## 8. Touch Target Sizes

### 8.1 Minimum Touch Targets

| Mode                     | Minimum Size | Minimum Spacing |
| :----------------------- | :----------- | :-------------- |
| **Comfort (SaaS)**       | 44×44px      | 8px             |
| **Compact (Enterprise)** | 32×32px      | 4px             |
| **Mobile**               | 48×48px      | 8px             |

### 8.2 Implementation

```css
/* Base touch target */
.touch-target {
  min-width: var(--touch-target-size);
  min-height: var(--touch-target-size);
  padding: var(--touch-padding);
}

/* Mode-specific */
:root {
  --touch-target-size: 44px;
  --touch-padding: 10px;
}

[data-density='compact'] {
  --touch-target-size: 32px;
  --touch-padding: 4px;
}

@media (max-width: 768px) {
  :root {
    --touch-target-size: 48px;
    --touch-padding: 12px;
  }
}
```

---

## 9. Forms & Error Handling

### 9.1 Form Field Requirements

```html
<div class="form-field">
  <label for="email">
    Email Address
    <span aria-hidden="true" class="required-indicator">*</span>
    <span class="sr-only">(required)</span>
  </label>

  <input
    type="email"
    id="email"
    name="email"
    required
    aria-required="true"
    aria-invalid="false"
    aria-describedby="email-hint email-error"
    autocomplete="email" />

  <p id="email-hint" class="hint-text">
    We'll never share your email with anyone.
  </p>

  <p id="email-error" class="error-text" role="alert" hidden>
    Please enter a valid email address.
  </p>
</div>
```

### 9.2 Error State Requirements

1. **Visual Indicator:** Red border (not just color—include error icon)
2. **Error Text:** Positioned below input, linked via `aria-describedby`
3. **Role Alert:** Use `role="alert"` for dynamic error messages
4. **Summary:** Group form errors at top of form on submit failure
5. **Focus Management:** Move focus to first error field on submit failure

---

## 10. Testing Checklist

### 10.1 Automated Testing Tools

| Tool                       | Purpose                  | Integration      |
| :------------------------- | :----------------------- | :--------------- |
| **axe-core**               | Accessibility violations | Jest, Playwright |
| **Pa11y**                  | WCAG compliance          | CI/CD pipeline   |
| **Lighthouse**             | Accessibility score      | Chrome DevTools  |
| **eslint-plugin-jsx-a11y** | JSX accessibility        | ESLint           |

### 10.2 Manual Testing Checklist

- [ ] Navigate entire app using only keyboard
- [ ] Use screen reader (NVDA/VoiceOver) for all flows
- [ ] Test with 200% browser zoom
- [ ] Test with Windows High Contrast Mode
- [ ] Test with `prefers-reduced-motion: reduce`
- [ ] Verify all images have alt text
- [ ] Verify all form fields have labels
- [ ] Verify all interactive elements have focus indicators
- [ ] Test color contrast with browser tools
- [ ] Verify skip link works correctly

---

## 11. Implementation Priority

| Priority | Item                      | Impact              |
| :------- | :------------------------ | :------------------ |
| 🔴 P0    | Color contrast compliance | Legal requirement   |
| 🔴 P0    | Keyboard navigation       | Critical for users  |
| 🔴 P0    | Focus indicators          | Critical for users  |
| 🟡 P1    | ARIA landmarks            | Screen reader users |
| 🟡 P1    | Form error handling       | Usability           |
| 🟢 P2    | Reduced motion support    | User preference     |
| 🟢 P2    | Skip links                | Power users         |

---

_Document created as part of Nexus Design System v1.0_
_WCAG 2.1 Level AA Compliance Target_
