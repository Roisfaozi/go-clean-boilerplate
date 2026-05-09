"use client";

import { AccessRightsProvider } from "./_components/access-rights-context";
import { CreateArDialog } from "./_components/create-ar-dialog";
import { RegisterEpDialog } from "./_components/register-ep-dialog";
import { AccessRightsList } from "./_components/access-rights-list";
import { EndpointsList } from "./_components/endpoints-list";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "~/components/ui/tabs";

export default function AccessRightsPage() {
  return (
    <AccessRightsProvider>
      <div className="space-y-6">
        <div className="flex items-center justify-between">
          <div>
            <h2 className="text-2xl font-bold tracking-tight">
              Access Rights & Endpoints
            </h2>
            <p className="text-muted-foreground">
              Define resource groups and register API endpoints.
            </p>
          </div>
        </div>

        <Tabs defaultValue="access-rights" className="w-full">
          <TabsList className="grid w-full max-w-[400px] grid-cols-2">
            <TabsTrigger value="access-rights">Access Rights</TabsTrigger>
            <TabsTrigger value="endpoints">All Endpoints</TabsTrigger>
          </TabsList>

          <TabsContent value="access-rights" className="mt-4 space-y-4">
            <div className="flex justify-end">
              <CreateArDialog />
            </div>
            <AccessRightsList />
          </TabsContent>

          <TabsContent value="endpoints" className="mt-4 space-y-4">
            <div className="flex justify-end">
              <RegisterEpDialog />
            </div>
            <EndpointsList />
          </TabsContent>
        </Tabs>
      </div>
    </AccessRightsProvider>
  );
}
