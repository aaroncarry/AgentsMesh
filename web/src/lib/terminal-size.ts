/**
 * Terminal size estimation utilities
 *
 * Used to estimate terminal dimensions (cols/rows) before xterm.js is initialized.
 * This helps provide better initial PTY dimensions when creating pods.
 *
 * Note: These are estimates only. The actual terminal size will be determined
 * by xterm.js FitAddon after the terminal is mounted and will be sent to PTY
 * immediately via forceResize.
 */

// =============================================================================
// Font Configuration (should match useTerminal.ts)
// =============================================================================

/** Default terminal font size in pixels */
const DEFAULT_FONT_SIZE = 14;

/** Default line height multiplier */
const DEFAULT_LINE_HEIGHT = 1.2;

/**
 * Monospace font width ratio relative to font size.
 * Most monospace fonts (Menlo, Monaco, Consolas) have a width of ~0.6x their height.
 * This is an approximation - actual value depends on the specific font.
 */
const MONOSPACE_WIDTH_RATIO = 0.6;

// =============================================================================
// Terminal Dimension Bounds
// =============================================================================

/** Minimum columns - ensures terminal is usable */
const MIN_COLS = 20;

/** Minimum rows - ensures terminal is usable */
const MIN_ROWS = 5;

/** Maximum columns - reasonable upper bound to prevent extreme values */
const MAX_COLS = 300;

/** Maximum rows - reasonable upper bound to prevent extreme values */
const MAX_ROWS = 100;

/** Conservative fallback dimensions (standard VT100) */
const FALLBACK_COLS = 80;
const FALLBACK_ROWS = 24;

// =============================================================================
// Layout Configuration
// =============================================================================

/** Mobile breakpoint in pixels (matches useBreakpoint.ts) */
const MOBILE_BREAKPOINT = 768;

// Desktop layout dimensions (matches IDEShell.tsx)
/** Activity bar width in pixels */
const DESKTOP_ACTIVITY_BAR_WIDTH = 48;

/** Default sidebar width in pixels when expanded */
const DESKTOP_SIDEBAR_WIDTH = 240;

/** Horizontal padding/margins in pixels */
const DESKTOP_HORIZONTAL_PADDING = 32;

/** Estimated terminal height ratio (accounting for bottom panel, headers) */
const DESKTOP_HEIGHT_RATIO = 0.65;

// Mobile layout dimensions (matches MobileShell.tsx)
/** Horizontal padding on mobile (8px each side) */
const MOBILE_HORIZONTAL_PADDING = 16;

/** Estimated terminal height ratio on mobile */
const MOBILE_HEIGHT_RATIO = 0.6;

/** Height reserved for mobile controls (pagination, swipe indicator) */
const MOBILE_CONTROLS_HEIGHT = 40;

// =============================================================================
// Public Functions
// =============================================================================

/**
 * Estimate terminal dimensions based on container size
 *
 * @param containerWidth - Container width in pixels
 * @param containerHeight - Container height in pixels
 * @param fontSize - Terminal font size (default: 14)
 * @returns Estimated cols and rows, bounded by MIN/MAX values
 */
export function estimateTerminalSize(
  containerWidth: number,
  containerHeight: number,
  fontSize: number = DEFAULT_FONT_SIZE
): { cols: number; rows: number } {
  // Calculate character dimensions
  const charWidth = fontSize * MONOSPACE_WIDTH_RATIO;
  const lineHeight = fontSize * DEFAULT_LINE_HEIGHT;

  // Calculate cols and rows with bounds
  const cols = Math.min(
    MAX_COLS,
    Math.max(MIN_COLS, Math.floor(containerWidth / charWidth))
  );
  const rows = Math.min(
    MAX_ROWS,
    Math.max(MIN_ROWS, Math.floor(containerHeight / lineHeight))
  );

  return { cols, rows };
}

/**
 * Estimate terminal size for the main workspace area.
 * Accounts for typical UI chrome (sidebars, headers, etc.)
 *
 * This function provides a reasonable initial estimate for PTY creation.
 * The actual size will be corrected immediately after terminal connection
 * via xterm.js FitAddon.
 *
 * @param fontSize - Terminal font size (default: 14)
 * @returns Estimated cols and rows
 */
export function estimateWorkspaceTerminalSize(
  fontSize: number = DEFAULT_FONT_SIZE
): { cols: number; rows: number } {
  // SSR fallback - use conservative defaults
  if (typeof window === "undefined") {
    return { cols: FALLBACK_COLS, rows: FALLBACK_ROWS };
  }

  const isMobile = window.innerWidth < MOBILE_BREAKPOINT;

  if (isMobile) {
    // Mobile layout: terminal takes most of the width, ~60% of height
    const terminalWidth = window.innerWidth - MOBILE_HORIZONTAL_PADDING;
    const terminalHeight =
      window.innerHeight * MOBILE_HEIGHT_RATIO - MOBILE_CONTROLS_HEIGHT;
    return estimateTerminalSize(terminalWidth, terminalHeight, fontSize);
  }

  // Desktop layout: account for activity bar, sidebar, and padding
  const terminalWidth =
    window.innerWidth -
    DESKTOP_ACTIVITY_BAR_WIDTH -
    DESKTOP_SIDEBAR_WIDTH -
    DESKTOP_HORIZONTAL_PADDING;
  const terminalHeight = window.innerHeight * DESKTOP_HEIGHT_RATIO;

  return estimateTerminalSize(terminalWidth, terminalHeight, fontSize);
}

/**
 * Get conservative default terminal size (VT100 standard)
 * Used as fallback when window dimensions are unavailable
 */
export function getDefaultTerminalSize(): { cols: number; rows: number } {
  return { cols: FALLBACK_COLS, rows: FALLBACK_ROWS };
}
