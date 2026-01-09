"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { useAuthStore } from "@/stores/auth";
import { organizationApi } from "@/lib/api/organization";

export default function OnboardingPage() {
  const router = useRouter();
  const { user, setOrganizations, setCurrentOrg } = useAuthStore();
  const [inviteCode, setInviteCode] = useState("");
  const [showInviteInput, setShowInviteInput] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  // Check if user already has organizations
  useEffect(() => {
    const checkOrgs = async () => {
      try {
        const { organizations } = await organizationApi.list();
        if (organizations && organizations.length > 0) {
          setOrganizations(organizations);
          setCurrentOrg(organizations[0]);
          router.push(`/${organizations[0].slug}`);
        }
      } catch {
        // No organizations, stay on onboarding
      }
    };
    checkOrgs();
  }, [router, setOrganizations, setCurrentOrg]);

  const handleQuickStart = async () => {
    if (!user) return;

    setLoading(true);
    setError("");

    try {
      // Create personal workspace with username as slug
      const slug = `${user.username}-workspace`;
      const name = `${user.name || user.username}'s Workspace`;

      await organizationApi.create({ name, slug });

      // Refresh organizations
      const { organizations } = await organizationApi.list();
      setOrganizations(organizations);

      const newOrg = organizations.find((o) => o.slug === slug);
      if (newOrg) {
        setCurrentOrg(newOrg);
      }

      // Go to runner setup
      router.push("/onboarding/setup-runner");
    } catch {
      setError("Failed to create workspace. The name might already be taken.");
    } finally {
      setLoading(false);
    }
  };

  const handleJoinWithInvite = async () => {
    if (!inviteCode.trim()) {
      setError("Please enter an invite code");
      return;
    }

    setLoading(true);
    setError("");

    try {
      // TODO: Implement invite code API
      // await inviteApi.accept(inviteCode);

      // For now, show not implemented message
      setError("Invite code feature is coming soon");
    } catch {
      setError("Invalid or expired invite code");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="flex min-h-screen items-center justify-center bg-background px-4">
      <div className="w-full max-w-md space-y-8">
        {/* Header */}
        <div className="text-center">
          <Link href="/" className="inline-flex items-center gap-2">
            <div className="w-10 h-10 rounded-lg bg-primary flex items-center justify-center">
              <svg
                className="w-6 h-6 text-primary-foreground"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M9 3v2m6-2v2M9 19v2m6-2v2M5 9H3m2 6H3m18-6h-2m2 6h-2M7 19h10a2 2 0 002-2V7a2 2 0 00-2-2H7a2 2 0 00-2 2v10a2 2 0 002 2zM9 9h6v6H9V9z"
                />
              </svg>
            </div>
            <span className="text-2xl font-bold text-foreground">AgentMesh</span>
          </Link>
          <h1 className="mt-6 text-2xl font-semibold text-foreground">
            Welcome{user?.name ? `, ${user.name}` : ""}!
          </h1>
          <p className="mt-2 text-sm text-muted-foreground">
            Let&apos;s set up your workspace to get started.
          </p>
        </div>

        {/* Error Message */}
        {error && (
          <div className="p-3 text-sm text-destructive bg-destructive/10 rounded-md">
            {error}
          </div>
        )}

        {/* Options */}
        <div className="space-y-4">
          {/* Quick Start */}
          <div className="p-6 border border-border rounded-lg hover:border-primary/50 transition-colors">
            <div className="flex items-start gap-4">
              <div className="w-12 h-12 rounded-lg bg-primary/10 flex items-center justify-center flex-shrink-0">
                <svg
                  className="w-6 h-6 text-primary"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M13 10V3L4 14h7v7l9-11h-7z"
                  />
                </svg>
              </div>
              <div className="flex-1">
                <h3 className="font-semibold text-foreground">Quick Start</h3>
                <p className="mt-1 text-sm text-muted-foreground">
                  Create a personal workspace and start using AgentMesh immediately.
                </p>
                <Button
                  className="mt-4 w-full"
                  onClick={handleQuickStart}
                  disabled={loading}
                >
                  {loading ? "Creating..." : "Create Personal Workspace"}
                </Button>
              </div>
            </div>
          </div>

          {/* Create Team */}
          <div className="p-6 border border-border rounded-lg hover:border-primary/50 transition-colors">
            <div className="flex items-start gap-4">
              <div className="w-12 h-12 rounded-lg bg-blue-500/10 flex items-center justify-center flex-shrink-0">
                <svg
                  className="w-6 h-6 text-blue-500"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z"
                  />
                </svg>
              </div>
              <div className="flex-1">
                <h3 className="font-semibold text-foreground">Create Team Workspace</h3>
                <p className="mt-1 text-sm text-muted-foreground">
                  Set up a workspace for your team with a custom name and invite members.
                </p>
                <Link href="/onboarding/create-org">
                  <Button variant="outline" className="mt-4 w-full">
                    Create Team Workspace
                  </Button>
                </Link>
              </div>
            </div>
          </div>
        </div>

        {/* Divider */}
        <div className="relative">
          <div className="absolute inset-0 flex items-center">
            <span className="w-full border-t border-border" />
          </div>
          <div className="relative flex justify-center text-xs uppercase">
            <span className="bg-background px-2 text-muted-foreground">
              Or join an existing workspace
            </span>
          </div>
        </div>

        {/* Join with Invite */}
        {showInviteInput ? (
          <div className="space-y-3">
            <Input
              placeholder="Enter invite code"
              value={inviteCode}
              onChange={(e) => setInviteCode(e.target.value)}
            />
            <div className="flex gap-2">
              <Button
                variant="outline"
                className="flex-1"
                onClick={() => {
                  setShowInviteInput(false);
                  setInviteCode("");
                  setError("");
                }}
              >
                Cancel
              </Button>
              <Button
                className="flex-1"
                onClick={handleJoinWithInvite}
                disabled={loading || !inviteCode.trim()}
              >
                {loading ? "Joining..." : "Join"}
              </Button>
            </div>
          </div>
        ) : (
          <Button
            variant="ghost"
            className="w-full"
            onClick={() => setShowInviteInput(true)}
          >
            Have an invite code?
          </Button>
        )}
      </div>
    </div>
  );
}
