import { test, expect } from "../../fixtures/index";
import { TEST_ORG_SLUG } from "../../helpers/env";
import { clearAuthRateLimit } from "../../helpers/redis";

/**
 * Runner registration token tests.
 * The registration tokens use the gRPC tokens API.
 * Maps to: e2e/runner/tokens/
 */

const TOKENS_BASE = `/api/v1/orgs/${TEST_ORG_SLUG}/runners/grpc/tokens`;

test.describe("Runner Registration Tokens", () => {
  test.beforeEach(async () => { clearAuthRateLimit(); });

  /**
   * TC-RTOKEN-001: List registration tokens
   */
  test("list registration tokens", async ({ api }) => {
    const res = await api.get(TOKENS_BASE);
    expect(res.status).toBe(200);
    const data = await res.json();
    expect(data.tokens).toBeTruthy();
  });

  /**
   * TC-RTOKEN-002: Create registration token
   */
  test("create registration token", async ({ api, db }) => {
    const res = await api.post(TOKENS_BASE, { name: "E2E reg token test" });
    expect(res.status).toBe(201);
    const data = await res.json();
    expect(data.token).toBeTruthy();

    db.cleanup(
      `DELETE FROM runner_grpc_registration_tokens WHERE name = 'E2E reg token test'`
    );
  });

  /**
   * TC-RTOKEN-003: Revoke registration token
   */
  test("revoke registration token", async ({ api }) => {
    await api.post(TOKENS_BASE, { name: "E2E revoke test" });

    // Find ID via list
    const listRes = await api.get(TOKENS_BASE);
    const tokens = (await listRes.json()).tokens;
    const token = tokens.find((t: { name?: string }) => t.name === "E2E revoke test");
    expect(token).toBeTruthy();

    const delRes = await api.delete(`${TOKENS_BASE}/${token.id}`);
    expect(delRes.status).toBe(200);
  });
});
