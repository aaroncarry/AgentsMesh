import type { Page } from "@playwright/test";

/**
 * Page Object Model for the Add Runner Modal.
 * Based on: web/src/components/ide/modals/AddRunnerModal.tsx
 */
export class AddRunnerModal {
  constructor(private page: Page) {}

  /** Wait for the modal to be visible. */
  async waitForOpen(): Promise<void> {
    await this.page.locator(".fixed.inset-0").first().waitFor({ state: "visible" });
  }

  /** Get the generated token text. */
  async getToken(): Promise<string | null> {
    const code = this.page.locator("code, pre").first();
    if (await code.isVisible()) return code.textContent();
    return null;
  }

  /** Click the Generate Token button. */
  async generateToken(): Promise<void> {
    await this.page.getByRole("button", { name: /generate/i }).click();
  }

  /** Click the Done/Close button. */
  async close(): Promise<void> {
    const done = this.page.getByRole("button", { name: /done|close|cancel/i });
    if (await done.isVisible()) await done.click();
  }
}
