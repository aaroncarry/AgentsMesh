import { test, expect } from "../../fixtures/index";
import { TEST_ORG_SLUG } from "../../helpers/env";
import { clearAuthRateLimit } from "../../helpers/redis";

test.describe("Runner API CRUD", () => {
  test.beforeEach(async () => { clearAuthRateLimit(); });

  /**
   * TC-RUNNER-001: List runners
   * Maps to: e2e/runner/list/TC-RUNNER-001-list-runners.yaml
   */
  test("list runners returns array", async ({ api }) => {
    const res = await api.get(`/api/v1/orgs/${TEST_ORG_SLUG}/runners`);
    expect(res.status).toBe(200);
    const data = await res.json();
    expect(Array.isArray(data.runners)).toBe(true);
  });

  test("list runners without auth returns 401", async ({ api }) => {
    const res = await api.getWithToken(
      `/api/v1/orgs/${TEST_ORG_SLUG}/runners`, "bad-token"
    );
    expect(res.status).toBe(401);
  });

  /**
   * TC-RUNNER-002: List available runners
   * Maps to: e2e/runner/list/TC-RUNNER-002-list-available.yaml
   */
  test("list available runners", async ({ api }) => {
    const res = await api.get(`/api/v1/orgs/${TEST_ORG_SLUG}/runners/available`);
    expect(res.status).toBe(200);
  });

  /**
   * TC-RUNNER-003: Get single runner detail
   * Maps to: e2e/runner/list/TC-RUNNER-003-get-runner.yaml
   */
  test("get runner by id", async ({ api, db }) => {
    const id = db.queryValue(
      `SELECT id FROM runners WHERE organization_id = (SELECT id FROM organizations WHERE slug = '${TEST_ORG_SLUG}') LIMIT 1`
    );
    if (!id) { test.skip(); return; }

    const res = await api.get(`/api/v1/orgs/${TEST_ORG_SLUG}/runners/${id}`);
    expect(res.status).toBe(200);
    const data = await res.json();
    expect(data.runner.id).toBeTruthy();
  });

  test("get non-existent runner returns 404", async ({ api }) => {
    const res = await api.get(`/api/v1/orgs/${TEST_ORG_SLUG}/runners/999999`);
    expect(res.status).toBe(404);
  });

  /**
   * TC-CONFIG-001: Update runner config
   * Maps to: e2e/runner/config/TC-CONFIG-001-update-runner.yaml
   */
  test("update runner description", async ({ api, db }) => {
    db.setup(`
      INSERT INTO runners (organization_id, node_id, description, status, max_concurrent_pods, is_enabled)
      SELECT id, 'test-runner-config', 'Config Test', 'offline', 5, true
      FROM organizations WHERE slug = '${TEST_ORG_SLUG}'
      ON CONFLICT (organization_id, node_id) DO UPDATE SET description = NULL
    `);
    const id = db.queryValue(
      `SELECT id FROM runners WHERE node_id = 'test-runner-config'`
    );

    const res = await api.put(`/api/v1/orgs/${TEST_ORG_SLUG}/runners/${id}`, {
      description: "Updated by E2E test",
    });
    expect(res.status).toBe(200);

    const desc = db.queryValue(
      `SELECT description FROM runners WHERE node_id = 'test-runner-config'`
    );
    expect(desc).toContain("Updated by E2E");

    db.cleanup(`DELETE FROM runners WHERE node_id = 'test-runner-config'`);
  });

  /**
   * TC-CONFIG-002: Disable and enable runner
   * Maps to: e2e/runner/config/TC-CONFIG-002-disable-enable.yaml
   */
  test("disable and enable runner", async ({ api, db }) => {
    db.setup(`
      INSERT INTO runners (organization_id, node_id, description, status, max_concurrent_pods, is_enabled)
      SELECT id, 'test-runner-disable', 'Disable Test', 'offline', 5, true
      FROM organizations WHERE slug = '${TEST_ORG_SLUG}'
      ON CONFLICT (organization_id, node_id) DO UPDATE SET is_enabled = true
    `);
    const id = db.queryValue(
      `SELECT id FROM runners WHERE node_id = 'test-runner-disable'`
    );

    // Disable (set is_enabled to false)
    const disableRes = await api.put(
      `/api/v1/orgs/${TEST_ORG_SLUG}/runners/${id}`, { is_enabled: false }
    );
    expect(disableRes.status).toBe(200);

    let flag = db.queryValue(
      `SELECT is_enabled::text FROM runners WHERE node_id = 'test-runner-disable'`
    );
    expect(flag).toBe("false");

    // Enable (set is_enabled to true)
    const enableRes = await api.put(
      `/api/v1/orgs/${TEST_ORG_SLUG}/runners/${id}`, { is_enabled: true }
    );
    expect(enableRes.status).toBe(200);

    flag = db.queryValue(
      `SELECT is_enabled::text FROM runners WHERE node_id = 'test-runner-disable'`
    );
    expect(flag).toBe("true");

    db.cleanup(`DELETE FROM runners WHERE node_id = 'test-runner-disable'`);
  });

  /**
   * TC-DELETE-001: Delete runner
   * Maps to: e2e/runner/delete/TC-DELETE-001-delete-runner.yaml
   */
  test("delete runner", async ({ api, db }) => {
    db.setup(`
      INSERT INTO runners (organization_id, node_id, description, status, max_concurrent_pods, is_enabled)
      SELECT id, 'test-runner-delete', 'Delete Test', 'offline', 5, true
      FROM organizations WHERE slug = '${TEST_ORG_SLUG}'
      ON CONFLICT (organization_id, node_id) DO NOTHING
    `);
    const id = db.queryValue(
      `SELECT id FROM runners WHERE node_id = 'test-runner-delete'`
    );

    const delRes = await api.delete(`/api/v1/orgs/${TEST_ORG_SLUG}/runners/${id}`);
    expect([200, 204]).toContain(delRes.status);

    const getRes = await api.get(`/api/v1/orgs/${TEST_ORG_SLUG}/runners/${id}`);
    expect(getRes.status).toBe(404);
  });
});
