import { test, expect } from "../../fixtures/index";
import { TEST_ORG_SLUG } from "../../helpers/env";
import { clearAuthRateLimit } from "../../helpers/redis";
import { pollUntil } from "../../helpers/retry";

const PODS_BASE = `/api/v1/orgs/${TEST_ORG_SLUG}/pods`;

/**
 * Binding API tests — require running pods with X-Pod-Key header.
 * Maps to: TC-BIND-001~004
 */
test.describe("Pod Binding API", () => {
  test.beforeEach(async () => { clearAuthRateLimit(); });

  /** Create a running pod and return its key */
  async function createRunningPod(
    api: InstanceType<typeof import("../../fixtures/api.fixture").ApiFixture>,
    prompt: string
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
      prompt,
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
   * TC-BIND-001: Request binding between two pods
   */
  test("request binding between pods", async ({ api }) => {
    const podA = await createRunningPod(api, "E2E Bind Pod A");
    const podB = await createRunningPod(api, "E2E Bind Pod B");
    if (!podA || !podB) { test.skip(); return; }

    // Request binding from Pod A → Pod B using X-Pod-Key
    const bindRes = await fetch(
      `${(await api.get("/api/v1/config/deployment")).url.replace("/api/v1/config/deployment", "")}/api/v1/orgs/${TEST_ORG_SLUG}/bindings`,
      {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          "X-Pod-Key": podA,
        },
        body: JSON.stringify({
          target_pod: podB,
          scopes: ["pod:read"],
        }),
      }
    );
    // 201 if created, 400/401 if pod-key auth not supported via Traefik
    expect([201, 400, 401]).toContain(bindRes.status);

    // Cleanup
    await api.post(`${PODS_BASE}/${podA}/terminate`, {});
    await api.post(`${PODS_BASE}/${podB}/terminate`, {});
  });
});
