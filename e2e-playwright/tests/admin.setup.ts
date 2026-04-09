import { test as setup } from "@playwright/test";
import { ADMIN_USER, getWebBaseUrl } from "../helpers/env";

const ADMIN_AUTH_FILE = ".auth/admin.json";

/**
 * Global setup: logs in as the admin user and saves browser state.
 * Tests requiring admin auth use storageState: '.auth/admin.json'.
 */
setup("authenticate as admin user", async ({ browser }) => {
  const context = await browser.newContext();
  const page = await context.newPage();

  await page.goto(`${getWebBaseUrl()}/login`);
  await page.locator("#email").fill(ADMIN_USER.email);
  await page.locator("#password").fill(ADMIN_USER.password);
  await page.locator('button[type="submit"]').click();

  await page.waitForURL((url) => !url.pathname.includes("/login"), {
    timeout: 15_000,
  });

  await context.storageState({ path: ADMIN_AUTH_FILE });
  await context.close();
});
