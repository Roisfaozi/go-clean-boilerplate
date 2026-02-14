"use client";

import {
  AccessControlProvider,
  useAccessControl,
} from "./_components/access-control-context";
import { PermissionMatrixView } from "~/components/dashboard/access/permission-matrix-view";
import { PolicyEditorView } from "~/components/dashboard/access/policy-editor-view";
import { RoleCardsView } from "~/components/dashboard/access/role-cards-view";
import { RoleDialog } from "~/components/dashboard/roles/role-dialog";
import { Icon } from "~/components/shared/icon";
import { Button } from "~/components/ui/button";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "~/components/ui/tabs";

export default function AccessPage() {
  return (
    <AccessControlProvider>
      <AccessControlContent />
    </AccessControlProvider>
  );
}

function AccessControlContent() {
  const {
    activeTab,
    setActiveTab,
    roleDialogOpen,
    setRoleDialogOpen,
    handleRoleClick,
  } = useAccessControl();

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <div className="mb-1 flex items-center gap-2">
            <Icon name="Shield" className="text-primary h-6 w-6" />
            <h2 className="text-2xl font-bold tracking-tight">
              Access Control
            </h2>
          </div>
          <p className="text-muted-foreground">
            Configure role permissions and access policies
          </p>
        </div>
        <div className="flex gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={() => setRoleDialogOpen(true)}
          >
            <Icon name="Plus" className="mr-2 h-4 w-4" />
            Add Role
          </Button>
          <Button size="sm">
            <Icon name="Plus" className="mr-2 h-4 w-4" />
            Add Resource
          </Button>
        </div>
      </div>

      <Tabs value={activeTab} onValueChange={setActiveTab}>
        <TabsList>
          <TabsTrigger value="matrix" className="gap-1.5">
            <Icon name="Table" className="h-3.5 w-3.5" />
            Matrix View
          </TabsTrigger>
          <TabsTrigger value="cards" className="gap-1.5">
            <Icon name="LayoutGrid" className="h-3.5 w-3.5" />
            Role Cards
          </TabsTrigger>
          <TabsTrigger value="policy" className="gap-1.5">
            <Icon name="GitBranch" className="h-3.5 w-3.5" />
            Policy Editor
          </TabsTrigger>
        </TabsList>

        <TabsContent value="matrix" className="mt-4">
          <PermissionMatrixView onRoleClick={handleRoleClick} />
        </TabsContent>

        <TabsContent value="cards" className="mt-4">
          <RoleCardsView onRoleClick={handleRoleClick} />
        </TabsContent>

        <TabsContent value="policy" className="mt-4">
          <PolicyEditorView />
        </TabsContent>
      </Tabs>

      <RoleDialog
        open={roleDialogOpen}
        onOpenChange={setRoleDialogOpen}
        onSuccess={() => setActiveTab("cards")}
      />
    </div>
  );
}
