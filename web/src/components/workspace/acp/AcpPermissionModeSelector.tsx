"use client";

import { useCallback, useState } from "react";
import { Shield, ChevronDown } from "lucide-react";
import { relayPool } from "@/stores/relayConnection";

const MODES = [
  { value: "bypassPermissions", label: "Bypass", desc: "Auto-approve all" },
  { value: "acceptEdits", label: "Accept Edits", desc: "Auto-approve file edits" },
  { value: "default", label: "Default", desc: "Approve each tool" },
  { value: "dontAsk", label: "Don't Ask", desc: "Deny unless allowlisted" },
] as const;

interface AcpPermissionModeSelectorProps {
  podKey: string;
  currentMode?: string;
}

export function AcpPermissionModeSelector({ podKey, currentMode = "bypassPermissions" }: AcpPermissionModeSelectorProps) {
  const [open, setOpen] = useState(false);
  const [mode, setMode] = useState(currentMode);

  const handleSelect = useCallback((value: string) => {
    if (!relayPool.isConnected(podKey)) return;
    relayPool.sendAcpCommand(podKey, { type: "set_permission_mode", mode: value });
    setMode(value);
    setOpen(false);
  }, [podKey]);

  const current = MODES.find((m) => m.value === mode) || MODES[0];

  return (
    <div className="relative">
      <button
        onClick={() => setOpen(!open)}
        className="flex items-center gap-1 px-2 py-1 text-xs rounded hover:bg-muted transition-colors"
        title={current.desc}
      >
        <Shield className="h-3 w-3 text-muted-foreground" />
        <span className="text-muted-foreground">{current.label}</span>
        <ChevronDown className="h-3 w-3 text-muted-foreground" />
      </button>
      {open && (
        <>
          <div className="fixed inset-0 z-40" onClick={() => setOpen(false)} />
          <div className="absolute right-0 bottom-full mb-1 z-50 w-48 rounded-md border bg-popover shadow-md py-1">
            {MODES.map((m) => (
              <button
                key={m.value}
                onClick={() => handleSelect(m.value)}
                className={`w-full text-left px-3 py-1.5 text-xs hover:bg-muted transition-colors ${
                  mode === m.value ? "bg-muted font-medium" : ""
                }`}
              >
                <div>{m.label}</div>
                <div className="text-muted-foreground text-[10px]">{m.desc}</div>
              </button>
            ))}
          </div>
        </>
      )}
    </div>
  );
}
