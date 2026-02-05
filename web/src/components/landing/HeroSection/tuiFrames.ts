import type { TuiFrame } from "./types";

/**
 * Generate TUI demonstration frames with translated text
 */
export function getTuiFrames(t: (key: string) => string): TuiFrame[] {
  return [
    {
      time: 0,
      content: {
        header: "Claude Code",
        podInfoKey: "initializing",
        mainContent: [
          { type: "system", textKey: "landing.heroDemo.connecting" },
        ],
        input: "",
        topology: { nodes: [], connections: [] }
      }
    },
    {
      time: 1500,
      content: {
        header: "Claude Code",
        podInfoKey: "running",
        mainContent: [
          { type: "system", text: "✓ " + t("landing.heroDemo.connected") },
          { type: "system", text: "✓ " + t("landing.heroDemo.workspace") },
          { type: "system", text: "" },
          { type: "user", textKey: "landing.heroDemo.monitorPod" },
        ],
        input: "",
        topology: {
          nodes: [{ id: "alpha", label: "alpha-dev", agent: "Claude Code", status: "running", x: 25, y: 40 }],
          connections: []
        }
      }
    },
    {
      time: 3500,
      content: {
        header: "Claude Code",
        podInfoKey: "running",
        mainContent: [
          { type: "system", text: "✓ " + t("landing.heroDemo.connected") },
          { type: "system", text: "✓ " + t("landing.heroDemo.workspace") },
          { type: "system", text: "" },
          { type: "user", textKey: "landing.heroDemo.monitorPod" },
          { type: "system", text: "" },
          { type: "assistant", textKey: "landing.heroDemo.bindingPod" },
          { type: "tool", text: "⚡ mesh_bind_pod beta-dev --scopes observe,message" },
        ],
        input: "",
        topology: {
          nodes: [
            { id: "alpha", label: "alpha-dev", agent: "Claude Code", status: "running", x: 25, y: 40 },
            { id: "beta", label: "beta-dev", agent: "Codex CLI", status: "running", x: 75, y: 40 },
          ],
          connections: []
        }
      }
    },
    {
      time: 5500,
      content: {
        header: "Claude Code",
        podInfoKey: "observing",
        mainContent: [
          { type: "user", textKey: "landing.heroDemo.monitorPod" },
          { type: "system", text: "" },
          { type: "assistant", textKey: "landing.heroDemo.bindingPod" },
          { type: "tool", text: "⚡ mesh_bind_pod beta-dev --scopes observe,message" },
          { type: "success", text: "✓ " + t("landing.heroDemo.boundToPod") },
          { type: "system", text: "" },
          { type: "observe-header", text: "━━━ " + t("landing.heroDemo.observingHeader") + " ━━━" },
          { type: "observe", text: "│ " + t("landing.heroDemo.analyzing") },
        ],
        input: "",
        topology: {
          nodes: [
            { id: "alpha", label: "alpha-dev", agent: "Claude Code", status: "running", x: 25, y: 40 },
            { id: "beta", label: "beta-dev", agent: "Codex CLI", status: "running", x: 75, y: 40 },
          ],
          connections: [
            { from: "alpha", to: "beta", label: "observing", animated: true }
          ]
        }
      }
    },
    {
      time: 7500,
      content: {
        header: "Claude Code",
        podInfoKey: "observing",
        mainContent: [
          { type: "tool", text: "⚡ mesh_bind_pod beta-dev --scopes observe,message" },
          { type: "success", text: "✓ " + t("landing.heroDemo.boundToPod") },
          { type: "system", text: "" },
          { type: "observe-header", text: "━━━ " + t("landing.heroDemo.observingHeader") + " ━━━" },
          { type: "observe", text: "│ " + t("landing.heroDemo.analyzing") },
          { type: "observe", text: "│ " + t("landing.heroDemo.creatingOAuth") },
          { type: "observe", text: "│ " + t("landing.heroDemo.writingOAuth") },
          { type: "system", text: "" },
          { type: "assistant", textKey: "landing.heroDemo.noticedOAuth" },
        ],
        input: "",
        topology: {
          nodes: [
            { id: "alpha", label: "alpha-dev", agent: "Claude Code", status: "running", x: 25, y: 40 },
            { id: "beta", label: "beta-dev", agent: "Codex CLI", status: "running", x: 75, y: 40 },
          ],
          connections: [
            { from: "alpha", to: "beta", label: "observing", animated: true }
          ]
        }
      }
    },
    {
      time: 9500,
      content: {
        header: "Claude Code",
        podInfoKey: "observing",
        mainContent: [
          { type: "observe-header", text: "━━━ " + t("landing.heroDemo.observingHeader") + " ━━━" },
          { type: "observe", text: "│ " + t("landing.heroDemo.creatingOAuth") },
          { type: "observe", text: "│ " + t("landing.heroDemo.writingOAuth") },
          { type: "system", text: "" },
          { type: "assistant", textKey: "landing.heroDemo.noticedOAuth" },
          { type: "tool", text: "⚡ mesh_send_message beta-dev" },
          { type: "message-sent", text: "📤 " + t("landing.heroDemo.messageSent") },
          { type: "system", text: "" },
          { type: "observe", text: "│ " + t("landing.heroDemo.receivedSuggestion") },
        ],
        input: "",
        topology: {
          nodes: [
            { id: "alpha", label: "alpha-dev", agent: "Claude Code", status: "running", x: 25, y: 40 },
            { id: "beta", label: "beta-dev", agent: "Codex CLI", status: "running", x: 75, y: 40 },
          ],
          connections: [
            { from: "alpha", to: "beta", label: "observing", animated: true },
            { from: "alpha", to: "beta", label: "message", type: "message", animated: true }
          ]
        }
      }
    },
    {
      time: 12000,
      content: {
        header: "Claude Code",
        podInfoKey: "observing",
        mainContent: [
          { type: "tool", text: "⚡ mesh_send_message beta-dev" },
          { type: "message-sent", text: "📤 " + t("landing.heroDemo.messageSent") },
          { type: "system", text: "" },
          { type: "observe", text: "│ " + t("landing.heroDemo.receivedSuggestion") },
          { type: "observe", text: "│ " + t("landing.heroDemo.writingRateLimit") },
          { type: "observe", text: "│ ✓ " + t("landing.heroDemo.rateLimitAdded") },
          { type: "system", text: "" },
          { type: "assistant", textKey: "landing.heroDemo.reviewChanges" },
          { type: "tool", text: "⚡ mesh_read_file beta-dev:src/middleware/rateLimit.ts" },
        ],
        input: "",
        topology: {
          nodes: [
            { id: "alpha", label: "alpha-dev", agent: "Claude Code", status: "running", x: 25, y: 40 },
            { id: "beta", label: "beta-dev", agent: "Codex CLI", status: "running", x: 75, y: 40 },
          ],
          connections: [
            { from: "alpha", to: "beta", label: "collaborating", animated: true }
          ]
        }
      }
    },
  ];
}
