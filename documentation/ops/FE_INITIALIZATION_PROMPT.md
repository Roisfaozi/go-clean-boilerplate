# TASK: Initialize NexusOS Frontend (Based on ChadNext)

You are an expert Frontend Architect. Your goal is to scaffold the "NexusOS" frontend using the **ChadNext** (moinulmoin/chadnext) architecture as your foundation.

## 1. Core Mission
Transform the lightweight ChadNext starter into "NexusOS" - a high-density Enterprise SaaS platform. You will retain the clean folder structure and auth logic but completely overhaul the styling and components to support the "Fluid Density" system.

## 2. Context & Assets (Source of Truth)
You MUST read and implement the specifications from these files:
*   `documentation/productplan/design/design-tokens-complete.json` (Nebula Palette)
*   `documentation/productplan/design/readme.md` (Fluid Density Logic)
*   `documentation/productplan/design/spekui.md` (Visual Philosophy)

## 3. Step-by-Step Execution Plan

### Phase 1: Scaffolding (The ChadNext Way)
1.  Study the `moinulmoin/chadnext` repository structure.
2.  Initialize a Next.js 15+ application following that exact pattern (app router, lucia/auth.js, shadcn).
3.  **Route Architecture:**
    *   `app/(marketing)/`: Landing page, pricing.
    *   `app/(dashboard)/`: Authenticated dashboard.
    *   `app/(auth)/`: Auth flows.

### Phase 2: Design System Injection (Critical)
1.  **Tailwind CSS v4 + Nebula Palette:**
    *   Inject the `design-tokens-complete.json` into `app/globals.css`.
    *   Implement the **Dual-Density Logic**: Define CSS Variables for Comfort (default) and Compact (`[data-density="compact"]`).
    *   **Mandatory Variables:** `--radius`, `--input-height`, `--sidebar-width`, `--font-size-body`.
2.  **Shadcn Overrides:** Modify `components/ui` to use these variables so components resize automatically when the density attribute changes.

### Phase 3: Enterprise Sidebar & Navigation
1.  Extend the basic dashboard layout to support a **Collapsible Sidebar**.
2.  When `density` is "compact", the sidebar must shrink to a 72px **Rail Mode** (icon-only with tooltips).
3.  Implement a global **Density Toggle** in the Navbar using Zustand.

### Phase 4: Data Layer Preparation
1.  Install `@tanstack/react-table` and `@tanstack/react-query`.
2.  Create a `HyperGrid` component base: A high-performance wrapper for TanStack Table that supports sticky headers and the specific "Nebula" zebra-striping for dark mode.

## 4. Deliverable
A production-ready Next.js application where:
1.  The architecture follows ChadNext's best practices.
2.  The UI instantly adapts between "Comfort" (friendly SaaS) and "Compact" (precise Enterprise) modes.
3.  The color palette is 100% compliant with the "Nebula" specifications.

**START NOW with Phase 1.**
