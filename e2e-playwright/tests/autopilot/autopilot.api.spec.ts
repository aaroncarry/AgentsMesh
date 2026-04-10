import { test, expect } from "../../fixtures/index";
import { TEST_ORG_SLUG } from "../../helpers/env";
import { clearAuthRateLimit } from "../../helpers/redis";

test.describe("Autopilot API", () => {
  test.beforeEach(async () => { clearAuthRateLimit(); });

  /**
   * TC-AUTOPILOT-001: List autopilot controllers
   */
  test("list autopilot controllers", async ({ api }) => {
    const res = await api.get(
      `/api/v1/orgs/${TEST_ORG_SLUG}/autopilot-controllers`
    );
    expect(res.status).toBe(200);
  });
});
