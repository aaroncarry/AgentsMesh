import { test, expect } from "../../fixtures/index";
import { TEST_ORG_SLUG } from "../../helpers/env";
import { clearAuthRateLimit } from "../../helpers/redis";

test.describe("Tickets API & UI", () => {
  test.beforeEach(async () => { clearAuthRateLimit(); });

  test("list tickets", async ({ api }) => {
    const res = await api.get(`/api/v1/orgs/${TEST_ORG_SLUG}/tickets`);
    expect(res.status).toBe(200);
  });

  test("get ticket board", async ({ api }) => {
    const res = await api.get(`/api/v1/orgs/${TEST_ORG_SLUG}/tickets/board`);
    expect(res.status).toBe(200);
  });

  test("create and delete ticket", async ({ api }) => {
    const createRes = await api.post(`/api/v1/orgs/${TEST_ORG_SLUG}/tickets`, {
      title: "E2E Test Ticket",
      description: "Created by Playwright E2E",
    });
    expect([200, 201]).toContain(createRes.status);
    const data = await createRes.json();
    const slug = data.ticket?.slug || data.slug;

    if (slug) {
      const delRes = await api.delete(
        `/api/v1/orgs/${TEST_ORG_SLUG}/tickets/${slug}`
      );
      expect([200, 204]).toContain(delRes.status);
    }
  });

  test("tickets page loads in UI", async ({ page }) => {
    await page.goto(`/${TEST_ORG_SLUG}/tickets`);
    await page.waitForLoadState("networkidle");
    const body = await page.textContent("body");
    expect(body).toMatch(/ticket|工单|任务/i);
  });
});
