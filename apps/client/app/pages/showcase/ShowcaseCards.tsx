import { PageHeader } from "@/components/layout/page-header";
import {
  NexusCard,
  NexusCardHeader,
  NexusCardTitle,
  NexusCardDescription,
  NexusCardContent,
  NexusCardFooter,
} from "@casbin/ui";
import { StatCard } from "@/components/patterns/stat-card";
import { NexusButton } from "@casbin/ui";
import { Switch } from "@casbin/ui";
import { Checkbox } from "@casbin/ui";
import { Users, ShieldCheck, Activity, TrendingUp } from "lucide-react";

export default function ShowcaseCards() {
  return (
    <div className="max-w-5xl space-y-10">
      <PageHeader
        title="Cards"
        description="NexusCard and StatCard components."
      />

      <section className="space-y-4">
        <h2 className="text-h2 text-foreground">Basic Cards</h2>
        <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
          <NexusCard>
            <NexusCardHeader>
              <NexusCardTitle>Basic Card</NexusCardTitle>
              <NexusCardDescription>
                With header, content, and footer.
              </NexusCardDescription>
            </NexusCardHeader>
            <NexusCardContent>
              <p className="text-body text-muted-foreground">
                Card body content here.
              </p>
            </NexusCardContent>
            <NexusCardFooter className="justify-end gap-2">
              <NexusButton variant="outline" size="sm">
                Cancel
              </NexusButton>
              <NexusButton size="sm">Save</NexusButton>
            </NexusCardFooter>
          </NexusCard>
          <NexusCard>
            <NexusCardHeader>
              <NexusCardTitle>Interactive</NexusCardTitle>
              <NexusCardDescription>With controls.</NexusCardDescription>
            </NexusCardHeader>
            <NexusCardContent className="space-y-4">
              <div className="flex items-center justify-between">
                <span className="text-body">Notifications</span>
                <Switch />
              </div>
              <div className="flex items-center gap-2">
                <Checkbox id="terms2" />
                <label
                  htmlFor="terms2"
                  className="text-body text-muted-foreground"
                >
                  Accept terms
                </label>
              </div>
            </NexusCardContent>
          </NexusCard>
        </div>
      </section>

      <section className="space-y-4">
        <h2 className="text-h2 text-foreground">Stat Cards</h2>
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
          <StatCard
            title="Total Users"
            value="2,847"
            trend={{ value: 12.5, label: "vs last month" }}
            icon={Users}
          />
          <StatCard
            title="Active Roles"
            value="18"
            trend={{ value: 2, label: "new" }}
            icon={ShieldCheck}
          />
          <StatCard
            title="Activity"
            value="94.2%"
            trend={{ value: -1.3, label: "vs last week" }}
            icon={Activity}
          />
          <StatCard
            title="Growth"
            value="+340"
            trend={{ value: 8.1, label: "this quarter" }}
            icon={TrendingUp}
          />
        </div>
      </section>
    </div>
  );
}
