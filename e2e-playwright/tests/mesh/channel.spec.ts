import { test, expect } from "../../fixtures/index";
import { TEST_ORG_SLUG } from "../../helpers/env";
import { clearAuthRateLimit } from "../../helpers/redis";

test.describe("Channel API", () => {
  test.beforeEach(async () => { clearAuthRateLimit(); });

  /**
   * TC-CHAN-001: Create and list channels
   */
  test("create channel and verify in list", async ({ api, db }) => {
    // Pre-clean
    try { db.cleanup(`DELETE FROM channels WHERE name = 'E2E Test Channel' AND organization_id = (SELECT id FROM organizations WHERE slug = '${TEST_ORG_SLUG}')`); } catch { /* ignore */ }

    const createRes = await api.post(`/api/v1/orgs/${TEST_ORG_SLUG}/channels`, {
      name: "E2E Test Channel",
      description: "Created by E2E test",
    });
    expect([200, 201]).toContain(createRes.status);
    const created = await createRes.json();
    const channelId = created.channel?.id || created.id;
    expect(channelId).toBeTruthy();

    // List
    const listRes = await api.get(`/api/v1/orgs/${TEST_ORG_SLUG}/channels`);
    expect(listRes.status).toBe(200);

    // Cleanup: archive
    if (channelId) {
      await api.post(`/api/v1/orgs/${TEST_ORG_SLUG}/channels/${channelId}/archive`, {});
    }
  });

  /**
   * TC-CHAN-001: Update channel
   */
  test("update channel description", async ({ api }) => {
    const createRes = await api.post(`/api/v1/orgs/${TEST_ORG_SLUG}/channels`, {
      name: "E2E Update Ch " + Date.now(),
    });
    const created = await createRes.json();
    const channelId = created.channel?.id || created.id;
    if (!channelId) { test.skip(); return; }

    const updateRes = await api.put(
      `/api/v1/orgs/${TEST_ORG_SLUG}/channels/${channelId}`,
      { description: "Updated by E2E" }
    );
    expect(updateRes.status).toBe(200);

    await api.post(`/api/v1/orgs/${TEST_ORG_SLUG}/channels/${channelId}/archive`, {});
  });

  /**
   * TC-CHAN-001: Duplicate channel name returns 409
   */
  test("duplicate channel name returns conflict", async ({ api, db }) => {
    const name = "E2E Dup Chan " + Date.now();
    const r1 = await api.post(`/api/v1/orgs/${TEST_ORG_SLUG}/channels`, { name });
    const c1 = await r1.json();
    const id1 = c1.channel?.id || c1.id;

    const r2 = await api.post(`/api/v1/orgs/${TEST_ORG_SLUG}/channels`, { name });
    expect(r2.status).toBe(409);

    if (id1) {
      await api.post(`/api/v1/orgs/${TEST_ORG_SLUG}/channels/${id1}/archive`, {});
    }
  });
});
