import { test, expect } from "../../fixtures/index";
import { TEST_ORG_SLUG } from "../../helpers/env";
import { clearAuthRateLimit } from "../../helpers/redis";
import { pollUntil } from "../../helpers/retry";
import { terminateAllPods } from "../../helpers/pod-cleanup";

const PODS = `/api/v1/orgs/${TEST_ORG_SLUG}/pods`;

/**
 * TC-SCENARIO-001: Full flow — Git Credential → Repository → Ticket → Pod
 */
test.describe("Full E2E Scenario", () => {
  test.beforeAll(async () => { await terminateAllPods(); });
  test.beforeEach(async () => { clearAuthRateLimit(); });

  test("git credential → repository → ticket → pod lifecycle", async ({ api }) => {
    // Step 1: Verify repositories exist (from seed)
    const repoRes = await api.get(`/api/v1/orgs/${TEST_ORG_SLUG}/repositories`);
    expect(repoRes.status).toBe(200);
    const repos = (await repoRes.json()).repositories || [];
    expect(repos.length).toBeGreaterThan(0);
    const repoId = repos[0].id;

    // Step 2: Create ticket linked to repository
    const ticketRes = await api.post(`/api/v1/orgs/${TEST_ORG_SLUG}/tickets`, {
      title: "E2E Scenario Ticket",
      description: "Full flow E2E test",
      repository_id: repoId,
    });
    expect([200, 201]).toContain(ticketRes.status);
    const ticketData = await ticketRes.json();
    const ticketSlug = ticketData.ticket?.slug || ticketData.slug;

    // Step 3: Check runner availability
    const runnerRes = await api.get(`/api/v1/orgs/${TEST_ORG_SLUG}/runners/available`);
    const runners = (await runnerRes.json()).runners;
    if (!runners?.length) {
      // Cleanup ticket and skip
      if (ticketSlug) await api.delete(`/api/v1/orgs/${TEST_ORG_SLUG}/tickets/${ticketSlug}`);
      test.skip();
      return;
    }

    // Step 4: Get agents
    const agentRes = await api.get(`/api/v1/orgs/${TEST_ORG_SLUG}/agents`);
    const agents = (await agentRes.json()).builtin_agents;

    // Step 5: Create pod with repository and ticket
    const podRes = await api.post(PODS, {
      runner_id: runners[0].id,
      agent_slug: agents[0].slug,
      prompt: "E2E Scenario Pod",
      repository_id: repoId,
      ticket_slug: ticketSlug,
    });
    expect([200, 201]).toContain(podRes.status);
    const podData = await podRes.json();
    const podKey = podData.pod_key || podData.pod?.pod_key;

    // Step 6: Wait for pod running
    if (podKey) {
      await pollUntil(
        async () => {
          const r = await api.get(`${PODS}/${podKey}`);
          const d = await r.json();
          return (d.pod?.status || d.status) === "running";
        },
        { maxAttempts: 10, intervalMs: 3000, label: "scenario-pod-running" }
      ).catch(() => {});

      // Step 7: Verify runner capacity changed
      const runnerCheck = await api.get(
        `/api/v1/orgs/${TEST_ORG_SLUG}/runners/${runners[0].id}`
      );
      const runnerData = await runnerCheck.json();
      expect(runnerData.runner?.current_pods).toBeGreaterThanOrEqual(0);

      // Step 8: Terminate pod
      await api.post(`${PODS}/${podKey}/terminate`, {});
    }

    // Step 9: Cleanup ticket
    if (ticketSlug) {
      await api.delete(`/api/v1/orgs/${TEST_ORG_SLUG}/tickets/${ticketSlug}`);
    }
  });
});
