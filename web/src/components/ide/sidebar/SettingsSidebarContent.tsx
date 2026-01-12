"use client";

import React, { useState } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { cn } from "@/lib/utils";
import { useAuthStore } from "@/stores/auth";
import { useTranslations } from "@/lib/i18n/client";
import {
  Settings,
  Users,
  Bot,
  Server,
  CreditCard,
  User,
  GitBranch,
  Bell,
  Building2,
  ChevronDown,
  ChevronRight,
} from "lucide-react";

interface SettingsSidebarContentProps {
  className?: string;
}

type SettingsScope = "organization" | "personal";

export function SettingsSidebarContent({ className }: SettingsSidebarContentProps) {
  const router = useRouter();
  const searchParams = useSearchParams();
  const { currentOrg } = useAuthStore();
  const t = useTranslations();

  // Get current scope and tab from URL params - default to personal scope
  const currentScope: SettingsScope = (searchParams.get("scope") as SettingsScope) || "personal";
  const currentTab = searchParams.get("tab") || "general";

  // Track expanded sections - expand personal by default
  const [expandedSections, setExpandedSections] = useState<Record<SettingsScope, boolean>>({
    personal: true,
    organization: currentScope === "organization",
  });

  // Organization settings tabs
  const orgSettingsTabs = [
    { id: "general", labelKey: "ide.sidebar.settings.tabs.general", icon: Settings },
    { id: "members", labelKey: "ide.sidebar.settings.tabs.members", icon: Users },
    { id: "agents", labelKey: "ide.sidebar.settings.tabs.agents", icon: Bot },
    { id: "runners", labelKey: "ide.sidebar.settings.tabs.runners", icon: Server },
    { id: "billing", labelKey: "ide.sidebar.settings.tabs.billing", icon: CreditCard },
  ];

  // Personal settings tabs
  const personalSettingsTabs = [
    { id: "general", labelKey: "ide.sidebar.settings.tabs.general", icon: Settings },
    { id: "git", labelKey: "settings.personal.tabs.git", icon: GitBranch },
    { id: "notifications", labelKey: "settings.personal.tabs.notifications", icon: Bell },
  ];

  // Toggle section expansion
  const toggleSection = (scope: SettingsScope) => {
    setExpandedSections(prev => ({
      ...prev,
      [scope]: !prev[scope],
    }));
  };

  // Handle tab click
  const handleTabClick = (scope: SettingsScope, tabId: string) => {
    if (scope === "organization") {
      router.push(`/${currentOrg?.slug}/settings?scope=organization&tab=${tabId}`);
    } else {
      router.push(`/${currentOrg?.slug}/settings?scope=personal&tab=${tabId}`);
    }
    // Auto expand the section when clicking a tab
    setExpandedSections(prev => ({
      ...prev,
      [scope]: true,
    }));
  };

  // Render a collapsible section
  const renderSection = (
    scope: SettingsScope,
    titleKey: string,
    Icon: typeof Building2,
    tabs: typeof orgSettingsTabs
  ) => {
    const isExpanded = expandedSections[scope];
    const isCurrentScope = currentScope === scope;

    return (
      <div className="mb-1">
        {/* Section header */}
        <button
          className={cn(
            "w-full flex items-center gap-2 px-3 py-2 text-left transition-colors",
            "hover:bg-muted/50",
            isCurrentScope && "text-foreground"
          )}
          onClick={() => toggleSection(scope)}
        >
          {isExpanded ? (
            <ChevronDown className="w-4 h-4 text-muted-foreground" />
          ) : (
            <ChevronRight className="w-4 h-4 text-muted-foreground" />
          )}
          <Icon className="w-4 h-4" />
          <span className="text-sm font-medium">{t(titleKey)}</span>
        </button>

        {/* Section items */}
        {isExpanded && (
          <div className="ml-4 border-l border-border">
            {tabs.map((tab) => {
              const TabIcon = tab.icon;
              const isActive = isCurrentScope && currentTab === tab.id;

              return (
                <button
                  key={tab.id}
                  className={cn(
                    "w-full flex items-center gap-2 pl-4 pr-3 py-1.5 text-left transition-colors",
                    isActive
                      ? "bg-muted text-foreground"
                      : "text-muted-foreground hover:bg-muted/50 hover:text-foreground"
                  )}
                  onClick={() => handleTabClick(scope, tab.id)}
                >
                  <TabIcon className={cn(
                    "w-4 h-4 flex-shrink-0",
                    isActive && "text-primary"
                  )} />
                  <span className={cn(
                    "text-sm truncate",
                    isActive && "font-medium"
                  )}>
                    {t(tab.labelKey)}
                  </span>
                </button>
              );
            })}
          </div>
        )}
      </div>
    );
  };

  return (
    <div className={cn("flex flex-col h-full", className)}>
      {/* Settings navigation */}
      <div className="flex-1 overflow-y-auto py-2">
        {/* Personal settings section (on top) */}
        {renderSection(
          "personal",
          "ide.sidebar.settings.scopePersonal",
          User,
          personalSettingsTabs
        )}

        {/* Organization settings section */}
        {renderSection(
          "organization",
          "ide.sidebar.settings.scopeOrg",
          Building2,
          orgSettingsTabs
        )}
      </div>

      {/* Organization info at bottom */}
      {currentOrg && (
        <div className="border-t border-border px-3 py-3">
          <div className="text-xs text-muted-foreground mb-1">{t("ide.sidebar.settings.currentOrg")}</div>
          <div className="text-sm font-medium truncate">{currentOrg.name}</div>
          <div className="text-xs text-muted-foreground truncate">/{currentOrg.slug}</div>
        </div>
      )}
    </div>
  );
}

export default SettingsSidebarContent;
