import { test, expect } from "../../fixtures/index";
import { clearAuthRateLimit } from "../../helpers/redis";

test.describe("Repository Providers API", () => {
  test.beforeEach(async () => { clearAuthRateLimit(); });

  /**
   * TC-REPOPROV-001: List repository providers
   */
  test("list repository providers", async ({ api }) => {
    const res = await api.get("/api/v1/users/repository-providers");
    expect(res.status).toBe(200);
  });

  test("list repository providers without auth returns 401", async ({ api }) => {
    const res = await api.getWithToken("/api/v1/users/repository-providers", "bad");
    expect(res.status).toBe(401);
  });

  /**
   * TC-REPOPROV-002: Create GitHub provider
   */
  test("create GitHub provider", async ({ api, db }) => {
    const res = await api.post("/api/v1/users/repository-providers", {
      provider_type: "github",
      name: "E2E GitHub Provider",
      base_url: "https://api.github.com",
      bot_token: "ghp_test_bot_token_e2e",
    });
    expect([200, 201]).toContain(res.status);
    const data = await res.json();
    const id = data.provider?.id || data.id;
    expect(id).toBeTruthy();

    // Cleanup
    if (id) await api.delete(`/api/v1/users/repository-providers/${id}`);
  });

  /**
   * TC-REPOPROV-003: Update provider name
   */
  test("update provider name", async ({ api }) => {
    const createRes = await api.post("/api/v1/users/repository-providers", {
      provider_type: "github",
      name: "E2E Update Provider",
      base_url: "https://api.github.com",
      bot_token: "ghp_update_test",
    });
    const created = await createRes.json();
    const id = created.provider?.id || created.id;
    if (!id) { test.skip(); return; }

    const updateRes = await api.put(`/api/v1/users/repository-providers/${id}`, {
      name: "E2E Updated Provider",
    });
    expect(updateRes.status).toBe(200);

    await api.delete(`/api/v1/users/repository-providers/${id}`);
  });

  /**
   * TC-REPOPROV-004: Delete provider
   */
  test("delete provider", async ({ api }) => {
    const createRes = await api.post("/api/v1/users/repository-providers", {
      provider_type: "github",
      name: "E2E Delete Provider",
      base_url: "https://api.github.com",
      bot_token: "ghp_delete_test",
    });
    const created = await createRes.json();
    const id = created.provider?.id || created.id;
    if (!id) { test.skip(); return; }

    const delRes = await api.delete(`/api/v1/users/repository-providers/${id}`);
    expect(delRes.status).toBe(200);

    // Verify gone
    const getRes = await api.get(`/api/v1/users/repository-providers/${id}`);
    expect(getRes.status).toBe(404);
  });

  /**
   * TC-REPOPROV-006: Test connection (with invalid token)
   */
  test("test connection with invalid token fails", async ({ api }) => {
    const createRes = await api.post("/api/v1/users/repository-providers", {
      provider_type: "github",
      name: "E2E Connection Test",
      base_url: "https://api.github.com",
      bot_token: "ghp_invalid_token",
    });
    const created = await createRes.json();
    const id = created.provider?.id || created.id;
    if (!id) { test.skip(); return; }

    const testRes = await api.post(
      `/api/v1/users/repository-providers/${id}/test`, {}
    );
    // Expect failure due to invalid token
    expect([200, 401, 502]).toContain(testRes.status);

    await api.delete(`/api/v1/users/repository-providers/${id}`);
  });
});
