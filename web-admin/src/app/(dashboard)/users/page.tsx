"use client";

import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { Search, MoreHorizontal, Shield, ShieldOff, UserX, UserCheck, MailCheck, MailX } from "lucide-react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  listUsers,
  disableUser,
  enableUser,
  grantAdmin,
  revokeAdmin,
  verifyUserEmail,
  unverifyUserEmail,
  User,
} from "@/lib/api/admin";
import { formatDate, formatRelativeTime } from "@/lib/utils";

export default function UsersPage() {
  const [search, setSearch] = useState("");
  const [page, setPage] = useState(1);
  const queryClient = useQueryClient();

  const { data, isLoading } = useQuery({
    queryKey: ["users", { search, page }],
    queryFn: () => listUsers({ search, page, page_size: 20 }),
  });

  const disableMutation = useMutation({
    mutationFn: disableUser,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["users"] });
      toast.success("User disabled successfully");
    },
    onError: (err: { error: string }) => {
      toast.error(err.error || "Failed to disable user");
    },
  });

  const enableMutation = useMutation({
    mutationFn: enableUser,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["users"] });
      toast.success("User enabled successfully");
    },
    onError: (err: { error: string }) => {
      toast.error(err.error || "Failed to enable user");
    },
  });

  const grantAdminMutation = useMutation({
    mutationFn: grantAdmin,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["users"] });
      toast.success("Admin privileges granted");
    },
    onError: (err: { error: string }) => {
      toast.error(err.error || "Failed to grant admin privileges");
    },
  });

  const revokeAdminMutation = useMutation({
    mutationFn: revokeAdmin,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["users"] });
      toast.success("Admin privileges revoked");
    },
    onError: (err: { error: string }) => {
      toast.error(err.error || "Failed to revoke admin privileges");
    },
  });

  const verifyEmailMutation = useMutation({
    mutationFn: verifyUserEmail,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["users"] });
      toast.success("Email verified successfully");
    },
    onError: (err: { error: string }) => {
      toast.error(err.error || "Failed to verify email");
    },
  });

  const unverifyEmailMutation = useMutation({
    mutationFn: unverifyUserEmail,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["users"] });
      toast.success("Email unverified successfully");
    },
    onError: (err: { error: string }) => {
      toast.error(err.error || "Failed to unverify email");
    },
  });

  return (
    <div className="space-y-4">
      {/* Search */}
      <div className="flex items-center gap-4">
        <div className="relative flex-1 max-w-sm">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            placeholder="Search users..."
            value={search}
            onChange={(e) => {
              setSearch(e.target.value);
              setPage(1);
            }}
            className="pl-9"
          />
        </div>
      </div>

      {/* Users Table */}
      <Card>
        <CardHeader>
          <CardTitle>Users ({data?.total || 0})</CardTitle>
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
              {data?.data.map((user) => (
                <UserRow
                  key={user.id}
                  user={user}
                  onDisable={() => disableMutation.mutate(user.id)}
                  onEnable={() => enableMutation.mutate(user.id)}
                  onGrantAdmin={() => grantAdminMutation.mutate(user.id)}
                  onRevokeAdmin={() => revokeAdminMutation.mutate(user.id)}
                  onVerifyEmail={() => verifyEmailMutation.mutate(user.id)}
                  onUnverifyEmail={() => unverifyEmailMutation.mutate(user.id)}
                />
              ))}
              {data?.data.length === 0 && (
                <p className="py-8 text-center text-muted-foreground">
                  No users found
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

function UserRow({
  user,
  onDisable,
  onEnable,
  onGrantAdmin,
  onRevokeAdmin,
  onVerifyEmail,
  onUnverifyEmail,
}: {
  user: User;
  onDisable: () => void;
  onEnable: () => void;
  onGrantAdmin: () => void;
  onRevokeAdmin: () => void;
  onVerifyEmail: () => void;
  onUnverifyEmail: () => void;
}) {
  return (
    <div className="flex items-center justify-between rounded-lg border border-border p-4">
      <div className="flex items-center gap-4">
        {user.avatar_url ? (
          <img
            src={user.avatar_url}
            alt={user.username}
            className="h-10 w-10 rounded-full"
          />
        ) : (
          <div className="flex h-10 w-10 items-center justify-center rounded-full bg-primary/20 text-sm font-medium text-primary">
            {user.username[0].toUpperCase()}
          </div>
        )}
        <div>
          <div className="flex items-center gap-2">
            <span className="font-medium">{user.name || user.username}</span>
            {user.is_system_admin && (
              <Badge variant="default" className="text-xs">
                <Shield className="mr-1 h-3 w-3" />
                Admin
              </Badge>
            )}
            {!user.is_active && (
              <Badge variant="destructive" className="text-xs">
                Disabled
              </Badge>
            )}
            {!user.is_email_verified && (
              <Badge variant="outline" className="text-xs">
                Unverified
              </Badge>
            )}
          </div>
          <p className="text-sm text-muted-foreground">{user.email}</p>
        </div>
      </div>
      <div className="flex items-center gap-4">
        <div className="text-right text-xs text-muted-foreground">
          <p>Joined {formatDate(user.created_at)}</p>
          {user.last_login_at && (
            <p>Last login {formatRelativeTime(user.last_login_at)}</p>
          )}
        </div>
        <div className="flex gap-1">
          {user.is_active ? (
            <Button
              variant="ghost"
              size="icon"
              onClick={onDisable}
              title="Disable user"
            >
              <UserX className="h-4 w-4" />
            </Button>
          ) : (
            <Button
              variant="ghost"
              size="icon"
              onClick={onEnable}
              title="Enable user"
            >
              <UserCheck className="h-4 w-4" />
            </Button>
          )}
          {user.is_email_verified ? (
            <Button
              variant="ghost"
              size="icon"
              onClick={onUnverifyEmail}
              title="Unverify email"
            >
              <MailX className="h-4 w-4" />
            </Button>
          ) : (
            <Button
              variant="ghost"
              size="icon"
              onClick={onVerifyEmail}
              title="Verify email"
            >
              <MailCheck className="h-4 w-4" />
            </Button>
          )}
          {user.is_system_admin ? (
            <Button
              variant="ghost"
              size="icon"
              onClick={onRevokeAdmin}
              title="Revoke admin"
            >
              <ShieldOff className="h-4 w-4" />
            </Button>
          ) : (
            <Button
              variant="ghost"
              size="icon"
              onClick={onGrantAdmin}
              title="Grant admin"
            >
              <Shield className="h-4 w-4" />
            </Button>
          )}
        </div>
      </div>
    </div>
  );
}
