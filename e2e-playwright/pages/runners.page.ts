import type { Locator, Page } from "@playwright/test";

/**
 * Page Object Model for the Runners page.
 * Based on: web/src/app/(dashboard)/[org]/runners/page.tsx
 *           web/src/components/ide/RunnersSidebarContent.tsx
 */
export class RunnersPage {
  readonly addRunnerButton: Locator;
  readonly runnerTable: Locator;

  constructor(
    private page: Page,
    private orgSlug: string
  ) {
    // Scope to main content area to avoid matching sidebar's "Add Runner" button
    const main = page.getByRole("main");
    this.addRunnerButton = main.getByRole("button", {
      name: /add runner/i,
    });
    this.runnerTable = main.locator("table");
  }

  async goto(): Promise<void> {
    await this.page.goto(`/${this.orgSlug}/runners`);
    await this.page.waitForLoadState("networkidle");
  }

  /**
   * Wait for the runner list to be rendered (sidebar or main content).
   */
  async waitForList(): Promise<void> {
    // Wait for at least one runner item or the empty state
    await this.page
      .locator("table tbody tr, [data-testid='empty-state']")
      .first()
      .waitFor({ state: "visible", timeout: 15_000 })
      .catch(() => {
        // Table may not exist if using card layout on mobile
      });
  }

  /**
   * Get a runner row by its node ID text.
   */
  getRunnerByNodeId(nodeId: string): Locator {
    return this.page.locator("table tbody tr").filter({
      hasText: nodeId,
    });
  }

  /**
   * Get the stat card values (Total, Online, Active Pods, Capacity).
   */
  async getStatValue(label: string): Promise<string | null> {
    const card = this.page.locator("div").filter({ hasText: label }).first();
    const value = card.locator(".text-2xl, .text-3xl").first();
    if (await value.isVisible()) {
      return value.textContent();
    }
    return null;
  }

  /**
   * Count visible runner rows in the table.
   */
  async getRunnerCount(): Promise<number> {
    return this.page.locator("table tbody tr").count();
  }
}
