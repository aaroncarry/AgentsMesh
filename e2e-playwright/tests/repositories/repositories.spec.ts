import { test, expect } from "../../fixtures/index";
import { TEST_ORG_SLUG } from "../../helpers/env";
import { clearAuthRateLimit } from "../../helpers/redis";

test.describe("Repositories API & UI", () => {
  test.beforeEach(async () => { clearAuthRateLimit(); });

  test("list repositories", async ({ api }) => {
    const res = await api.get(`/api/v1/orgs/${TEST_ORG_SLUG}/repositories`);
    expect(res.status).toBe(200);
  });

  test("repositories page loads in UI", async ({ page }) => {
    await page.goto(`/${TEST_ORG_SLUG}/repositories`);
    await page.waitForLoadState("networkidle");
    const body = await page.textContent("body");
    expect(body).toMatch(/repositor|仓库|代码库/i);
  });
});
