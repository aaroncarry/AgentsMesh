import type { Page } from "@playwright/test";

/**
 * Page Object Model for the Create Pod Modal.
 * Based on: web/src/components/ide/CreatePodModal.tsx
 *           web/src/components/pod/CreatePodForm/index.tsx
 */
export class CreatePodModal {
  constructor(private page: Page) {}

  /** Wait for the modal dialog to appear. */
  async waitForOpen(): Promise<void> {
    await this.page
      .locator('[role="dialog"]')
      .first()
      .waitFor({ state: "visible" });
  }

  /** Select an agent by matching text. */
  async selectAgent(agentName: string): Promise<void> {
    const agent = this.page.getByText(agentName, { exact: false }).first();
    if (await agent.isVisible()) await agent.click();
  }

  /** Fill the prompt text area. */
  async fillPrompt(prompt: string): Promise<void> {
    const textarea = this.page.locator("textarea").first();
    await textarea.fill(prompt);
  }

  /** Click the submit/create button. */
  async submit(): Promise<void> {
    const btn = this.page
      .locator('[role="dialog"]')
      .getByRole("button", { name: /create|创建/i });
    await btn.click();
  }

  /** Close the modal. */
  async cancel(): Promise<void> {
    const btn = this.page
      .locator('[role="dialog"]')
      .getByRole("button", { name: /cancel|取消/i });
    if (await btn.isVisible()) await btn.click();
  }
}
