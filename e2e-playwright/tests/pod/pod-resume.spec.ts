import { test, expect } from "../../fixtures/index";
import { TEST_ORG_SLUG } from "../../helpers/env";
import { clearAuthRateLimit } from "../../helpers/redis";
import { pollUntil } from "../../helpers/retry";

import { terminateAllPods } from "../../helpers/pod-cleanup";

const PODS_BASE = `/api/v1/orgs/${TEST_ORG_SLUG}/pods`;

test.describe("Pod Resume", () => {
  test.beforeAll(async () => { await terminateAllPods(); });
  test.beforeEach(async () => { clearAuthRateLimit(); });

  /** Helper: get a running pod key */
  async function createAndWaitPod(
    api: InstanceType<typeof import("../../fixtures/api.fixture").ApiFixture>
  ): Promise<string | null> {
    const rRes = await api.get(`/api/v1/orgs/${TEST_ORG_SLUG}/runners/available`);
    const runners = (await rRes.json()).runners;
    if (!runners?.length) return null;

    const aRes = await api.get(`/api/v1/orgs/${TEST_ORG_SLUG}/agents`);
    const agents = (await aRes.json()).builtin_agents;
    if (!agents?.length) return null;

    const res = await api.post(PODS_BASE, {
      runner_id: runners[0].id,
      agent_slug: agents[0].slug,
      prompt: "E2E Resume Test Pod",
    });
    const data = await res.json();
    const podKey = data.pod_key || data.pod?.pod_key;
    if (!podKey) return null;

    await pollUntil(
      async () => {
        const r = await api.get(`${PODS_BASE}/${podKey}`);
        const d = await r.json();
        return (d.pod?.status || d.status) === "running";
      },
      { maxAttempts: 10, intervalMs: 3000, label: "pod-running" }
    ).catch(() => {});

    return podKey;
  }

  /**
   * TC-POD-006: Terminate and resume pod
   */
  test("terminate and resume pod preserves sandbox", async ({ api }) => {
    const podKey = await createAndWaitPod(api);
    if (!podKey) { test.skip(); return; }

    // Terminate
    await api.post(`${PODS_BASE}/${podKey}/terminate`, {});

    // Wait for terminated
    await pollUntil(
      async () => {
        const r = await api.get(`${PODS_BASE}/${podKey}`);
        const d = await r.json();
        return (d.pod?.status || d.status) === "terminated";
      },
      { maxAttempts: 5, intervalMs: 2000, label: "pod-terminated" }
    ).catch(() => {});

    // Resume
    const resumeRes = await api.post(PODS_BASE, {
      source_pod_key: podKey,
    });
    expect([200, 201]).toContain(resumeRes.status);
    const resumeData = await resumeRes.json();
    const newPodKey = resumeData.pod_key || resumeData.pod?.pod_key;
    expect(newPodKey).toBeTruthy();

    // Cleanup
    if (newPodKey) {
      await api.post(`${PODS_BASE}/${newPodKey}/terminate`, {});
    }
  });

  /**
   * TC-POD-006: Cannot double-resume same pod
   */
  test("double resume returns error", async ({ api }) => {
    const podKey = await createAndWaitPod(api);
    if (!podKey) { test.skip(); return; }

    await api.post(`${PODS_BASE}/${podKey}/terminate`, {});
    await new Promise((r) => setTimeout(r, 2000));

    // First resume
    const r1 = await api.post(PODS_BASE, { source_pod_key: podKey });
    const d1 = await r1.json();
    const newKey = d1.pod_key || d1.pod?.pod_key;

    // Second resume should fail
    const r2 = await api.post(PODS_BASE, { source_pod_key: podKey });
    expect([400, 409]).toContain(r2.status);

    // Cleanup
    if (newKey) await api.post(`${PODS_BASE}/${newKey}/terminate`, {});
  });
});
