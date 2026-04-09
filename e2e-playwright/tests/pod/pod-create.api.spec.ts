import { test, expect } from "../../fixtures/index";
import { TEST_ORG_SLUG } from "../../helpers/env";
import { clearAuthRateLimit } from "../../helpers/redis";

const PODS_BASE = `/api/v1/orgs/${TEST_ORG_SLUG}/pods`;

test.describe("Pod Create API", () => {
  test.beforeEach(async () => { clearAuthRateLimit(); });

  /**
   * TC-POD-001: Create basic pod
   * Maps to: e2e/pod/lifecycle/TC-POD-001-create-basic.yaml
   */
  test("create basic pod", async ({ api }) => {
    // Check available runners first
    const runnersRes = await api.get(
      `/api/v1/orgs/${TEST_ORG_SLUG}/runners/available`
    );
    const runners = (await runnersRes.json()).runners;
    if (!runners || runners.length === 0) { test.skip(); return; }

    // Get agents
    const agentsRes = await api.get(`/api/v1/orgs/${TEST_ORG_SLUG}/agents`);
    const agents = (await agentsRes.json()).builtin_agents;
    if (!agents || agents.length === 0) { test.skip(); return; }

    const res = await api.post(PODS_BASE, {
      runner_id: runners[0].id,
      agent_slug: agents[0].slug,
      prompt: "E2E Test Pod - Basic",
    });
    expect([200, 201]).toContain(res.status);
    const data = await res.json();
    expect(data.pod_key || data.pod?.pod_key).toBeTruthy();

    // Cleanup: terminate
    const podKey = data.pod_key || data.pod?.pod_key;
    if (podKey) {
      await api.post(`${PODS_BASE}/${podKey}/terminate`, {});
    }
  });

  /**
   * TC-POD-003: Terminate pod
   * Maps to: e2e/pod/lifecycle/TC-POD-003-terminate-pod.yaml
   */
  test("terminate pod", async ({ api }) => {
    const runnersRes = await api.get(
      `/api/v1/orgs/${TEST_ORG_SLUG}/runners/available`
    );
    const runners = (await runnersRes.json()).runners;
    if (!runners || runners.length === 0) { test.skip(); return; }

    const agentsRes = await api.get(`/api/v1/orgs/${TEST_ORG_SLUG}/agents`);
    const agents = (await agentsRes.json()).builtin_agents;
    if (!agents || agents.length === 0) { test.skip(); return; }

    // Create pod
    const createRes = await api.post(PODS_BASE, {
      runner_id: runners[0].id,
      agent_slug: agents[0].slug,
      prompt: "E2E Test Pod - Terminate",
    });
    const createData = await createRes.json();
    const podKey = createData.pod_key || createData.pod?.pod_key;
    if (!podKey) { test.skip(); return; }

    // Terminate
    const termRes = await api.post(`${PODS_BASE}/${podKey}/terminate`, {});
    expect(termRes.status).toBe(200);
  });
});
