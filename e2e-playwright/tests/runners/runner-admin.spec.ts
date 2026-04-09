import { test, expect } from "../../fixtures/index";
import { ADMIN_USER, TEST_ORG_SLUG } from "../../helpers/env";
import { clearAuthRateLimit } from "../../helpers/redis";

/**
 * Runner admin tests — require system admin auth.
 * Maps to: e2e/runner/admin/TC-ADMIN-001~005
 */
test.describe("Runner Admin API", () => {
  test.beforeEach(async () => { clearAuthRateLimit(); });

  /**
   * TC-ADMIN-001: Admin list all runners
   */
  test("admin lists all runners across orgs", async ({ api }) => {
    await api.loginAs(ADMIN_USER.email, ADMIN_USER.password);
    const res = await api.get("/api/v1/admin/runners");
    expect(res.status).toBe(200);
    const data = await res.json();
    expect(data.data).toBeTruthy(); // admin endpoint uses "data" key
  });

  test("non-admin gets 403 on admin runners endpoint", async ({ api }) => {
    // Default login is non-admin user
    const res = await api.get("/api/v1/admin/runners");
    expect(res.status).toBe(403);
  });

  /**
   * TC-ADMIN-002: Admin get single runner
   */
  test("admin gets single runner detail", async ({ api, db }) => {
    db.setup(`
      INSERT INTO runners (organization_id, node_id, description, status, max_concurrent_pods, is_enabled)
      SELECT id, 'admin-single-test', 'Admin Single Test', 'offline', 5, true
      FROM organizations WHERE slug = '${TEST_ORG_SLUG}'
      ON CONFLICT (organization_id, node_id) DO NOTHING
    `);
    const id = db.queryValue(
      `SELECT id FROM runners WHERE node_id = 'admin-single-test'`
    );

    await api.loginAs(ADMIN_USER.email, ADMIN_USER.password);
    const res = await api.get(`/api/v1/admin/runners/${id}`);
    expect(res.status).toBe(200);

    db.cleanup(`DELETE FROM runners WHERE node_id = 'admin-single-test'`);
  });

  /**
   * TC-ADMIN-003: Admin disable/enable runner
   */
  test("admin disables and enables runner", async ({ api, db }) => {
    db.setup(`
      INSERT INTO runners (organization_id, node_id, description, status, max_concurrent_pods, is_enabled)
      SELECT id, 'admin-disable-test', 'Admin Disable Test', 'offline', 5, true
      FROM organizations WHERE slug = '${TEST_ORG_SLUG}'
      ON CONFLICT (organization_id, node_id) DO UPDATE SET is_enabled = true
    `);
    const id = db.queryValue(
      `SELECT id FROM runners WHERE node_id = 'admin-disable-test'`
    );

    await api.loginAs(ADMIN_USER.email, ADMIN_USER.password);

    // Disable
    const disRes = await api.post(`/api/v1/admin/runners/${id}/disable`, {});
    expect(disRes.status).toBe(200);
    expect(db.queryValue(
      `SELECT is_enabled::text FROM runners WHERE node_id = 'admin-disable-test'`
    )).toBe("false");

    // Enable
    const enRes = await api.post(`/api/v1/admin/runners/${id}/enable`, {});
    expect(enRes.status).toBe(200);
    expect(db.queryValue(
      `SELECT is_enabled::text FROM runners WHERE node_id = 'admin-disable-test'`
    )).toBe("true");

    db.cleanup(`DELETE FROM runners WHERE node_id = 'admin-disable-test'`);
  });

  /**
   * TC-ADMIN-004: Admin delete runner
   */
  test("admin deletes runner", async ({ api, db }) => {
    db.setup(`
      INSERT INTO runners (organization_id, node_id, description, status, max_concurrent_pods, is_enabled)
      SELECT id, 'admin-delete-test', 'Admin Delete Test', 'offline', 5, true
      FROM organizations WHERE slug = '${TEST_ORG_SLUG}'
      ON CONFLICT (organization_id, node_id) DO NOTHING
    `);
    const id = db.queryValue(
      `SELECT id FROM runners WHERE node_id = 'admin-delete-test'`
    );

    await api.loginAs(ADMIN_USER.email, ADMIN_USER.password);
    const res = await api.delete(`/api/v1/admin/runners/${id}`);
    expect([200, 204]).toContain(res.status);

    const getRes = await api.get(`/api/v1/admin/runners/${id}`);
    expect(getRes.status).toBe(404);
  });
});
