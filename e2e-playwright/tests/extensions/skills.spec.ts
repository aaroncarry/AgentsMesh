import { test, expect } from "../../fixtures/index";
import { TEST_ORG_SLUG } from "../../helpers/env";
import { clearAuthRateLimit } from "../../helpers/redis";

test.describe("Extensions Skills API", () => {
  test.beforeEach(async () => { clearAuthRateLimit(); });

  /**
   * TC-SKILL-007: List marketplace skills
   */
  test("list marketplace skills", async ({ api }) => {
    const res = await api.get(`/api/v1/orgs/${TEST_ORG_SLUG}/market/skills`);
    expect(res.status).toBe(200);
  });

  /**
   * TC-SKILL-001: Skills tab UI displays
   */
  test("extensions settings page loads", async ({ page }) => {
    await page.goto(`/${TEST_ORG_SLUG}/settings?scope=organization&tab=extensions`);
    await page.waitForLoadState("networkidle");
    const body = await page.textContent("body");
    expect(body).toMatch(/extension|skill|扩展|技能/i);
  });
});
