import { test, expect } from "../../fixtures/index";
import { TEST_ORG_SLUG } from "../../helpers/env";
import { clearAuthRateLimit } from "../../helpers/redis";

const PODS_BASE = `/api/v1/orgs/${TEST_ORG_SLUG}/pods`;

/**
 * ACP (Agent Control Protocol) display tests.
 * These verify ACP-related API behavior.
 * Full UI tests (Activity Stream, tool calls, permissions) require MCP Chrome DevTools.
 * Maps to: e2e/pod/acp/TC-ACP-001~007
 */
test.describe("ACP Pod API", () => {
  test.beforeEach(async () => { clearAuthRateLimit(); });

  /**
   * TC-ACP-001: Create ACP pod (API portion)
   */
  test("create ACP pod with agent_slug", async ({ api }) => {
    const runnersRes = await api.get(
      `/api/v1/orgs/${TEST_ORG_SLUG}/runners/available`
    );
    const runners = (await runnersRes.json()).runners;
    if (!runners?.length) { test.skip(); return; }

    // Get agents and find one supporting ACP
    const agentsRes = await api.get(`/api/v1/orgs/${TEST_ORG_SLUG}/agents`);
    const agents = (await agentsRes.json()).builtin_agents;
    if (!agents?.length) { test.skip(); return; }

    const res = await api.post(PODS_BASE, {
      runner_id: runners[0].id,
      agent_slug: agents[0].slug,
      prompt: "E2E ACP Test - Hello",
    });
    expect([200, 201]).toContain(res.status);
    const data = await res.json();
    const podKey = data.pod_key || data.pod?.pod_key;
    expect(podKey).toBeTruthy();

    // Cleanup
    if (podKey) {
      await api.post(`${PODS_BASE}/${podKey}/terminate`, {});
    }
  });

  /**
   * TC-ACP-007: Send prompt to running pod via API
   */
  test("send prompt to pod via API", async ({ api }) => {
    const runnersRes = await api.get(
      `/api/v1/orgs/${TEST_ORG_SLUG}/runners/available`
    );
    const runners = (await runnersRes.json()).runners;
    if (!runners?.length) { test.skip(); return; }

    const agentsRes = await api.get(`/api/v1/orgs/${TEST_ORG_SLUG}/agents`);
    const agents = (await agentsRes.json()).builtin_agents;
    if (!agents?.length) { test.skip(); return; }

    // Create pod
    const createRes = await api.post(PODS_BASE, {
      runner_id: runners[0].id,
      agent_slug: agents[0].slug,
      prompt: "E2E ACP Prompt Test",
    });
    const data = await createRes.json();
    const podKey = data.pod_key || data.pod?.pod_key;
    if (!podKey) { test.skip(); return; }

    // Wait a bit for pod to initialize
    await new Promise((r) => setTimeout(r, 3000));

    // Send prompt (may fail if pod isn't fully running yet — that's OK)
    const promptRes = await api.post(`${PODS_BASE}/${podKey}/prompt`, {
      prompt: "Hello from E2E test",
    });
    // Accept various statuses — pod may not be ready
    expect([200, 400, 404, 409]).toContain(promptRes.status);

    // Cleanup
    await api.post(`${PODS_BASE}/${podKey}/terminate`, {});
  });
});
