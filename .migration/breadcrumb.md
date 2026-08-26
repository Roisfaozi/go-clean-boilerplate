# breadcrumb

2026-07-08, transformation engine + render-prop consumer sweep. Verdict: migrated wrapper shape, preserved public `asChild` compatibility, call sites moved to `render`.

## Changed

- `packages/ui/src/breadcrumb.tsx:1-120` — replaced Radix Slot with Base UI `useRender` + `mergeProps`. `BreadcrumbLink` now accepts `render` and still accepts `asChild` for compatibility.
- `apps/client/app/components/layout/app-breadcrumb.tsx:42-69` — updated both `BreadcrumbLink asChild` call sites to `render={<Link ... />}`.
- Leftover scan on component files clean: no `@radix-ui` imports remain in `packages/ui/src/breadcrumb.tsx`.

## Left alone

- `BreadcrumbPage`, `BreadcrumbSeparator`, `BreadcrumbEllipsis` — already native; no Radix dependency.
- `Breadcrumb` / `BreadcrumbList` / `BreadcrumbItem` — unchanged layout wrappers.

## Behavior changes

- Preferred composition now `render`, not `asChild`.
- `render` path is functionally same for app link cases; no visible UI delta expected.

## Verify by hand

- Open breadcrumb trail in client app.
- Click root and inner crumbs.
- Confirm last crumb stays plain text.
