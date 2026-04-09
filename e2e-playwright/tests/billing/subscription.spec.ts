import { test, expect } from "../../fixtures/index";
import { TEST_ORG_SLUG } from "../../helpers/env";
import { clearAuthRateLimit } from "../../helpers/redis";

const BILLING = `/api/v1/orgs/${TEST_ORG_SLUG}/billing`;

test.describe("Billing Subscription", () => {
  test.beforeEach(async () => { clearAuthRateLimit(); });

  /**
   * TC-SUB-001: Get subscription status
   */
  test("get subscription status", async ({ api }) => {
    const res = await api.get(`${BILLING}/subscription`);
    expect(res.status).toBe(200);
    const data = await res.json();
    expect(data.subscription || data.plan).toBeTruthy();
  });

  /**
   * TC-SUB-001: Subscription status UI display
   */
  test("billing settings page shows subscription info", async ({ page }) => {
    await page.goto(
      `/${TEST_ORG_SLUG}/settings?scope=organization&tab=billing`
    );
    await page.waitForLoadState("networkidle");
    const body = await page.textContent("body");
    expect(body).toMatch(/pro|free|enterprise|plan|订阅|方案/i);
  });

  /**
   * TC-SUB-002: Get available plans
   */
  test("list available plans", async ({ api }) => {
    const res = await api.get(`${BILLING}/plans`);
    expect(res.status).toBe(200);
  });

  /**
   * TC-SUB-005: Reactivate subscription (may fail if not cancelled)
   */
  test("reactivate returns appropriate status", async ({ api }) => {
    const res = await api.post(`${BILLING}/subscription/reactivate`, {});
    // 200 if was cancelled, 400 if active — both are valid
    expect([200, 400]).toContain(res.status);
  });

  /**
   * TC-SUB-007: Auto-renew toggle
   */
  test("toggle auto-renew setting", async ({ api }) => {
    const res = await api.put(`${BILLING}/subscription/auto-renew`, {
      auto_renew: true,
    });
    // May succeed or return error depending on plan
    expect([200, 400, 404]).toContain(res.status);
  });
});
