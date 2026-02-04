# NexusOS Responsive Design Specification

**Version:** 1.0
**Last Updated:** January 19, 2026
**Purpose:** Define responsive behavior across all breakpoints for the NexusOS dashboard

---

## 1. Breakpoint System

### 1.1 Primary Breakpoints

| Name            | Width         | Device Context            | Density Default |
| :-------------- | :------------ | :------------------------ | :-------------- |
| **Mobile**      | 0 - 767px     | Smartphones (portrait)    | Comfort         |
| **Tablet**      | 768 - 1023px  | Tablets, large phones     | Comfort         |
| **Desktop**     | 1024 - 1439px | Laptops, small monitors   | User preference |
| **Desktop XL**  | 1440 - 1919px | Desktop monitors          | User preference |
| **Desktop XXL** | ≥1920px       | Large/ultra-wide monitors | User preference |

### 1.2 CSS Implementation

```css
/* Mobile First Approach */
:root {
  /* Base mobile styles */
}

/* Tablet */
@media (min-width: 768px) {
  /* Tablet styles */
}

/* Desktop */
@media (min-width: 1024px) {
  /* Desktop styles */
}

/* Desktop XL */
@media (min-width: 1440px) {
  /* Large desktop styles */
}

/* Desktop XXL */
@media (min-width: 1920px) {
  /* Ultra-wide styles */
}
```

### 1.3 Tailwind v4 Configuration

```css
@theme {
  --breakpoint-sm: 640px;
  --breakpoint-md: 768px;
  --breakpoint-lg: 1024px;
  --breakpoint-xl: 1440px;
  --breakpoint-2xl: 1920px;
}
```

---

## 2. Layout Transformations

### 2.1 Main Layout Grid

```
┌─ Desktop (≥1024px) ─────────────────────────────────────────┐
│ ┌─────────────────────────────────────────────────────────┐ │
│ │                      NAVBAR (full)                      │ │
│ ├─────────┬───────────────────────────────────────────────┤ │
│ │         │                                               │ │
│ │ SIDEBAR │              MAIN CONTENT                     │ │
│ │  (280px │               (flex-1)                        │ │
│ │   or    │                                               │ │
│ │  72px)  │                                               │ │
│ │         │                                               │ │
│ └─────────┴───────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘

┌─ Tablet (768px - 1023px) ───────────────────────────────────┐
│ ┌─────────────────────────────────────────────────────────┐ │
│ │  ☰ NAVBAR (hamburger menu)                              │ │
│ ├─────────────────────────────────────────────────────────┤ │
│ │                                                         │ │
│ │                    MAIN CONTENT                         │ │
│ │                     (full width)                        │ │
│ │                                                         │ │
│ └─────────────────────────────────────────────────────────┘ │
│ ┌─ SIDEBAR DRAWER ─────────────────────────────────────────┐ │
│ │ (slides in from left when hamburger clicked)            │ │
│ └─────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘

┌─ Mobile (<768px) ───────────────────────────────────────────┐
│ ┌─────────────────────────────────────────────────────────┐ │
│ │  ☰ NAVBAR (simplified)                                  │ │
│ ├─────────────────────────────────────────────────────────┤ │
│ │                                                         │ │
│ │                    MAIN CONTENT                         │ │
│ │                  (full width, stacked)                  │ │
│ │                                                         │ │
│ ├─────────────────────────────────────────────────────────┤ │
│ │  🏠  👥  🔐  📋  ⚙️   BOTTOM NAV                        │ │
│ └─────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

### 2.2 Layout CSS Grid Definition

```css
/* Desktop Layout */
.layout-container {
  display: grid;
  grid-template-areas:
    'navbar navbar'
    'sidebar main';
  grid-template-columns: var(--sidebar-width) 1fr;
  grid-template-rows: var(--navbar-height) 1fr;
  height: 100vh;
}

/* Tablet Layout */
@media (max-width: 1023px) {
  .layout-container {
    grid-template-areas:
      'navbar'
      'main';
    grid-template-columns: 1fr;
    grid-template-rows: var(--navbar-height) 1fr;
  }

  .sidebar {
    position: fixed;
    left: -280px;
    top: 0;
    height: 100vh;
    transition: left 300ms ease;
    z-index: 40;
  }

  .sidebar.open {
    left: 0;
  }

  .sidebar-overlay {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.5);
    z-index: 39;
  }
}

/* Mobile Layout */
@media (max-width: 767px) {
  .layout-container {
    grid-template-areas:
      'navbar'
      'main'
      'bottomnav';
    grid-template-rows: 56px 1fr 64px;
  }

  .bottom-nav {
    display: flex;
    justify-content: space-around;
    align-items: center;
    background: var(--surface);
    border-top: 1px solid var(--border);
  }
}
```

---

## 3. Component Responsive Behavior

### 3.1 Navbar

| Breakpoint | Height                          | Logo        | Search             | Actions        | Profile       |
| :--------- | :------------------------------ | :---------- | :----------------- | :------------- | :------------ |
| Mobile     | 56px                            | Icon only   | Hidden (in drawer) | Theme only     | Avatar only   |
| Tablet     | 64px                            | Icon + Text | Icon trigger       | Theme, Density | Avatar only   |
| Desktop    | 80px (Comfort) / 56px (Compact) | Full logo   | Full bar           | All actions    | Avatar + Name |

**Mobile Navbar:**

```
┌──────────────────────────────────────┐
│  ☰  │  NexusOS  │  🔍  │  🌙  │  👤  │
└──────────────────────────────────────┘
```

### 3.2 Sidebar

| Breakpoint        | Display           | Width | Behavior             |
| :---------------- | :---------------- | :---- | :------------------- |
| Mobile            | Bottom nav        | 100%  | 5 icon tabs          |
| Tablet            | Off-canvas drawer | 280px | Hamburger trigger    |
| Desktop (Comfort) | Visible           | 280px | Expanded with labels |
| Desktop (Compact) | Visible           | 72px  | Rail (icons only)    |

### 3.3 Data Grid (Hyper-Grid)

| Breakpoint | Display Mode            | Columns              | Pagination      |
| :--------- | :---------------------- | :------------------- | :-------------- |
| Mobile     | **Card Stack**          | Single column cards  | Infinite scroll |
| Tablet     | Horizontal scroll table | 4-5 priority columns | Numbered pages  |
| Desktop    | Full table              | All columns          | Numbered pages  |

**Mobile Card Layout:**

```
┌────────────────────────────────┐
│  ○ John Doe                    │
│  john@example.com              │
│  ┌─────────┐  ┌───────────┐   │
│  │ Admin   │  │ ● Active  │   │
│  └─────────┘  └───────────┘   │
│  Created: Jan 15, 2026         │
│  ─────────────────────────     │
│  [Edit]  [Assign Role]  [•••] │
└────────────────────────────────┘
```

### 3.4 Permission Matrix

| Breakpoint | Display Mode                                                       |
| :--------- | :----------------------------------------------------------------- |
| Mobile     | **Accordion List** - Role expands to show permissions              |
| Tablet     | **Swipeable Columns** - Horizontal scroll with sticky first column |
| Desktop    | Full matrix grid                                                   |

**Mobile Accordion:**

```
┌────────────────────────────────┐
│  ▸ admin                       │
├────────────────────────────────┤
│  ▾ editor                      │
│    /users   ☑R ☐C ☐U ☐D       │
│    /roles   ☑R ☐C ☐U ☐D       │
│    /content ☑R ☑C ☑U ☐D       │
├────────────────────────────────┤
│  ▸ viewer                      │
└────────────────────────────────┘
```

### 3.5 KPI Cards

| Breakpoint | Grid        | Card Size                              |
| :--------- | :---------- | :------------------------------------- |
| Mobile     | 2 per row   | Compact                                |
| Tablet     | 2-3 per row | Medium                                 |
| Desktop    | 4 per row   | Full (Comfort) or Compact (Enterprise) |

```css
.kpi-grid {
  display: grid;
  gap: var(--component-gap);
  grid-template-columns: repeat(2, 1fr); /* Mobile */
}

@media (min-width: 768px) {
  .kpi-grid {
    grid-template-columns: repeat(3, 1fr);
  }
}

@media (min-width: 1024px) {
  .kpi-grid {
    grid-template-columns: repeat(4, 1fr);
  }
}
```

### 3.6 Modals & Dialogs

| Breakpoint | Display Mode                                  | Width                  |
| :--------- | :-------------------------------------------- | :--------------------- |
| Mobile     | **Full-screen sheet** (slides up from bottom) | 100%                   |
| Tablet     | Centered modal                                | 80% or max-width 560px |
| Desktop    | Centered modal                                | max-width 640px        |

```css
.modal-dialog {
  width: 100%;
  max-width: 640px;
  margin: auto;
}

@media (max-width: 767px) {
  .modal-dialog {
    position: fixed;
    bottom: 0;
    left: 0;
    right: 0;
    max-width: 100%;
    max-height: 90vh;
    border-radius: 16px 16px 0 0;
    animation: slideUp 300ms ease;
  }

  @keyframes slideUp {
    from {
      transform: translateY(100%);
    }
    to {
      transform: translateY(0);
    }
  }
}
```

### 3.7 AI Chat Panel

| Breakpoint       | Display Mode                                   |
| :--------------- | :--------------------------------------------- |
| Mobile           | Full-screen overlay                            |
| Tablet           | Full-screen or slide-in from right (50% width) |
| Desktop (Float)  | Fixed bottom-right widget (380x600px)          |
| Desktop (Docked) | Side panel (30% screen width)                  |

### 3.8 Forms

| Breakpoint        | Label Position        | Layout               |
| :---------------- | :-------------------- | :------------------- |
| Mobile            | Stacked (above input) | Single column        |
| Tablet            | Stacked               | Single or two column |
| Desktop (Comfort) | Stacked               | Multi-column         |
| Desktop (Compact) | Inline (side-by-side) | Dense multi-column   |

---

## 4. Touch Adaptations

### 4.1 Touch Target Sizes

| Breakpoint | Minimum Size                          | Spacing |
| :--------- | :------------------------------------ | :------ |
| Mobile     | 48×48px                               | 8px     |
| Tablet     | 44×44px                               | 8px     |
| Desktop    | 32×32px (Compact) / 44×44px (Comfort) | 4-8px   |

### 4.2 Swipe Gestures (Mobile/Tablet)

| Gesture               | Action                              |
| :-------------------- | :---------------------------------- |
| Swipe left on row     | Reveal quick actions (Edit, Delete) |
| Swipe right on row    | Select row                          |
| Pull down             | Refresh data                        |
| Swipe right from edge | Open sidebar drawer                 |

### 4.3 Long Press Actions

| Element   | Long Press Action     |
| :-------- | :-------------------- |
| Table row | Enter selection mode  |
| Card      | Show context menu     |
| Avatar    | Quick preview profile |

---

## 5. Navigation Transformations

### 5.1 Bottom Navigation (Mobile Only)

```
┌─────────────────────────────────────────────────┐
│   🏠      │   👥      │   🔐      │   📋      │ ⚙️
│  Home     │  Users    │  Access   │  Audit    │ More
└─────────────────────────────────────────────────┘
```

**Specification:**

- Fixed at bottom
- Height: 64px
- 5 items maximum
- Active state: Filled icon + label
- Inactive state: Outlined icon + muted label

### 5.2 Hamburger Menu (Tablet)

When hamburger is clicked:

1. Overlay appears (opacity: 0.5)
2. Sidebar slides in from left (280px)
3. Focus trapped in sidebar
4. Escape or overlay click closes

### 5.3 Quick Actions

| Breakpoint | Implementation                               |
| :--------- | :------------------------------------------- |
| Mobile     | FAB (Floating Action Button) at bottom-right |
| Tablet     | FAB or contextual action bar                 |
| Desktop    | Toolbar buttons                              |

---

## 6. Content Prioritization

### 6.1 Dashboard Zones by Breakpoint

| Zone          | Mobile                            | Tablet      | Desktop     |
| :------------ | :-------------------------------- | :---------- | :---------- |
| KPI Cards     | 2 cards top, rest in "More Stats" | All visible | All visible |
| Recent Logs   | 3 items, "View All" link          | 5 items     | 10 items    |
| Quick Actions | FAB expand                        | Action bar  | Toolbar     |
| Charts        | Swipeable carousel                | 2 per row   | Full grid   |

### 6.2 Table Column Priority

Define which columns show at each breakpoint:

| Column        | Mobile              | Tablet | Desktop |
| :------------ | :------------------ | :----- | :------ |
| Checkbox      | ❌                  | ✅     | ✅      |
| Avatar + Name | ✅ (Primary)        | ✅     | ✅      |
| Email         | ❌ (In card detail) | ✅     | ✅      |
| Role          | ✅ (Badge)          | ✅     | ✅      |
| Status        | ✅ (Badge)          | ✅     | ✅      |
| Created Date  | ❌                  | ❌     | ✅      |
| Last Login    | ❌                  | ❌     | ✅      |
| Actions       | ✅ (Swipe/Tap)      | ✅     | ✅      |

---

## 7. Performance Considerations

### 7.1 Mobile-Specific Optimizations

| Optimization         | Implementation                     |
| :------------------- | :--------------------------------- |
| **Image Loading**    | Load lower resolution on mobile    |
| **Infinite Scroll**  | Virtual scrolling for large lists  |
| **Lazy Load**        | Load off-screen content on demand  |
| **Skeleton Loading** | Show skeletons instead of spinners |
| **Touch Debounce**   | 100ms debounce on rapid taps       |

### 7.2 Network-Aware Loading

```javascript
// Detect slow connections
const connection = navigator.connection || navigator.mozConnection

if (connection.effectiveType === '2g' || connection.saveData) {
  // Load minimal data
  // Disable animations
  // Use lower quality images
}
```

---

## 8. Testing Requirements

### 8.1 Device Testing Matrix

| Device             | Screen Size | Priority       |
| :----------------- | :---------- | :------------- |
| iPhone 13/14/15    | 390×844     | 🔴 Required    |
| iPhone SE          | 375×667     | 🟡 Recommended |
| Samsung Galaxy S21 | 360×800     | 🔴 Required    |
| iPad 10th Gen      | 820×1180    | 🔴 Required    |
| iPad Pro 12.9"     | 1024×1366   | 🟡 Recommended |
| MacBook 13"        | 1440×900    | 🔴 Required    |
| Desktop 1080p      | 1920×1080   | 🔴 Required    |
| Desktop 4K         | 3840×2160   | 🟢 Optional    |

### 8.2 Orientation Testing

All tablet and mobile views must work in:

- Portrait orientation
- Landscape orientation
- Dynamic orientation changes

---

_Responsive design specification for Nexus Design System v1.0_
_Mobile-first, progressively enhanced for all devices_
