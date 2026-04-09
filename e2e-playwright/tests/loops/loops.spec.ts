import { test, expect } from "../../fixtures/index";
import { TEST_ORG_SLUG } from "../../helpers/env";
import { clearAuthRateLimit } from "../../helpers/redis";

test.describe("Loops API & UI", () => {
  test.beforeEach(async () => { clearAuthRateLimit(); });

  test("list loops", async ({ api }) => {
    const res = await api.get(`/api/v1/orgs/${TEST_ORG_SLUG}/loops`);
    expect(res.status).toBe(200);
  });

  test("loops page loads in UI", async ({ page }) => {
    await page.goto(`/${TEST_ORG_SLUG}/loops`);
    await page.waitForLoadState("networkidle");
    const body = await page.textContent("body");
    expect(body).toMatch(/loop|循环|定时/i);
  });
});
