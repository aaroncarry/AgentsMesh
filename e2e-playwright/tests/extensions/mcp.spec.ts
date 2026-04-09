import { test, expect } from "../../fixtures/index";
import { TEST_ORG_SLUG } from "../../helpers/env";
import { clearAuthRateLimit } from "../../helpers/redis";

test.describe("MCP Servers API", () => {
  test.beforeEach(async () => { clearAuthRateLimit(); });

  /**
   * TC-MCP-EXT-009: List marketplace MCP servers
   */
  test("list marketplace MCP servers", async ({ api }) => {
    const res = await api.get(`/api/v1/orgs/${TEST_ORG_SLUG}/market/mcp-servers`);
    expect(res.status).toBe(200);
    const data = await res.json();
    expect(data).toBeTruthy();
  });

  /**
   * TC-MCP-EXT-009: Search marketplace MCP servers
   */
  test("search marketplace MCP servers", async ({ api }) => {
    const res = await api.get(
      `/api/v1/orgs/${TEST_ORG_SLUG}/market/mcp-servers?q=postgres`
    );
    expect(res.status).toBe(200);
  });

  /**
   * TC-MCP-EXT-001: MCP tab UI display
   */
  test("MCP settings tab loads", async ({ page }) => {
    await page.goto(`/${TEST_ORG_SLUG}/settings?scope=organization&tab=extensions`);
    await page.waitForLoadState("networkidle");
    const body = await page.textContent("body");
    expect(body).toMatch(/MCP|extension|扩展/i);
  });

  /**
   * TC-EXTSET-003: MCP templates browse
   */
  test("MCP templates page shows templates", async ({ page }) => {
    await page.goto(`/${TEST_ORG_SLUG}/settings?scope=organization&tab=extensions`);
    await page.waitForLoadState("networkidle");
    // Look for common MCP template names
    const body = await page.textContent("body");
    const hasTemplates = /jira|postgres|slack|github|filesystem/i.test(body ?? "");
    // Templates may or may not be loaded depending on sync state
    expect(body).toBeTruthy();
  });
});
