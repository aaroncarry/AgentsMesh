import type { Locator, Page } from "@playwright/test";

/**
 * Activity bar navigation sections.
 * Href pattern: /{orgSlug}/{section}
 * Based on: web/src/components/ide/ActivityBar.tsx
 */
export type NavSection =
  | "workspace"
  | "tickets"
  | "channels"
  | "mesh"
  | "loops"
  | "repositories"
  | "runners"
  | "settings";

/**
 * Page Object Model for sidebar/activity bar navigation.
 */
export class SidebarPage {
  constructor(
    private page: Page,
    private orgSlug: string
  ) {}

  /**
   * Get the activity bar nav link locator for a given section.
   * Uses the w-10 h-10 class unique to ActivityBar links to avoid
   * matching other links with the same href (e.g., Logo link).
   */
  getNavLink(section: NavSection): Locator {
    return this.page.locator(
      `a.w-10.h-10[href="/${this.orgSlug}/${section}"]`
    );
  }

  /**
   * Remove Next.js dev overlay that intercepts pointer events in dev mode.
   */
  async dismissDevOverlay(): Promise<void> {
    await this.page.evaluate(() => {
      document.querySelectorAll("nextjs-portal").forEach((el) => el.remove());
    });
  }

  /**
   * Navigate to a section via the activity bar.
   */
  async navigateTo(section: NavSection): Promise<void> {
    await this.dismissDevOverlay();
    const link = this.getNavLink(section);
    await link.click();
    await this.page.waitForURL(`**/${this.orgSlug}/${section}**`);
  }

  /**
   * Check if the current URL matches the given section.
   */
  async isOnSection(section: NavSection): Promise<boolean> {
    const url = this.page.url();
    return url.includes(`/${this.orgSlug}/${section}`);
  }
}
