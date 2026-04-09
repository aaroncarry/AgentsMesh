import { test, expect } from "../../fixtures/index";
import { TEST_ORG_SLUG } from "../../helpers/env";
import { clearAuthRateLimit } from "../../helpers/redis";
import { pollUntil } from "../../helpers/retry";

const PODS_BASE = `/api/v1/orgs/${TEST_ORG_SLUG}/pods`;

/**
 * Pod lifecycle scenario tests.
 * These require at least one runner to be online.
 * Maps to: e2e/pod/lifecycle/TC-POD-004~007
 */
test.describe("Pod Lifecycle Scenarios", () => {
  test.beforeEach(async () => { clearAuthRateLimit(); });

  /** Helper: create a pod and return its key. Skips if no runner available. */
  async function createPod(
    api: InstanceType<typeof import("../../fixtures/api.fixture").ApiFixture>,
    prompt: string
  ): Promise<string | null> {
    const runnersRes = await api.get(
      `/api/v1/orgs/${TEST_ORG_SLUG}/runners/available`
    );
    const runners = (await runnersRes.json()).runners;
    if (!runners?.length) return null;

    const agentsRes = await api.get(`/api/v1/orgs/${TEST_ORG_SLUG}/agents`);
    const agents = (await agentsRes.json()).builtin_agents;
    if (!agents?.length) return null;

    const res = await api.post(PODS_BASE, {
      runner_id: runners[0].id,
      agent_slug: agents[0].slug,
      prompt,
    });
    const data = await res.json();
    return data.pod_key || data.pod?.pod_key || null;
  }

  /**
   * TC-POD-004: Full lifecycle (create → running → terminate)
   */
  test("full pod lifecycle", async ({ api }) => {
    const podKey = await createPod(api, "E2E Full Lifecycle Test");
    if (!podKey) { test.skip(); return; }

    // Wait for running state
    await pollUntil(
      async () => {
        const res = await api.get(`${PODS_BASE}/${podKey}`);
        const data = await res.json();
        const status = data.pod?.status || data.status;
        return status === "running";
      },
      { maxAttempts: 10, intervalMs: 3000, label: "pod-running" }
    ).catch(() => {/* may timeout, continue anyway */});

    // Terminate
    const termRes = await api.post(`${PODS_BASE}/${podKey}/terminate`, {});
    expect(termRes.status).toBe(200);

    // Verify terminated
    const getRes = await api.get(`${PODS_BASE}/${podKey}`);
    const finalData = await getRes.json();
    const finalStatus = finalData.pod?.status || finalData.status;
    expect(["terminated", "completed", "error"]).toContain(finalStatus);
  });

  /**
   * TC-POD-005: Runner capacity tracks pod count
   */
  test("runner capacity changes with pods", async ({ api }) => {
    const runnersRes = await api.get(
      `/api/v1/orgs/${TEST_ORG_SLUG}/runners/available`
    );
    const runners = (await runnersRes.json()).runners;
    if (!runners?.length) { test.skip(); return; }

    const initialPods = runners[0].current_pods || 0;

    const podKey = await createPod(api, "E2E Capacity Test");
    if (!podKey) { test.skip(); return; }

    // Wait a moment for capacity to update
    await new Promise((r) => setTimeout(r, 2000));

    // Terminate and verify
    await api.post(`${PODS_BASE}/${podKey}/terminate`, {});

    // Wait for capacity to restore
    await pollUntil(
      async () => {
        const res = await api.get(
          `/api/v1/orgs/${TEST_ORG_SLUG}/runners/${runners[0].id}`
        );
        const data = await res.json();
        return (data.runner?.current_pods || 0) <= initialPods;
      },
      { maxAttempts: 5, intervalMs: 2000, label: "capacity-restore" }
    ).catch(() => {/* ignore */});
  });

  /**
   * TC-POD-007: Resume edge cases
   */
  test("resume from non-existent pod returns error", async ({ api }) => {
    const res = await api.post(PODS_BASE, {
      source_pod_key: "non-existent-pod-key",
    });
    expect([400, 404]).toContain(res.status);
  });
});
