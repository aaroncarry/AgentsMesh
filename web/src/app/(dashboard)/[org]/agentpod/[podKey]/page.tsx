"use client";

import { useParams } from "next/navigation";
import { PodDetail } from "@/components/agentpod";

export default function PodDetailPage() {
  const params = useParams();
  const podKey = params.podKey as string;

  return (
    <div className="h-full">
      <PodDetail podKey={podKey} />
    </div>
  );
}
