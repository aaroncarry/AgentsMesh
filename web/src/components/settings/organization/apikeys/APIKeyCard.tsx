"use client";

import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Key, Pencil, Ban } from "lucide-react";
import type { APIKeyData } from "@/lib/api/apikey";
import type { TranslationFn } from "../GeneralSettings";

interface APIKeyCardProps {
  apiKey: APIKeyData;
  onEdit: (apiKey: APIKeyData) => void;
  onRevoke: (id: number) => void;
  t: TranslationFn;
}

function formatRelativeTime(dateString?: string, t?: TranslationFn): string {
  if (!dateString) return t?.("settings.apiKeys.neverUsed") || "Never used";
  const date = new Date(dateString);
  const now = new Date();
  const diffMs = now.getTime() - date.getTime();
  const diffSec = Math.floor(diffMs / 1000);

  if (diffSec < 60) return t?.("settings.apiKeys.justNow") || "just now";
  if (diffSec < 3600) {
    const count = Math.floor(diffSec / 60);
    return t?.("settings.apiKeys.minutesAgo", { count }) || `${count}m ago`;
  }
  if (diffSec < 86400) {
    const count = Math.floor(diffSec / 3600);
    return t?.("settings.apiKeys.hoursAgo", { count }) || `${count}h ago`;
  }
  const count = Math.floor(diffSec / 86400);
  return t?.("settings.apiKeys.daysAgo", { count }) || `${count}d ago`;
}

export function APIKeyCard({ apiKey, onEdit, onRevoke, t }: APIKeyCardProps) {
  const isExpired = apiKey.expires_at && new Date(apiKey.expires_at) < new Date();
  const isActive = apiKey.is_enabled && !isExpired;

  return (
    <div className="flex items-center justify-between p-4 border border-border rounded-lg">
      <div className="flex items-start gap-3 min-w-0 flex-1">
        <div className="w-9 h-9 rounded-lg bg-muted flex items-center justify-center flex-shrink-0 mt-0.5">
          <Key className="w-4 h-4 text-muted-foreground" />
        </div>
        <div className="min-w-0 flex-1">
          <div className="flex items-center gap-2 mb-1">
            <span className="font-medium truncate">{apiKey.name}</span>
            <code className="text-xs text-muted-foreground bg-muted px-1.5 py-0.5 rounded">
              {apiKey.key_prefix}...
            </code>
            {isActive ? (
              <Badge variant="default" className="text-xs">
                {t("settings.apiKeys.enabled")}
              </Badge>
            ) : (
              <Badge variant="secondary" className="text-xs">
                {apiKey.is_enabled
                  ? t("settings.apiKeys.expired")
                  : t("settings.apiKeys.disabled")}
              </Badge>
            )}
          </div>

          {/* Scopes */}
          <div className="flex flex-wrap gap-1 mb-1.5">
            {apiKey.scopes.map((scope) => (
              <span
                key={scope}
                className="text-xs bg-muted text-muted-foreground px-1.5 py-0.5 rounded"
              >
                {scope}
              </span>
            ))}
          </div>

          {/* Metadata */}
          <div className="flex items-center gap-3 text-xs text-muted-foreground">
            <span>
              {t("settings.apiKeys.lastUsed", {
                time: formatRelativeTime(apiKey.last_used_at, t),
              })}
            </span>
            <span>·</span>
            <span>
              {apiKey.expires_at
                ? t("settings.apiKeys.expiresAt", {
                    date: new Date(apiKey.expires_at).toLocaleDateString(),
                  })
                : t("settings.apiKeys.neverExpires")}
            </span>
          </div>
        </div>
      </div>

      <div className="flex items-center gap-1 ml-2 flex-shrink-0">
        <Button
          variant="ghost"
          size="sm"
          onClick={() => onEdit(apiKey)}
          aria-label={t("settings.apiKeys.editDialog.title")}
        >
          <Pencil className="w-4 h-4" />
        </Button>
        {apiKey.is_enabled && (
          <Button
            variant="ghost"
            size="sm"
            className="text-destructive hover:text-destructive"
            onClick={() => onRevoke(apiKey.id)}
            aria-label={t("settings.apiKeys.revokeDialog.title")}
          >
            <Ban className="w-4 h-4" />
          </Button>
        )}
      </div>
    </div>
  );
}
