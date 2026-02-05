"use client";

import type { TuiFrame } from "./types";
import { getLineStyle } from "./useTuiAnimation";

interface TuiWindowProps {
  frame: TuiFrame;
  frameIndex: number;
  displayedLines: number;
  isTyping: boolean;
  t: (key: string) => string;
}

/**
 * TuiWindow - Renders the Claude Code TUI demonstration window
 */
export function TuiWindow({
  frame,
  frameIndex,
  displayedLines,
  isTyping,
  t,
}: TuiWindowProps) {
  return (
    <div className="relative bg-card rounded-xl border border-border overflow-hidden shadow-2xl">
      {/* TUI Header */}
      <div className="flex items-center justify-between px-4 py-2 bg-muted border-b border-border">
        <div className="flex items-center gap-2">
          <div className="flex gap-2">
            <div className="w-3 h-3 rounded-full bg-red-500/80" />
            <div className="w-3 h-3 rounded-full bg-yellow-500/80" />
            <div className="w-3 h-3 rounded-full bg-green-500/80" />
          </div>
          <span className="text-sm font-semibold text-foreground ml-2">{frame.content.header}</span>
        </div>
        <div className="flex items-center gap-2">
          <span className="text-xs px-2 py-0.5 bg-primary/20 text-primary rounded">MCP</span>
          <span className="text-xs px-2 py-0.5 bg-green-500/20 text-green-600 dark:text-green-400 rounded">AgentsMesh</span>
        </div>
      </div>

      {/* Pod Info Bar */}
      <div className="px-4 py-1.5 bg-muted/70 border-b border-border text-xs font-mono text-muted-foreground">
        {t(`landing.heroDemo.podInfo.${frame.content.podInfoKey}`)}
      </div>

      {/* Main Content Area */}
      <div className="p-4 font-mono text-sm h-[260px] overflow-hidden">
        {frame.content.mainContent.slice(0, displayedLines).map((line, index) => {
          const lineText = line.textKey ? t(line.textKey) : line.text;
          return (
            <div
              key={`${frameIndex}-${index}`}
              className={`${getLineStyle(line.type)} ${lineText ? '' : 'h-4'}`}
            >
              {line.type === "user" && <span className="text-blue-500 dark:text-blue-400 mr-2">❯</span>}
              {line.type === "assistant" && <span className="text-primary mr-2">●</span>}
              {lineText}
            </div>
          );
        })}
        {isTyping && (
          <span className="animate-pulse text-primary">▋</span>
        )}
      </div>

      {/* Input Area */}
      <div className="px-4 py-2 bg-muted/70 border-t border-border">
        <div className="flex items-center gap-2 text-sm font-mono">
          <span className="text-primary">❯</span>
          <span className="text-muted-foreground">{t("landing.heroDemo.typeMessage")}</span>
          <span className="animate-pulse text-primary">▋</span>
        </div>
      </div>
    </div>
  );
}

export default TuiWindow;
