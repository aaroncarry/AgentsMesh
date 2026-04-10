import { test, expect } from "../../fixtures/index";
import { TEST_ORG_SLUG } from "../../helpers/env";
import { clearAuthRateLimit } from "../../helpers/redis";

const BILLING = `/api/v1/orgs/${TEST_ORG_SLUG}/billing`;

test.describe("Billing Quota", () => {
  test.beforeEach(async () => { clearAuthRateLimit(); });

  /**
   * TC-QUOTA-001/002/003: Check quota
   */
  test("check user quota", async ({ api }) => {
    const res = await api.get(`${BILLING}/quota/check?resource=users`);
    expect(res.status).toBe(200);
  });

  test("check runner quota", async ({ api }) => {
    const res = await api.get(`${BILLING}/quota/check?resource=runners`);
    expect(res.status).toBe(200);
  });
});

test.describe("Billing Usage", () => {
  /**
   * TC-USAGE-001: Get current usage
   */
  test("get current usage", async ({ api }) => {
    const res = await api.get(`${BILLING}/usage`);
    expect(res.status).toBe(200);
  });

  /**
   * TC-USAGE-002: Get usage history
   */
  test("get usage history", async ({ api }) => {
    const res = await api.get(`${BILLING}/usage/history`);
    expect(res.status).toBe(200);
  });
});

test.describe("Billing Invoices", () => {
  /**
   * TC-INVOICE-001: List invoices
   */
  test("list invoices", async ({ api }) => {
    const res = await api.get(`${BILLING}/invoices`);
    expect(res.status).toBe(200);
  });
});

test.describe("Billing Checkout", () => {
  /**
   * TC-CHECKOUT-001: Create checkout session
   */
  test("create checkout session returns appropriate status", async ({ api }) => {
    const res = await api.post(`${BILLING}/checkout`, {
      plan: "pro",
      billing_cycle: "monthly",
    });
    // 200/201 if session created, 400 if already on plan
    expect([200, 201, 400]).toContain(res.status);
  });

  /**
   * TC-CHECKOUT-002: Query non-existent checkout status
   */
  test("query non-existent checkout returns 404", async ({ api }) => {
    const res = await api.get(`${BILLING}/checkout/non-existent-order`);
    expect(res.status).toBe(404);
  });
});

test.describe("Billing Deployment Info", () => {
  /**
   * TC-DEPLOY-001: Get deployment info (public)
   */
  test("get deployment config", async ({ api }) => {
    const res = await api.get("/api/v1/config/deployment");
    expect(res.status).toBe(200);
  });

  test("get pricing config", async ({ api }) => {
    const res = await api.get("/api/v1/config/pricing");
    expect(res.status).toBe(200);
  });
});
