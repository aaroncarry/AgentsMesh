"use client";

import { useEffect, useRef, useMemo, useState } from "react";
import { usePodStore } from "@/stores/pod";
import { ApiError } from "@/lib/api";

interface UsePodStatusResult {
  podStatus: string;
  isPodReady: boolean;
  podError: string | null;
}

/**
 * Hook for tracking pod readiness status
 * Uses realtime events via store - only fetches once on mount for initial state
 */
export function usePodStatus(podKey: string): UsePodStatusResult {
  const initialFetchDone = useRef(false);
  const [fetchError, setFetchError] = useState<string | null>(null);
  const { pods, fetchPod } = usePodStore();

  // Get pod from store (updated via realtime events)
  const storePod = pods.find((p) => p.pod_key === podKey);

  // Derive status from store - no local state needed
  const { podStatus, isPodReady, podError } = useMemo(() => {
    // If fetch failed (e.g. 404), report the error immediately
    if (fetchError) {
      return { podStatus: "error", isPodReady: false, podError: fetchError };
    }

    const status = storePod?.status ?? "unknown";
    const isReady = status === "running";

    // Terminal states that indicate pod cannot be used
    let error: string | null = null;
    if (status === "failed") {
      error = "Pod failed";
    } else if (status === "terminated") {
      error = "Pod terminated";
    } else if (status === "orphaned") {
      error = "Pod orphaned - Runner connection lost";
    } else if (status === "error") {
      error = storePod?.error_message || "Pod error";
    }

    return { podStatus: status, isPodReady: isReady, podError: error };
  }, [storePod?.status, storePod?.error_message, fetchError]);

  // Initial status fetch (once only) - updates store via fetchPod
  useEffect(() => {
    if (initialFetchDone.current || storePod) return;
    initialFetchDone.current = true;

    fetchPod(podKey).catch((error) => {
      if (error instanceof ApiError && error.status === 404) {
        setFetchError("Pod not found");
      } else {
        setFetchError("Failed to load pod");
      }
    });
  }, [podKey, fetchPod, storePod]);

  return { podStatus, isPodReady, podError };
}
