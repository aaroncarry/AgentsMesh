"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";

// Redirect to general settings by default
export default function PersonalSettingsPage() {
  const router = useRouter();

  useEffect(() => {
    router.replace("/settings/general");
  }, [router]);

  return (
    <div className="flex items-center justify-center h-full">
      <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
    </div>
  );
}
