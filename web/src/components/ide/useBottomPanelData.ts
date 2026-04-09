"use client";

import { useMemo, useCallback } from "react";
import { useWorkspaceStore } from "@/stores/workspace";
import { useMeshStore, type MeshEdge } from "@/stores/mesh";
import { usePodStore } from "@/stores/pod";
import { useAutopilotStore } from "@/stores/autopilot";

/** Derived pod-centric data for BottomPanel tabs */
export function useBottomPanelData() {
  const panes = useWorkspaceStore((s) => s.panes);
  const activePane = useWorkspaceStore((s) => s.activePane);
  const topology = useMeshStore((s) => s.topology);
  const fetchTopology = useMeshStore((s) => s.fetchTopology);

  const selectedPodKey = useMemo(() => {
    if (!activePane) return null;
    const pane = panes.find((p) => p.id === activePane);
    return pane?.podKey ?? null;
  }, [activePane, panes]);

  const currentPod = usePodStore((s) =>
    selectedPodKey ? s.pods.find((p) => p.pod_key === selectedPodKey) ?? null : null
  );

  const activePhases = ["initializing", "running", "paused", "user_takeover", "waiting_approval"];
  const activeAutopilot = useAutopilotStore((s) =>
    selectedPodKey
      ? s.autopilotControllers.find((c) => c.pod_key === selectedPodKey && activePhases.includes(c.phase))
      : undefined
  );

  const podChannels = useMemo(() => {
    if (!selectedPodKey || !topology) return [];
    return topology.channels.filter((c) => c.pod_keys.includes(selectedPodKey));
  }, [selectedPodKey, topology]);

  const podEdges = useMemo(() => {
    if (!selectedPodKey || !topology) return [];
    return topology.edges.filter((e) => e.source === selectedPodKey || e.target === selectedPodKey);
  }, [selectedPodKey, topology]);

  const { incomingBindings, outgoingBindings } = useMemo(() => {
    const incoming: MeshEdge[] = [];
    const outgoing: MeshEdge[] = [];
    podEdges.forEach((edge) => {
      if (edge.target === selectedPodKey) incoming.push(edge);
      else if (edge.source === selectedPodKey) outgoing.push(edge);
    });
    return { incomingBindings: incoming, outgoingBindings: outgoing };
  }, [podEdges, selectedPodKey]);

  const getPodInfo = useCallback(
    (podKey: string) => topology?.nodes.find((n) => n.pod_key === podKey),
    [topology]
  );

  return {
    selectedPodKey, currentPod, activeAutopilot,
    topology, fetchTopology,
    podChannels, incomingBindings, outgoingBindings, getPodInfo,
  };
}
