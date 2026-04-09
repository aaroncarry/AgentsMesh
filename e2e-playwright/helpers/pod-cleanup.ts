import { getApiBaseUrl, TEST_USER, TEST_ORG_SLUG } from "./env";

/**
 * Terminate ALL active pods via API.
 * Used in test cleanup to prevent quota exhaustion.
 */
export async function terminateAllPods(): Promise<number> {
  const baseUrl = getApiBaseUrl();
  try {
    const loginRes = await fetch(`${baseUrl}/api/v1/auth/login`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ email: TEST_USER.email, password: TEST_USER.password }),
    });
    if (!loginRes.ok) return 0;
    const { token } = await loginRes.json();

    const podsRes = await fetch(
      `${baseUrl}/api/v1/orgs/${TEST_ORG_SLUG}/pods?status=running,initializing`,
      { headers: { Authorization: `Bearer ${token}` } }
    );
    if (!podsRes.ok) return 0;
    const { pods = [] } = await podsRes.json();

    let count = 0;
    for (const pod of pods) {
      if (pod.pod_key) {
        await fetch(
          `${baseUrl}/api/v1/orgs/${TEST_ORG_SLUG}/pods/${pod.pod_key}/terminate`,
          {
            method: "POST",
            headers: {
              "Content-Type": "application/json",
              Authorization: `Bearer ${token}`,
            },
            body: "{}",
          }
        ).catch(() => {});
        count++;
      }
    }
    // Wait for pods to fully terminate and release runner capacity
    if (count > 0) {
      await new Promise((r) => setTimeout(r, 5000));
    }
    return count;
  } catch { return 0; }
}
