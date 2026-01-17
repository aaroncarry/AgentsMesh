"use client";

import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import Link from "next/link";
import {
  Search,
  Plus,
  Tag,
  Calendar,
  Users,
  Power,
  PowerOff,
  Trash2,
  MoreHorizontal,
  ChevronLeft,
  ChevronRight,
} from "lucide-react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
  listPromoCodes,
  activatePromoCode,
  deactivatePromoCode,
  deletePromoCode,
  PromoCode,
  PromoCodeType,
} from "@/lib/api/admin";
import { formatDate } from "@/lib/utils";

const typeLabels: Record<PromoCodeType, string> = {
  media: "Media",
  partner: "Partner",
  campaign: "Campaign",
  internal: "Internal",
  referral: "Referral",
};

const typeColors: Record<PromoCodeType, "default" | "secondary" | "outline" | "destructive"> = {
  media: "default",
  partner: "secondary",
  campaign: "outline",
  internal: "destructive",
  referral: "default",
};

export default function PromoCodesPage() {
  const [search, setSearch] = useState("");
  const [typeFilter, setTypeFilter] = useState<string>("all");
  const [statusFilter, setStatusFilter] = useState<string>("all");
  const [page, setPage] = useState(1);
  const pageSize = 20;
  const queryClient = useQueryClient();

  const { data, isLoading } = useQuery({
    queryKey: ["promo-codes", search, typeFilter, statusFilter, page],
    queryFn: () =>
      listPromoCodes({
        search: search || undefined,
        type: typeFilter !== "all" ? (typeFilter as PromoCodeType) : undefined,
        is_active: statusFilter === "all" ? undefined : statusFilter === "active",
        page,
        page_size: pageSize,
      }),
  });

  const activateMutation = useMutation({
    mutationFn: activatePromoCode,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["promo-codes"] });
      toast.success("Promo code activated");
    },
    onError: (err: { error: string }) => {
      toast.error(err.error || "Failed to activate promo code");
    },
  });

  const deactivateMutation = useMutation({
    mutationFn: deactivatePromoCode,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["promo-codes"] });
      toast.success("Promo code deactivated");
    },
    onError: (err: { error: string }) => {
      toast.error(err.error || "Failed to deactivate promo code");
    },
  });

  const deleteMutation = useMutation({
    mutationFn: deletePromoCode,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["promo-codes"] });
      toast.success("Promo code deleted");
    },
    onError: (err: { error: string }) => {
      toast.error(err.error || "Failed to delete promo code");
    },
  });

  const handleDelete = (code: PromoCode) => {
    if (
      confirm(
        `Are you sure you want to delete "${code.code}"? This action cannot be undone.`
      )
    ) {
      deleteMutation.mutate(code.id);
    }
  };

  const promoCodes = data?.data || [];
  const total = data?.total || 0;
  const totalPages = data?.total_pages || 1;

  const getRemainingUses = (code: PromoCode) => {
    if (code.max_uses === null) return "Unlimited";
    const remaining = code.max_uses - code.used_count;
    return `${remaining}/${code.max_uses}`;
  };

  const isExpired = (code: PromoCode) => {
    if (!code.expires_at) return false;
    return new Date(code.expires_at) < new Date();
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">Promo Codes</h1>
          <p className="text-sm text-muted-foreground">
            Manage promotional codes for subscriptions
          </p>
        </div>
        <Link href="/promo-codes/new">
          <Button>
            <Plus className="mr-2 h-4 w-4" />
            Create Promo Code
          </Button>
        </Link>
      </div>

      {/* Filters */}
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center">
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            placeholder="Search by code or name..."
            value={search}
            onChange={(e) => {
              setSearch(e.target.value);
              setPage(1);
            }}
            className="pl-10"
          />
        </div>
        <Select
          value={typeFilter}
          onValueChange={(value) => {
            setTypeFilter(value);
            setPage(1);
          }}
        >
          <SelectTrigger className="w-40">
            <SelectValue placeholder="All Types" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">All Types</SelectItem>
            <SelectItem value="media">Media</SelectItem>
            <SelectItem value="partner">Partner</SelectItem>
            <SelectItem value="campaign">Campaign</SelectItem>
            <SelectItem value="internal">Internal</SelectItem>
            <SelectItem value="referral">Referral</SelectItem>
          </SelectContent>
        </Select>
        <Select
          value={statusFilter}
          onValueChange={(value) => {
            setStatusFilter(value);
            setPage(1);
          }}
        >
          <SelectTrigger className="w-40">
            <SelectValue placeholder="All Status" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">All Status</SelectItem>
            <SelectItem value="active">Active</SelectItem>
            <SelectItem value="inactive">Inactive</SelectItem>
          </SelectContent>
        </Select>
      </div>

      {/* Table */}
      <div className="rounded-lg border border-border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Code</TableHead>
              <TableHead>Name</TableHead>
              <TableHead>Type</TableHead>
              <TableHead>Plan</TableHead>
              <TableHead>Duration</TableHead>
              <TableHead>Uses</TableHead>
              <TableHead>Expires</TableHead>
              <TableHead>Status</TableHead>
              <TableHead className="w-12"></TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {isLoading ? (
              Array.from({ length: 5 }).map((_, i) => (
                <TableRow key={i}>
                  <TableCell colSpan={9}>
                    <div className="h-12 animate-pulse rounded bg-muted" />
                  </TableCell>
                </TableRow>
              ))
            ) : promoCodes.length === 0 ? (
              <TableRow>
                <TableCell colSpan={9} className="py-8 text-center text-muted-foreground">
                  No promo codes found
                </TableCell>
              </TableRow>
            ) : (
              promoCodes.map((code) => (
                <TableRow key={code.id}>
                  <TableCell>
                    <Link
                      href={`/promo-codes/${code.id}`}
                      className="flex items-center gap-2 font-mono font-medium hover:text-primary"
                    >
                      <Tag className="h-4 w-4 text-muted-foreground" />
                      {code.code}
                    </Link>
                  </TableCell>
                  <TableCell>{code.name}</TableCell>
                  <TableCell>
                    <Badge variant={typeColors[code.type]}>
                      {typeLabels[code.type]}
                    </Badge>
                  </TableCell>
                  <TableCell className="capitalize">{code.plan_name}</TableCell>
                  <TableCell>{code.duration_months} months</TableCell>
                  <TableCell>
                    <div className="flex items-center gap-1">
                      <Users className="h-3 w-3 text-muted-foreground" />
                      {getRemainingUses(code)}
                    </div>
                  </TableCell>
                  <TableCell>
                    {code.expires_at ? (
                      <div className="flex items-center gap-1">
                        <Calendar className="h-3 w-3 text-muted-foreground" />
                        <span className={isExpired(code) ? "text-destructive" : ""}>
                          {formatDate(code.expires_at)}
                        </span>
                      </div>
                    ) : (
                      <span className="text-muted-foreground">Never</span>
                    )}
                  </TableCell>
                  <TableCell>
                    {code.is_active && !isExpired(code) ? (
                      <Badge variant="success">Active</Badge>
                    ) : isExpired(code) ? (
                      <Badge variant="destructive">Expired</Badge>
                    ) : (
                      <Badge variant="secondary">Inactive</Badge>
                    )}
                  </TableCell>
                  <TableCell>
                    <DropdownMenu>
                      <DropdownMenuTrigger asChild>
                        <Button variant="ghost" size="icon">
                          <MoreHorizontal className="h-4 w-4" />
                        </Button>
                      </DropdownMenuTrigger>
                      <DropdownMenuContent align="end">
                        <DropdownMenuItem asChild>
                          <Link href={`/promo-codes/${code.id}`}>View Details</Link>
                        </DropdownMenuItem>
                        <DropdownMenuSeparator />
                        {code.is_active ? (
                          <DropdownMenuItem
                            onClick={() => deactivateMutation.mutate(code.id)}
                            disabled={deactivateMutation.isPending}
                          >
                            <PowerOff className="mr-2 h-4 w-4" />
                            Deactivate
                          </DropdownMenuItem>
                        ) : (
                          <DropdownMenuItem
                            onClick={() => activateMutation.mutate(code.id)}
                            disabled={activateMutation.isPending}
                          >
                            <Power className="mr-2 h-4 w-4" />
                            Activate
                          </DropdownMenuItem>
                        )}
                        <DropdownMenuSeparator />
                        <DropdownMenuItem
                          onClick={() => handleDelete(code)}
                          disabled={deleteMutation.isPending}
                          className="text-destructive focus:text-destructive"
                        >
                          <Trash2 className="mr-2 h-4 w-4" />
                          Delete
                        </DropdownMenuItem>
                      </DropdownMenuContent>
                    </DropdownMenu>
                  </TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </div>

      {/* Pagination */}
      {totalPages > 1 && (
        <div className="flex items-center justify-between">
          <p className="text-sm text-muted-foreground">
            Showing {(page - 1) * pageSize + 1} to{" "}
            {Math.min(page * pageSize, total)} of {total} promo codes
          </p>
          <div className="flex items-center gap-2">
            <Button
              variant="outline"
              size="icon"
              onClick={() => setPage(page - 1)}
              disabled={page <= 1}
            >
              <ChevronLeft className="h-4 w-4" />
            </Button>
            <span className="text-sm">
              Page {page} of {totalPages}
            </span>
            <Button
              variant="outline"
              size="icon"
              onClick={() => setPage(page + 1)}
              disabled={page >= totalPages}
            >
              <ChevronRight className="h-4 w-4" />
            </Button>
          </div>
        </div>
      )}
    </div>
  );
}
