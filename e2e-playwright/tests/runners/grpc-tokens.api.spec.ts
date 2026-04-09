import { test, expect } from "../../fixtures/index";
import { TEST_ORG_SLUG } from "../../helpers/env";
import { clearAuthRateLimit } from "../../helpers/redis";

const GRPC_BASE = `/api/v1/orgs/${TEST_ORG_SLUG}/runners/grpc/tokens`;

test.describe("gRPC Registration Tokens", () => {
  test.beforeEach(async () => { clearAuthRateLimit(); });

  /**
   * TC-GRPC-001: List gRPC tokens
   * Maps to: e2e/runner/grpc-tokens/TC-GRPC-001-list-tokens.yaml
   */
  test("list gRPC tokens returns array", async ({ api }) => {
    const res = await api.get(GRPC_BASE);
    expect(res.status).toBe(200);
    const data = await res.json();
    expect(Array.isArray(data.tokens)).toBe(true);
  });

  /**
   * TC-GRPC-002: Create gRPC token
   * Maps to: e2e/runner/grpc-tokens/TC-GRPC-002-create-token.yaml
   */
  test("create gRPC token returns token string", async ({ api, db }) => {
    const res = await api.post(GRPC_BASE, { name: "E2E gRPC test token" });
    expect(res.status).toBe(201);
    const data = await res.json();
    expect(data.token).toBeTruthy();
    expect(data.command).toBeTruthy();

    db.cleanup(
      `DELETE FROM runner_grpc_registration_tokens WHERE name = 'E2E gRPC test token'`
    );
  });

  /**
   * TC-GRPC-003: Delete gRPC token
   * Maps to: e2e/runner/grpc-tokens/TC-GRPC-003-delete-token.yaml
   */
  test("delete gRPC token", async ({ api }) => {
    // Create
    await api.post(GRPC_BASE, { name: "E2E delete test token" });

    // Find its ID via list
    const listRes = await api.get(GRPC_BASE);
    const tokens = (await listRes.json()).tokens;
    const token = tokens.find(
      (t: { name?: string }) => t.name === "E2E delete test token"
    );
    expect(token).toBeTruthy();

    // Delete
    const delRes = await api.delete(`${GRPC_BASE}/${token.id}`);
    expect(delRes.status).toBe(200);

    // Re-delete should fail
    const reDelRes = await api.delete(`${GRPC_BASE}/${token.id}`);
    expect(reDelRes.status).toBe(404);
  });

  /**
   * TC-GRPC-004: Full CRUD flow
   * Maps to: e2e/runner/grpc-tokens/TC-GRPC-004-full-crud.yaml
   */
  test("gRPC tokens full CRUD flow", async ({ api }) => {
    const listRes1 = await api.get(GRPC_BASE);
    const initialCount = (await listRes1.json()).tokens.length;

    // Create 2 tokens
    await api.post(GRPC_BASE, { name: "CRUD test token 1" });
    await api.post(GRPC_BASE, { name: "CRUD test token 2" });

    // Verify count increased
    const listRes2 = await api.get(GRPC_BASE);
    const tokens2 = (await listRes2.json()).tokens;
    expect(tokens2.length).toBe(initialCount + 2);

    // Find IDs
    const t1 = tokens2.find((t: { name?: string }) => t.name === "CRUD test token 1");
    const t2 = tokens2.find((t: { name?: string }) => t.name === "CRUD test token 2");

    // Delete both
    await api.delete(`${GRPC_BASE}/${t1.id}`);
    await api.delete(`${GRPC_BASE}/${t2.id}`);

    // Verify count restored
    const listRes3 = await api.get(GRPC_BASE);
    expect((await listRes3.json()).tokens.length).toBe(initialCount);
  });
});
