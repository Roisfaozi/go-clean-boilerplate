"use client";

import { useAccessRights } from "./access-rights-context";
import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger,
} from "~/components/ui/accordion";
import { Badge } from "~/components/ui/badge";
import { Button } from "~/components/ui/button";
import { Checkbox } from "~/components/ui/checkbox";
import { Icon } from "~/components/shared/icon";
import { AccessRight, Endpoint } from "~/lib/api/access";
import { useMemo } from "react";

export function AccessRightsList() {
  const { accessRights, endpoints, isLoading } = useAccessRights();

  const groupedEndpoints = useMemo(() => {
    return endpoints.reduce(
      (acc, ep) => {
        const segments = ep.path.split("/");
        const groupName = segments[3] || "other";
        if (!acc[groupName]) acc[groupName] = [];
        acc[groupName].push(ep);
        return acc;
      },
      {} as Record<string, Endpoint[]>
    );
  }, [endpoints]);

  if (isLoading && accessRights.length === 0) {
    return <div className="py-12 text-center">Loading access rights...</div>;
  }

  if (accessRights.length === 0) {
    return (
      <div className="py-12 text-center">
        <p className="text-muted-foreground italic">
          No access rights created yet.
        </p>
      </div>
    );
  }

  return (
    <div className="bg-card rounded-md border">
      <Accordion type="multiple" className="w-full">
        {accessRights.map((ar) => (
          <AccessRightItem
            key={ar.id}
            accessRight={ar}
            groupedEndpoints={groupedEndpoints}
          />
        ))}
      </Accordion>
    </div>
  );
}

function AccessRightItem({
  accessRight,
  groupedEndpoints,
}: {
  accessRight: AccessRight;
  groupedEndpoints: Record<string, Endpoint[]>;
}) {
  const { deleteAccessRight, isProcessing, toggleLink } = useAccessRights();

  return (
    <AccordionItem value={accessRight.id} className="px-6">
      <div className="flex items-center">
        <AccordionTrigger className="py-6 hover:no-underline">
          <div className="flex flex-col items-start gap-1 text-left">
            <span className="text-lg font-semibold">{accessRight.name}</span>
            <span className="text-muted-foreground text-sm font-normal">
              {accessRight.description || "No description"} •{" "}
              {accessRight.endpoints?.length || 0} endpoints
            </span>
          </div>
        </AccordionTrigger>
        <Button
          variant="ghost"
          size="icon"
          className="text-destructive ml-4 h-8 w-8"
          disabled={isProcessing === accessRight.id}
          onClick={(e) => {
            e.stopPropagation();
            deleteAccessRight(accessRight.id);
          }}
        >
          {isProcessing === accessRight.id ? (
            <Icon name="Loader" className="h-4 w-4 animate-spin" />
          ) : (
            <Icon name="Trash2" className="h-4 w-4" />
          )}
        </Button>
      </div>
      <AccordionContent className="pb-6">
        <div className="bg-muted/30 rounded-lg border p-4">
          <Accordion type="multiple" className="w-full border-none">
            {Object.entries(groupedEndpoints)
              .sort(([a], [b]) => a.localeCompare(b))
              .map(([groupName, eps]) => (
                <EndpointGroup
                  key={`${accessRight.id}-${groupName}`}
                  groupName={groupName}
                  eps={eps}
                  accessRight={accessRight}
                  toggleLink={toggleLink}
                  isProcessing={isProcessing}
                />
              ))}
          </Accordion>
        </div>
      </AccordionContent>
    </AccordionItem>
  );
}

function EndpointGroup({
  groupName,
  eps,
  accessRight,
  toggleLink,
  isProcessing,
}: {
  groupName: string;
  eps: Endpoint[];
  accessRight: AccessRight;
  toggleLink: any;
  isProcessing: string | null;
}) {
  const selectedInGroup = eps.filter((ep) =>
    accessRight.endpoints?.some((e) => e.id === ep.id)
  ).length;

  return (
    <AccordionItem value={groupName} className="border-none">
      <AccordionTrigger className="py-2 hover:no-underline">
        <div className="flex flex-1 items-center justify-between pr-4">
          <div className="flex items-center gap-2">
            <span className="text-xs font-medium capitalize">{groupName}</span>
            <Badge variant="outline" className="h-4 text-[10px]">
              {selectedInGroup} / {eps.length}
            </Badge>
          </div>
        </div>
      </AccordionTrigger>
      <AccordionContent className="pt-2">
        <div className="grid grid-cols-1 gap-3 md:grid-cols-2 lg:grid-cols-3">
          {eps.map((ep) => {
            const isLinked = accessRight.endpoints?.some((e) => e.id === ep.id);
            const processId = `${accessRight.id}-${ep.id}`;
            return (
              <div
                key={ep.id}
                className="hover:bg-muted/50 hover:border-border flex items-center space-x-3 rounded-md border border-transparent p-2 transition-colors"
              >
                <Checkbox
                  id={`ar-${accessRight.id}-ep-${ep.id}`}
                  checked={isLinked}
                  disabled={isProcessing === processId}
                  onCheckedChange={() =>
                    toggleLink(accessRight.id, ep.id, !!isLinked)
                  }
                />
                <label
                  htmlFor={`ar-${accessRight.id}-ep-${ep.id}`}
                  className="flex flex-1 cursor-pointer items-center gap-2 text-xs leading-none font-medium peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
                >
                  <Badge
                    variant="outline"
                    className="h-4 px-1 font-mono text-[8px]"
                  >
                    {ep.method}
                  </Badge>
                  <span
                    className="truncate font-mono text-[10px] opacity-80"
                    title={ep.path}
                  >
                    {ep.path}
                  </span>
                </label>
              </div>
            );
          })}
        </div>
      </AccordionContent>
    </AccordionItem>
  );
}
