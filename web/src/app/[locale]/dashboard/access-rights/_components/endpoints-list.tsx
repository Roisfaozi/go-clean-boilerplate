"use client";

import { useAccessRights } from "./access-rights-context";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "~/components/ui/table";
import { Badge } from "~/components/ui/badge";
import { Button } from "~/components/ui/button";
import { Icon } from "~/components/shared/icon";

export function EndpointsList() {
  const { endpoints, isLoading, isProcessing, deleteEndpoint } =
    useAccessRights();

  return (
    <div className="rounded-md border">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Method</TableHead>
            <TableHead>Path</TableHead>
            <TableHead>Created At</TableHead>
            <TableHead className="text-right">Actions</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {isLoading && endpoints.length === 0 ? (
            <TableRow>
              <TableCell colSpan={4} className="h-24 text-center">
                Loading endpoints...
              </TableCell>
            </TableRow>
          ) : endpoints.length === 0 ? (
            <TableRow>
              <TableCell colSpan={4} className="h-24 text-center">
                No endpoints registered.
              </TableCell>
            </TableRow>
          ) : (
            endpoints.map((ep) => (
              <TableRow key={ep.id}>
                <TableCell>
                  <Badge variant="outline" className="font-mono">
                    {ep.method}
                  </Badge>
                </TableCell>
                <TableCell className="font-mono text-sm">{ep.path}</TableCell>
                <TableCell className="text-muted-foreground text-xs">
                  {new Date(ep.created_at).toLocaleDateString()}
                </TableCell>
                <TableCell className="text-right">
                  <Button
                    variant="ghost"
                    size="icon"
                    disabled={isProcessing === ep.id}
                    onClick={() => deleteEndpoint(ep.id)}
                  >
                    {isProcessing === ep.id ? (
                      <Icon
                        name="Loader"
                        className="text-muted-foreground h-4 w-4 animate-spin"
                      />
                    ) : (
                      <Icon
                        name="Trash2"
                        className="text-destructive h-4 w-4"
                      />
                    )}
                  </Button>
                </TableCell>
              </TableRow>
            ))
          )}
        </TableBody>
      </Table>
    </div>
  );
}
