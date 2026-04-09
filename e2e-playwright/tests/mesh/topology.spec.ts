import { test, expect } from "../../fixtures/index";
import { TEST_ORG_SLUG } from "../../helpers/env";
import { clearAuthRateLimit } from "../../helpers/redis";

test.describe("Mesh Topology API", () => {
  test.beforeEach(async () => { clearAuthRateLimit(); });

  /**
   * TC-MESH-001: Get mesh topology (empty or populated)
   */
  test("get mesh topology returns structure", async ({ api }) => {
    const res = await api.get(`/api/v1/orgs/${TEST_ORG_SLUG}/mesh/topology`);
    expect(res.status).toBe(200);
    const data = await res.json();
    expect(data).toBeTruthy();
  });

  /**
   * TC-MESH-001: Mesh page loads in UI
   */
  test("mesh page loads correctly", async ({ page }) => {
    await page.goto(`/${TEST_ORG_SLUG}/mesh`);
    await page.waitForLoadState("networkidle");
    const body = await page.textContent("body");
    expect(body).toMatch(/mesh|topology|网格|拓扑/i);
  });
});
