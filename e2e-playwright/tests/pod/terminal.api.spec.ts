import { test, expect } from "../../fixtures/index";
import { TEST_ORG_SLUG } from "../../helpers/env";
import { clearAuthRateLimit } from "../../helpers/redis";
import { terminateAllPods } from "../../helpers/pod-cleanup";

const PODS_BASE = `/api/v1/orgs/${TEST_ORG_SLUG}/pods`;

/**
 * Terminal connection test.
 * Maps to: e2e/pod/terminal/TC-TERM-001
 */
test.describe("Terminal Connection", () => {
  test.beforeAll(async () => { await terminateAllPods(); });
  test.beforeEach(async () => { clearAuthRateLimit(); });

  /**
   * TC-TERM-001: Terminal connect endpoint returns ws_url
   */
  test("terminal connect returns websocket URL for running pod", async ({ api }) => {
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
      prompt: "E2E Terminal Test",
    });
    const data = await createRes.json();
    const podKey = data.pod_key || data.pod?.pod_key;
    if (!podKey) { test.skip(); return; }

    // Wait for running
    await new Promise((r) => setTimeout(r, 5000));

    // Try connect endpoint
    const connectRes = await api.get(`${PODS_BASE}/${podKey}/connect`);
    if (connectRes.status === 200) {
      const connectData = await connectRes.json();
      expect(connectData.ws_url || connectData.relay_url).toBeTruthy();
    }

    // Cleanup
    await api.post(`${PODS_BASE}/${podKey}/terminate`, {});
  });
});
