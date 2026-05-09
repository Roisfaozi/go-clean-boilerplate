import { useState } from "react";
import { NexusCard } from "@casbin/ui";
import { NexusButton } from "@casbin/ui";
import { NexusInput } from "@casbin/ui";
import { Badge } from "@casbin/ui";
import { useToast } from "@casbin/ui";
import { Link2, Unlink, Search, Plus } from "lucide-react";

interface AccessRight {
  id: string;
  name: string;
  endpoints: string[];
}

interface AccessRightManagerProps {
  initialRights?: AccessRight[];
}

const defaultRights: AccessRight[] = [
  {
    id: "1",
    name: "User Management",
    endpoints: ["GET /api/users", "POST /api/users", "PUT /api/users/:id", "DELETE /api/users/:id"],
  },
  { id: "2", name: "Role Management", endpoints: ["GET /api/roles", "POST /api/roles"] },
  { id: "3", name: "Organization Read", endpoints: ["GET /api/organizations"] },
  {
    id: "4",
    name: "Project Admin",
    endpoints: [
      "GET /api/projects",
      "POST /api/projects",
      "PUT /api/projects/:id",
      "DELETE /api/projects/:id",
    ],
  },
  { id: "5", name: "Audit Viewer", endpoints: ["GET /api/audit-logs"] },
];

const availableEndpoints = [
  "GET /api/users",
  "POST /api/users",
  "PUT /api/users/:id",
  "DELETE /api/users/:id",
  "GET /api/roles",
  "POST /api/roles",
  "PUT /api/roles/:id",
  "DELETE /api/roles/:id",
  "GET /api/organizations",
  "POST /api/organizations",
  "PUT /api/organizations/:id",
  "GET /api/projects",
  "POST /api/projects",
  "PUT /api/projects/:id",
  "DELETE /api/projects/:id",
  "GET /api/permissions",
  "POST /api/permissions",
  "DELETE /api/permissions/:id",
  "GET /api/audit-logs",
];

export function AccessRightManager({ initialRights }: AccessRightManagerProps) {
  const { toast } = useToast();
  const [rights, setRights] = useState<AccessRight[]>(initialRights ?? defaultRights);
  const [search, setSearch] = useState("");
  const [selectedRight, setSelectedRight] = useState<string | null>(null);

  const filtered = rights.filter(
    (r) =>
      r.name.toLowerCase().includes(search.toLowerCase()) ||
      r.endpoints.some((e) => e.toLowerCase().includes(search.toLowerCase())),
  );

  const unlinkEndpoint = (rightId: string, endpoint: string) => {
    setRights((prev) =>
      prev.map((r) =>
        r.id === rightId ? { ...r, endpoints: r.endpoints.filter((e) => e !== endpoint) } : r,
      ),
    );
    toast({ title: "Endpoint Unlinked", description: endpoint });
  };

  const linkEndpoint = (rightId: string, endpoint: string) => {
    setRights((prev) =>
      prev.map((r) =>
        r.id === rightId && !r.endpoints.includes(endpoint)
          ? { ...r, endpoints: [...r.endpoints, endpoint] }
          : r,
      ),
    );
    toast({ title: "Endpoint Linked", description: endpoint });
  };

  const selected = rights.find((r) => r.id === selectedRight);
  const unlinkedEndpoints = selected
    ? availableEndpoints.filter((e) => !selected.endpoints.includes(e))
    : [];

  return (
    <div className="grid grid-cols-1 gap-4 lg:grid-cols-3">
      {/* Rights list */}
      <NexusCard className="lg:col-span-1">
        <div className="border-border border-b p-4">
          <h3 className="text-foreground text-sm font-semibold">Access Rights</h3>
          <div className="relative mt-2">
            <Search className="text-muted-foreground absolute top-1/2 left-3 h-4 w-4 -translate-y-1/2" />
            <NexusInput
              placeholder="Search rights..."
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              className="h-8 pl-9 text-sm"
            />
          </div>
        </div>
        <div className="divide-border max-h-[400px] divide-y overflow-y-auto">
          {filtered.map((right) => (
            <button
              key={right.id}
              onClick={() => setSelectedRight(right.id)}
              className={`hover:bg-muted/50 w-full px-4 py-3 text-left text-sm transition-colors ${
                selectedRight === right.id ? "bg-primary/5 border-l-primary border-l-2" : ""
              }`}
            >
              <div className="text-foreground font-medium">{right.name}</div>
              <div className="text-muted-foreground mt-0.5 text-xs">
                {right.endpoints.length} endpoints
              </div>
            </button>
          ))}
        </div>
      </NexusCard>

      {/* Linked endpoints */}
      <NexusCard className="lg:col-span-2">
        {selected ? (
          <>
            <div className="border-border flex items-center justify-between border-b p-4">
              <div>
                <h3 className="text-foreground text-sm font-semibold">{selected.name}</h3>
                <p className="text-muted-foreground mt-0.5 text-xs">Manage linked API endpoints</p>
              </div>
              <Badge variant="secondary">{selected.endpoints.length} linked</Badge>
            </div>
            <div className="space-y-4 p-4">
              {/* Linked */}
              <div>
                <p className="text-muted-foreground mb-2 text-xs font-semibold tracking-wider uppercase">
                  Linked Endpoints
                </p>
                <div className="space-y-1.5">
                  {selected.endpoints.map((ep) => {
                    const method = ep.split(" ")[0];
                    const methodColor: Record<string, string> = {
                      GET: "text-success",
                      POST: "text-info",
                      PUT: "text-warning",
                      DELETE: "text-danger",
                    };
                    return (
                      <div
                        key={ep}
                        className="bg-muted/30 group flex items-center justify-between rounded-md px-3 py-2"
                      >
                        <code className="text-xs">
                          <span className={methodColor[method] ?? "text-foreground"}>{method}</span>{" "}
                          <span className="text-foreground">{ep.slice(method.length + 1)}</span>
                        </code>
                        <NexusButton
                          variant="ghost"
                          size="icon"
                          className="h-7 w-7 opacity-0 transition-opacity group-hover:opacity-100"
                          onClick={() => unlinkEndpoint(selected.id, ep)}
                        >
                          <Unlink className="text-destructive h-3.5 w-3.5" />
                        </NexusButton>
                      </div>
                    );
                  })}
                </div>
              </div>

              {/* Available to link */}
              {unlinkedEndpoints.length > 0 && (
                <div>
                  <p className="text-muted-foreground mb-2 text-xs font-semibold tracking-wider uppercase">
                    Available Endpoints
                  </p>
                  <div className="space-y-1.5">
                    {unlinkedEndpoints.map((ep) => {
                      const method = ep.split(" ")[0];
                      const methodColor: Record<string, string> = {
                        GET: "text-success",
                        POST: "text-info",
                        PUT: "text-warning",
                        DELETE: "text-danger",
                      };
                      return (
                        <div
                          key={ep}
                          className="border-border group flex items-center justify-between rounded-md border border-dashed px-3 py-2"
                        >
                          <code className="text-muted-foreground text-xs">
                            <span className={methodColor[method] ?? ""}>{method}</span>{" "}
                            {ep.slice(method.length + 1)}
                          </code>
                          <NexusButton
                            variant="ghost"
                            size="icon"
                            className="h-7 w-7 opacity-0 transition-opacity group-hover:opacity-100"
                            onClick={() => linkEndpoint(selected.id, ep)}
                          >
                            <Link2 className="text-primary h-3.5 w-3.5" />
                          </NexusButton>
                        </div>
                      );
                    })}
                  </div>
                </div>
              )}
            </div>
          </>
        ) : (
          <div className="text-muted-foreground flex h-64 items-center justify-center text-sm">
            Select an access right to manage endpoints
          </div>
        )}
      </NexusCard>
    </div>
  );
}
