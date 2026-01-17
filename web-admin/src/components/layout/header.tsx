"use client";

import { usePathname } from "next/navigation";
import { Bell } from "lucide-react";
import { Button } from "@/components/ui/button";

const pageTitles: Record<string, string> = {
  "/": "Dashboard",
  "/users": "Users",
  "/organizations": "Organizations",
  "/runners": "Runners",
  "/audit-logs": "Audit Logs",
};

export function Header() {
  const pathname = usePathname();

  // Get title - handle dynamic routes
  let title = pageTitles[pathname];
  if (!title) {
    // Check for detail pages
    if (pathname.startsWith("/users/")) title = "User Details";
    else if (pathname.startsWith("/organizations/")) title = "Organization Details";
    else if (pathname.startsWith("/runners/")) title = "Runner Details";
    else title = "Admin Console";
  }

  return (
    <header className="flex h-16 items-center justify-between border-b border-border bg-card px-6">
      <h1 className="text-xl font-semibold">{title}</h1>
      <div className="flex items-center gap-2">
        <Button variant="ghost" size="icon">
          <Bell className="h-5 w-5" />
        </Button>
      </div>
    </header>
  );
}
