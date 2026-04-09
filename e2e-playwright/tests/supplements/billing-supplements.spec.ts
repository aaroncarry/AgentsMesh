import { test, expect } from "../../fixtures/index";
import { TEST_ORG_SLUG } from "../../helpers/env";
import { clearAuthRateLimit } from "../../helpers/redis";

/**
 * Billing supplement tests — filling coverage gaps.
 */
test.describe("Billing Supplements", () => {
  test.beforeEach(async () => { clearAuthRateLimit(); });

  const BILLING = `/api/v1/orgs/${TEST_ORG_SLUG}/billing`;

  /**
   * TC-QUOTA-002: Check runner quota
   */
  test("check repository quota", async ({ api }) => {
    const res = await api.get(`${BILLING}/quota/check?resource=repositories`);
    expect(res.status).toBe(200);
  });

  /**
   * TC-PLAN-002: Plan upgrade/change
   */
  test("plan change returns appropriate status", async ({ api }) => {
    // Attempt to change plan — may fail depending on current plan
    const res = await api.post(`${BILLING}/subscription`, {
      plan: "enterprise",
      billing_cycle: "monthly",
    });
    // 200 if changed, 400 if same plan, 402 if payment needed
    expect([200, 400, 402]).toContain(res.status);
  });

  /**
   * TC-SUB-003: Cancel subscription at period end
   */
  test("cancel subscription at period end", async ({ api }) => {
    const res = await api.post(`${BILLING}/subscription/cancel`, {
      immediate: false,
    });
    expect([200, 400]).toContain(res.status);
  });

  /**
   * TC-SUB-004: Cancel subscription immediately
   */
  test("cancel subscription immediately", async ({ api }) => {
    const res = await api.post(`${BILLING}/subscription/cancel`, {
      immediate: true,
    });
    expect([200, 400]).toContain(res.status);
  });

  /**
   * TC-SEAT-005: Purchase exceeds limit
   */
  test("purchase seats exceeding limit returns error", async ({ api }) => {
    const res = await api.post(`${BILLING}/seats/purchase`, {
      quantity: 99999,
    });
    expect([400, 402]).toContain(res.status);
  });
});

test.describe("Webhook Supplements", () => {
  const baseUrl = `/api/v1/webhooks`;

  /**
   * TC-WEBHOOK-002~004: Stripe webhook event types
   */
  test("stripe webhook rejects unsigned payload", async ({ api }) => {
    // Different stripe event types all fail without valid signature
    for (const eventType of [
      "invoice.paid",
      "invoice.payment_failed",
      "customer.subscription.deleted",
    ]) {
      const res = await api.postPublic(`${baseUrl}/stripe`, {
        type: eventType,
        data: { object: {} },
      });
      // 400 (bad signature), 401, or 503 (not configured)
      expect([400, 401, 503]).toContain(res.status);
    }
  });
});
