/**
 * Type definitions for Hero Section TUI demonstration
 */

export interface TuiLine {
  type: string;
  textKey?: string;
  text?: string;
}

export interface TopologyNode {
  id: string;
  label: string;
  agent: string;
  status: string;
  x: number;
  y: number;
}

export interface TopologyConnection {
  from: string;
  to: string;
  label: string;
  type?: string;
  animated?: boolean;
}

export interface Topology {
  nodes: TopologyNode[];
  connections: TopologyConnection[];
}

export interface TuiFrameContent {
  header: string;
  podInfoKey: "initializing" | "running" | "observing";
  mainContent: TuiLine[];
  input: string;
  topology: Topology;
}

export interface TuiFrame {
  time: number;
  content: TuiFrameContent;
}
