"use client";

import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import Link from "next/link";
import { Search, Trash2, Users, ExternalLink } from "lucide-react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { listOrganizations, deleteOrganization, Organization } from "@/lib/api/admin";
import { formatDate } from "@/lib/utils";

export default function OrganizationsPage() {
  const [search, setSearch] = useState("");
  const [page, setPage] = useState(1);
  const queryClient = useQueryClient();

  const { data, isLoading } = useQuery({
    queryKey: ["organizations", { search, page }],
    queryFn: () => listOrganizations({ search, page, page_size: 20 }),
  });

  const deleteMutation = useMutation({
    mutationFn: deleteOrganization,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["organizations"] });
      toast.success("Organization deleted successfully");
    },
    onError: (err: { error: string }) => {
      toast.error(err.error || "Failed to delete organization");
    },
  });

  const handleDelete = (org: Organization) => {
    if (confirm(`Are you sure you want to delete "${org.name}"? This action cannot be undone.`)) {
      deleteMutation.mutate(org.id);
    }
  };

  return (
    <div className="space-y-4">
      {/* Search */}
      <div className="flex items-center gap-4">
        <div className="relative flex-1 max-w-sm">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            placeholder="Search organizations..."
            value={search}
            onChange={(e) => {
              setSearch(e.target.value);
              setPage(1);
            }}
            className="pl-9"
          />
        </div>
      </div>

      {/* Organizations Table */}
      <Card>
        <CardHeader>
          <CardTitle>Organizations ({data?.total || 0})</CardTitle>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="space-y-3">
              {Array.from({ length: 5 }).map((_, i) => (
                <div key={i} className="h-16 animate-pulse rounded-lg bg-muted" />
              ))}
            </div>
          ) : (
            <div className="space-y-2">
              {data?.data.map((org) => (
                <OrgRow
                  key={org.id}
                  org={org}
                  onDelete={() => handleDelete(org)}
                />
              ))}
              {data?.data.length === 0 && (
                <p className="py-8 text-center text-muted-foreground">
                  No organizations found
                </p>
              )}
            </div>
          )}

          {/* Pagination */}
          {data && data.total_pages > 1 && (
            <div className="mt-4 flex items-center justify-between">
              <p className="text-sm text-muted-foreground">
                Page {data.page} of {data.total_pages}
              </p>
              <div className="flex gap-2">
                <Button
                  variant="outline"
                  size="sm"
                  disabled={page === 1}
                  onClick={() => setPage(page - 1)}
                >
                  Previous
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  disabled={page >= data.total_pages}
                  onClick={() => setPage(page + 1)}
                >
                  Next
                </Button>
              </div>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}

function OrgRow({
  org,
  onDelete,
}: {
  org: Organization;
  onDelete: () => void;
}) {
  return (
    <div className="flex items-center justify-between rounded-lg border border-border p-4">
      <div className="flex items-center gap-4">
        {org.logo_url ? (
          <img
            src={org.logo_url}
            alt={org.name}
            className="h-10 w-10 rounded-lg object-cover"
          />
        ) : (
          <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-primary/20 text-sm font-medium text-primary">
            {org.name[0].toUpperCase()}
          </div>
        )}
        <div>
          <div className="flex items-center gap-2">
            <span className="font-medium">{org.name}</span>
            <span className="text-sm text-muted-foreground">({org.slug})</span>
          </div>
          {org.description && (
            <p className="text-sm text-muted-foreground line-clamp-1">
              {org.description}
            </p>
          )}
        </div>
      </div>
      <div className="flex items-center gap-4">
        <div className="text-right text-xs text-muted-foreground">
          <p>Created {formatDate(org.created_at)}</p>
        </div>
        <div className="flex gap-1">
          <Link href={`/organizations/${org.id}`}>
            <Button variant="ghost" size="icon" title="View details">
              <ExternalLink className="h-4 w-4" />
            </Button>
          </Link>
          <Button
            variant="ghost"
            size="icon"
            onClick={onDelete}
            title="Delete organization"
            className="text-destructive hover:text-destructive"
          >
            <Trash2 className="h-4 w-4" />
          </Button>
        </div>
      </div>
    </div>
  );
}
