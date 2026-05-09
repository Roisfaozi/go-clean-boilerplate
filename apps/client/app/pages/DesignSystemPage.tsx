import { useState } from "react";
import { PageHeader } from "@/components/layout/page-header";
import {  NexusCard, NexusCardHeader, NexusCardTitle, NexusCardDescription, NexusCardContent  } from "@casbin/ui";
import {  Badge  } from "@casbin/ui";
import {  Tabs, TabsContent, TabsList, TabsTrigger  } from "@casbin/ui";
import {
  Palette, Type, Maximize, Layers, Square, Droplets, Moon, Sun,
} from "lucide-react";

const ColorSwatch = ({ name, variable, className }: { name: string; variable: string; className: string }) => (
  <div className="flex flex-col items-center gap-2 group">
    <div className={`w-16 h-16 rounded-lg border border-border shadow-sm transition-transform group-hover:scale-110 ${className}`} />
    <span className="text-caption font-medium text-foreground">{name}</span>
    <code className="text-[10px] font-mono text-muted-foreground bg-muted px-1.5 py-0.5 rounded">{variable}</code>
  </div>
);

const TokenRow = ({ label, value, preview }: { label: string; value: string; preview?: React.ReactNode }) => (
  <div className="flex items-center justify-between py-3 px-4 rounded-md hover:bg-muted/50 transition-colors">
    <div className="flex items-center gap-3">
      {preview}
      <span className="text-sm font-medium text-foreground">{label}</span>
    </div>
    <code className="text-xs font-mono text-muted-foreground bg-muted px-2 py-1 rounded">{value}</code>
  </div>
);

const DocSection = ({ title, description, children }: { title: string; description?: string; children: React.ReactNode }) => (
  <NexusCard>
    <NexusCardHeader>
      <NexusCardTitle>{title}</NexusCardTitle>
      {description && <NexusCardDescription>{description}</NexusCardDescription>}
    </NexusCardHeader>
    <NexusCardContent>{children}</NexusCardContent>
  </NexusCard>
);

export default function DesignSystemPage() {
  return (
    <div className="space-y-6 max-w-6xl">
      <PageHeader
        title="Design System"
        description="NexusOS design tokens, color palette, typography scale, spacing, and component guidelines."
      />

      <Tabs defaultValue="colors" className="space-y-6">
        <TabsList className="flex-wrap">
          <TabsTrigger value="colors"><Palette className="h-4 w-4 mr-2" />Colors</TabsTrigger>
          <TabsTrigger value="typography"><Type className="h-4 w-4 mr-2" />Typography</TabsTrigger>
          <TabsTrigger value="spacing"><Maximize className="h-4 w-4 mr-2" />Spacing</TabsTrigger>
          <TabsTrigger value="radii"><Square className="h-4 w-4 mr-2" />Radii & Shadows</TabsTrigger>
          <TabsTrigger value="guidelines"><Layers className="h-4 w-4 mr-2" />Guidelines</TabsTrigger>
        </TabsList>

        {/* ── Colors ── */}
        <TabsContent value="colors" className="space-y-6">
          <DocSection title="Core Palette" description="Foundation colors used across all components.">
            <div className="flex flex-wrap gap-5">
              <ColorSwatch name="Background" variable="--background" className="bg-background" />
              <ColorSwatch name="Foreground" variable="--foreground" className="bg-foreground" />
              <ColorSwatch name="Surface" variable="--surface" className="bg-surface" />
              <ColorSwatch name="Muted" variable="--muted" className="bg-muted" />
              <ColorSwatch name="Border" variable="--border" className="bg-border" />
              <ColorSwatch name="Card" variable="--card" className="bg-card" />
            </div>
          </DocSection>

          <DocSection title="Brand Colors" description="Primary interaction and accent colors.">
            <div className="flex flex-wrap gap-5">
              <ColorSwatch name="Primary" variable="--primary" className="bg-primary" />
              <ColorSwatch name="Secondary" variable="--secondary" className="bg-secondary" />
              <ColorSwatch name="Accent" variable="--accent" className="bg-accent" />
            </div>
          </DocSection>

          <DocSection title="Semantic Colors" description="Status and feedback indicators.">
            <div className="flex flex-wrap gap-5">
              <ColorSwatch name="Info" variable="--info" className="bg-info" />
              <ColorSwatch name="Success" variable="--success" className="bg-success" />
              <ColorSwatch name="Warning" variable="--warning" className="bg-warning" />
              <ColorSwatch name="Danger" variable="--danger" className="bg-danger" />
              <ColorSwatch name="Destructive" variable="--destructive" className="bg-destructive" />
            </div>
          </DocSection>

          <DocSection title="Usage Guidelines">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4 text-sm">
              <div className="bg-muted/50 rounded-lg p-4 space-y-2">
                <Badge variant="secondary">Do</Badge>
                <ul className="list-disc list-inside text-muted-foreground space-y-1">
                  <li>Use semantic tokens: <code className="text-xs bg-muted px-1 rounded">text-foreground</code></li>
                  <li>Reference CSS variables for consistency</li>
                  <li>Test both light and dark modes</li>
                </ul>
              </div>
              <div className="bg-destructive/5 rounded-lg p-4 space-y-2">
                <Badge variant="destructive">Don't</Badge>
                <ul className="list-disc list-inside text-muted-foreground space-y-1">
                  <li>Hardcode colors: <code className="text-xs bg-muted px-1 rounded line-through">text-white</code></li>
                  <li>Use hex/rgb values directly in components</li>
                  <li>Create one-off color values</li>
                </ul>
              </div>
            </div>
          </DocSection>
        </TabsContent>

        {/* ── Typography ── */}
        <TabsContent value="typography" className="space-y-6">
          <DocSection title="Type Scale" description="Consistent typographic hierarchy across the application.">
            <div className="space-y-4 divide-y divide-border">
              {[
                { cls: "text-display", label: "Display", spec: "36px / Bold" },
                { cls: "text-h1", label: "Heading 1", spec: "24px / Bold" },
                { cls: "text-h2", label: "Heading 2", spec: "20px / Semibold" },
                { cls: "text-h3", label: "Heading 3", spec: "18px / Semibold" },
                { cls: "text-h4", label: "Heading 4", spec: "16px / Semibold" },
                { cls: "text-body-lg", label: "Body Large", spec: "16px / Regular" },
                { cls: "text-body", label: "Body", spec: "14px / Regular" },
                { cls: "text-body-compact", label: "Body Compact", spec: "13px / Regular" },
                { cls: "text-small", label: "Small", spec: "13px / Regular" },
                { cls: "text-caption", label: "Caption", spec: "12px / Regular" },
              ].map((t) => (
                <div key={t.cls} className="flex items-baseline justify-between py-3">
                  <p className={`${t.cls} text-foreground`}>The quick brown fox</p>
                  <div className="flex items-center gap-3">
                    <code className="text-xs font-mono text-muted-foreground bg-muted px-2 py-0.5 rounded">{t.cls}</code>
                    <span className="text-xs text-muted-foreground w-28 text-right">{t.spec}</span>
                  </div>
                </div>
              ))}
            </div>
          </DocSection>
        </TabsContent>

        {/* ── Spacing ── */}
        <TabsContent value="spacing" className="space-y-6">
          <DocSection title="Spacing Scale" description="Consistent spacing using CSS custom properties.">
            <div className="divide-y divide-border">
              <TokenRow label="Layout Padding" value="--layout-padding: 32px" preview={<div className="h-4 bg-primary/20 rounded" style={{ width: 32 }} />} />
              <TokenRow label="Card Padding" value="--card-padding: 24px" preview={<div className="h-4 bg-primary/20 rounded" style={{ width: 24 }} />} />
              <TokenRow label="Component Gap" value="--component-gap: 16px" preview={<div className="h-4 bg-primary/20 rounded" style={{ width: 16 }} />} />
              <TokenRow label="Button Padding X" value="--button-padding-x: 20px" preview={<div className="h-4 bg-primary/20 rounded" style={{ width: 20 }} />} />
              <TokenRow label="Input Padding X" value="--input-padding-x: 16px" preview={<div className="h-4 bg-primary/20 rounded" style={{ width: 16 }} />} />
              <TokenRow label="Table Cell Padding" value="--table-cell-padding: 16px" preview={<div className="h-4 bg-primary/20 rounded" style={{ width: 16 }} />} />
            </div>
          </DocSection>

          <DocSection title="Sizing Tokens" description="Consistent component heights and widths.">
            <div className="divide-y divide-border">
              <TokenRow label="Button Height" value="--button-height: 44px" />
              <TokenRow label="Input Height" value="--input-height: 44px" />
              <TokenRow label="Table Row Height" value="--table-row-height: 64px" />
              <TokenRow label="Navbar Height" value="--navbar-height: 80px" />
              <TokenRow label="Sidebar Width" value="--sidebar-width: 280px" />
              <TokenRow label="Icon Size" value="--icon-size: 20px" />
            </div>
          </DocSection>
        </TabsContent>

        {/* ── Radii & Shadows ── */}
        <TabsContent value="radii" className="space-y-6">
          <DocSection title="Border Radius" description="Consistent roundness for UI elements.">
            <div className="flex flex-wrap gap-8 items-end">
              {[
                { label: "sm", cls: "rounded-sm", size: "2px" },
                { label: "md", cls: "rounded-md", size: "6px" },
                { label: "lg", cls: "rounded-lg", size: "8px" },
                { label: "xl", cls: "rounded-xl", size: "12px" },
                { label: "full", cls: "rounded-full", size: "9999px" },
              ].map((r) => (
                <div key={r.label} className="flex flex-col items-center gap-2">
                  <div className={`w-16 h-16 bg-primary ${r.cls}`} />
                  <span className="text-sm font-medium text-foreground">{r.label}</span>
                  <span className="text-xs text-muted-foreground">{r.size}</span>
                </div>
              ))}
            </div>
          </DocSection>

          <DocSection title="Shadows" description="Elevation layers for depth.">
            <div className="flex flex-wrap gap-8 items-end">
              {["shadow-xs", "shadow-sm", "shadow-md", "shadow-lg", "shadow-xl"].map((s) => (
                <div key={s} className="flex flex-col items-center gap-2">
                  <div className={`w-20 h-20 bg-card border border-border rounded-lg ${s}`} />
                  <code className="text-xs text-muted-foreground">{s}</code>
                </div>
              ))}
            </div>
          </DocSection>
        </TabsContent>

        {/* ── Guidelines ── */}
        <TabsContent value="guidelines" className="space-y-6">
          <DocSection title="Component Architecture" description="How to build components using the design system.">
            <div className="space-y-4">
              <div className="bg-muted/50 rounded-lg p-4">
                <h4 className="text-sm font-semibold text-foreground mb-2">File Structure</h4>
                <pre className="text-xs font-mono text-muted-foreground whitespace-pre">{`src/components/
├── ui/          # Base primitives (Button, Input, Card)
├── layout/      # Layout components (Sidebar, Navbar, Grid)
├── forms/       # Form-specific (DatePicker, FileUpload)
├── data/        # Data display (Table, Charts, Stats)
├── admin/       # Admin tools (AuditLog, HealthIndicator)
└── upload/      # Upload components (FileUploader, Queue)`}</pre>
              </div>

              <div className="bg-muted/50 rounded-lg p-4">
                <h4 className="text-sm font-semibold text-foreground mb-2">Naming Conventions</h4>
                <ul className="text-sm text-muted-foreground space-y-1 list-disc list-inside">
                  <li>Components: <code className="text-xs bg-muted px-1 rounded">PascalCase.tsx</code></li>
                  <li>Hooks: <code className="text-xs bg-muted px-1 rounded">use-kebab-case.ts</code></li>
                  <li>Services: <code className="text-xs bg-muted px-1 rounded">camelCaseService.ts</code></li>
                  <li>Stores: <code className="text-xs bg-muted px-1 rounded">kebab-case-store.ts</code></li>
                </ul>
              </div>

              <div className="bg-muted/50 rounded-lg p-4">
                <h4 className="text-sm font-semibold text-foreground mb-2">API Layer</h4>
                <pre className="text-xs font-mono text-muted-foreground whitespace-pre">{`// All API calls use typed client with Zod validation
import { apiClient } from '@/lib/api/client';
import { userSchema } from '@/lib/api/schemas';

// Schema validates response in development mode
const user = await apiClient.get('/users/1', userSchema);

// Input validation before sending
import { loginRequestSchema } from '@/lib/api/schemas';
loginRequestSchema.parse({ username, password });`}</pre>
              </div>

              <div className="bg-muted/50 rounded-lg p-4">
                <h4 className="text-sm font-semibold text-foreground mb-2">Error Handling</h4>
                <pre className="text-xs font-mono text-muted-foreground whitespace-pre">{`import {  ErrorBoundary  } from '/ui';
import {  LoadingBoundary  } from '/ui';

<ErrorBoundary>
  <LoadingBoundary variant="card">
    <MyAsyncComponent />
  </LoadingBoundary>
</ErrorBoundary>`}</pre>
              </div>
            </div>
          </DocSection>
        </TabsContent>
      </Tabs>
    </div>
  );
}
