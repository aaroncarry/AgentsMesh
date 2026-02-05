"use client";

import type { Topology } from "./types";

interface TopologyVisualizationProps {
  topology: Topology;
  t: (key: string) => string;
}

/**
 * TopologyVisualization - Renders the Pod mesh topology visualization
 */
export function TopologyVisualization({ topology, t }: TopologyVisualizationProps) {
  return (
    <div className="relative bg-card rounded-xl border border-border overflow-hidden">
      <div className="px-4 py-2 bg-muted border-b border-border flex items-center justify-between">
        <span className="text-xs font-mono text-muted-foreground">{t("landing.heroDemo.podTopology")}</span>
        <div className="flex items-center gap-2">
          <span className="text-xs text-green-500 dark:text-green-400">● {t("landing.heroDemo.live")}</span>
        </div>
      </div>

      <div className="relative h-[140px] p-4">
        {/* Connection lines */}
        <svg className="absolute inset-0 w-full h-full pointer-events-none" style={{ zIndex: 1 }}>
          <defs>
            <marker id="arrowhead" markerWidth="10" markerHeight="7" refX="9" refY="3.5" orient="auto">
              <polygon points="0 0, 10 3.5, 0 7" className="fill-primary" opacity="0.8" />
            </marker>
          </defs>
          {topology.connections.map((conn, i) => {
            const fromNode = topology.nodes.find(n => n.id === conn.from);
            const toNode = topology.nodes.find(n => n.id === conn.to);
            if (!fromNode || !toNode) return null;

            const midX = (fromNode.x + toNode.x) / 2;
            const midY = (fromNode.y + toNode.y) / 2;

            return (
              <g key={i}>
                <line
                  x1={`${fromNode.x + 8}%`}
                  y1={`${fromNode.y}%`}
                  x2={`${toNode.x - 8}%`}
                  y2={`${toNode.y}%`}
                  className={`stroke-primary ${conn.animated ? "animate-pulse" : ""}`}
                  strokeWidth="2"
                  strokeDasharray={conn.type === "message" ? "5,5" : "0"}
                  markerEnd="url(#arrowhead)"
                  opacity="0.6"
                />
                <text
                  x={`${midX}%`}
                  y={`${midY - 8}%`}
                  className="fill-muted-foreground font-mono"
                  fontSize="10"
                  textAnchor="middle"
                >
                  {conn.label}
                </text>
              </g>
            );
          })}
        </svg>

        {/* Pod nodes */}
        {topology.nodes.map((node) => (
          <div
            key={node.id}
            className="absolute transform -translate-x-1/2 -translate-y-1/2 transition-all duration-500 ease-out"
            style={{ left: `${node.x}%`, top: `${node.y}%`, zIndex: 2 }}
          >
            <div className="px-3 py-2 bg-muted border border-primary/40 rounded-lg shadow-lg shadow-primary/10 min-w-[110px]">
              <div className="flex items-center gap-2 mb-1">
                <div className={`w-2 h-2 rounded-full ${node.status === "running" ? "bg-green-500 animate-pulse" : "bg-gray-500"}`} />
                <span className="text-xs font-mono text-primary font-semibold">{node.label}</span>
              </div>
              <div className="text-[10px] text-muted-foreground">{node.agent}</div>
            </div>
          </div>
        ))}

        {/* Empty state */}
        {topology.nodes.length === 0 && (
          <div className="flex items-center justify-center h-full text-muted-foreground text-sm">
            <span className="animate-pulse">{t("landing.heroDemo.initializingPod")}</span>
          </div>
        )}
      </div>
    </div>
  );
}

export default TopologyVisualization;
