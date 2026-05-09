"use client";

import { useSettings } from "./settings-context";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "~/components/ui/card";
import { Button } from "~/components/ui/button";
import { Icon } from "~/components/shared/icon";

export function DangerZoneCard() {
  const { handleDelete, isDeleting, currentOrganization } = useSettings();

  if (!currentOrganization) return null;

  return (
    <Card className="border-destructive/20 bg-destructive/5">
      <CardHeader>
        <CardTitle className="text-destructive">Danger Zone</CardTitle>
        <CardDescription>
          Permanently delete this organization and all associated data.
        </CardDescription>
      </CardHeader>
      <CardContent>
        <p className="text-muted-foreground text-sm">
          Once you delete an organization, there is no going back. Please be
          certain.
        </p>
      </CardContent>
      <CardFooter className="border-destructive/10 flex items-center justify-between border-t px-6 py-4">
        <div className="text-muted-foreground text-xs italic">
          Owned by{" "}
          {currentOrganization.owner_id === "me" ? "you" : "another admin"}
        </div>
        <Button
          variant="destructive"
          onClick={handleDelete}
          disabled={isDeleting}
        >
          {isDeleting ? (
            <Icon name="Loader" className="mr-2 h-4 w-4 animate-spin" />
          ) : (
            <Icon name="Trash2" className="mr-2 h-4 w-4" />
          )}
          Delete Organization
        </Button>
      </CardFooter>
    </Card>
  );
}
