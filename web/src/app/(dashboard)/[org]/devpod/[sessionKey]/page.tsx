"use client";

import { useParams } from "next/navigation";
import { SessionDetail } from "@/components/devpod";

export default function SessionDetailPage() {
  const params = useParams();
  const sessionKey = params.sessionKey as string;

  return (
    <div className="h-full">
      <SessionDetail sessionKey={sessionKey} />
    </div>
  );
}
