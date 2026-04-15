/**
 * Retry and reconnect logic for relay WebSocket connections.
 */

import { ApiError } from "@/lib/api/base";

const NON_RETRYABLE_STATUSES = [400, 403, 404];

export function isNonRetryableError(error: unknown): boolean {
  return error instanceof ApiError && NON_RETRYABLE_STATUSES.includes(error.status);
}

/**
 * Schedule a reconnect with exponential backoff + jitter.
 */
export function scheduleReconnect(
  getConnection: (podKey: string) => import("./relayConnectionTypes").RelayConnection | undefined,
  reconnectFn: (podKey: string) => void,
  podKey: string,
  maxAttempts: number,
  baseDelay: number,
): void {
  const conn = getConnection(podKey);
  if (!conn || conn.reconnectAttempts >= maxAttempts) return;

  const delay = computeReconnectDelay(conn.reconnectAttempts, baseDelay);

  conn.reconnectTimer = setTimeout(() => {
    conn.reconnectAttempts++;
    reconnectFn(podKey);
  }, delay);
}

function computeReconnectDelay(attempt: number, baseDelay: number): number {
  const base = Math.min(baseDelay * Math.pow(2, attempt), 30000);
  const jitter = base * (Math.random() * 0.4 - 0.2);
  return Math.round(base + jitter);
}
