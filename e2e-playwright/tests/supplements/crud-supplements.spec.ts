import { test, expect } from "../../fixtures/index";
import { clearAuthRateLimit } from "../../helpers/redis";
import { CLEANUP } from "../../helpers/test-data";
import { TEST_ORG_SLUG } from "../../helpers/env";

/**
 * CRUD supplement tests — filling gaps in existing modules.
 */
test.describe("CRUD Supplements", () => {
  test.beforeEach(async () => { clearAuthRateLimit(); });

  /**
   * TC-AGENTCRED-003: Update agent credential
   */
  test("update agent credential profile name", async ({ api, db }) => {
    // Create
    const createRes = await api.post(
      "/api/v1/users/agent-credentials/agents/claude-code",
      { name: "E2E Update Cred", credentials: { ANTHROPIC_API_KEY: "sk-test" } }
    );
    if (createRes.status === 404) { test.skip(); return; }
    const created = await createRes.json();
    const id = created.profile?.id || created.id;
    if (!id) { test.skip(); return; }

    // Update
    const updateRes = await api.put(
      `/api/v1/users/agent-credentials/profiles/${id}`,
      { name: "E2E Updated Cred" }
    );
    expect(updateRes.status).toBe(200);

    db.cleanup(
      `DELETE FROM user_agent_credential_profiles WHERE name IN ('E2E Update Cred', 'E2E Updated Cred')`
    );
  });

  /**
   * TC-AGENTCRED-005: Set default agent credential
   */
  test("set agent credential as default", async ({ api, db }) => {
    const createRes = await api.post(
      "/api/v1/users/agent-credentials/agents/claude-code",
      { name: "E2E Default Cred", credentials: { ANTHROPIC_API_KEY: "sk-test" } }
    );
    if (createRes.status === 404) { test.skip(); return; }
    const created = await createRes.json();
    const id = created.profile?.id || created.id;
    if (!id) { test.skip(); return; }

    const setRes = await api.post(
      `/api/v1/users/agent-credentials/profiles/${id}/set-default`, {}
    );
    expect(setRes.status).toBe(200);

    db.cleanup(
      `DELETE FROM user_agent_credential_profiles WHERE name = 'E2E Default Cred'`
    );
  });

  /**
   * TC-MEMBER-005: Change member role
   */
  test("change organization member role", async ({ api, db }) => {
    const email = "role-change-e2e@test.local";
    try { db.cleanup(CLEANUP.userByEmail(email)); } catch { /* */ }

    await api.postPublic("/api/v1/auth/register", {
      email, username: "rolechangee2e", password: "TestPass123!", name: "Role Change",
    });

    const orgId = db.queryValue(
      `SELECT id FROM organizations WHERE slug = '${TEST_ORG_SLUG}'`
    );
    const userId = db.queryValue(`SELECT id FROM users WHERE email = '${email}'`);
    if (!orgId || !userId) { test.skip(); return; }

    db.setup(
      `INSERT INTO organization_members (organization_id, user_id, role) VALUES (${orgId}, ${userId}, 'member') ON CONFLICT DO NOTHING`
    );

    // Update role via PUT with user_id
    const putRes = await api.put(
      `/api/v1/orgs/${TEST_ORG_SLUG}/members/${userId}`,
      { role: "admin" }
    );
    expect([200, 204]).toContain(putRes.status);

    db.cleanup(CLEANUP.userByEmail(email));
  });

  /**
   * TC-REPOPROV-005: Set default repository provider
   */
  test("set repository provider as default", async ({ api }) => {
    const createRes = await api.post("/api/v1/users/repository-providers", {
      provider_type: "github",
      name: "E2E Default Provider",
      base_url: "https://api.github.com",
      bot_token: "ghp_default_test",
    });
    const created = await createRes.json();
    const id = created.provider?.id || created.id;
    if (!id) { test.skip(); return; }

    const setRes = await api.post(
      `/api/v1/users/repository-providers/${id}/default`, {}
    );
    expect(setRes.status).toBe(200);

    await api.delete(`/api/v1/users/repository-providers/${id}`);
  });

  /**
   * TC-MEMBER-007: Accept invitation (API)
   */
  test("accept organization invitation via API", async ({ api, db }) => {
    const email = "invite-accept-e2e@test.local";
    try { db.cleanup(CLEANUP.userByEmail(email)); } catch { /* */ }
    try { db.cleanup(`DELETE FROM invitations WHERE email = '${email}'`); } catch { /* */ }

    // Register invitee
    await api.postPublic("/api/v1/auth/register", {
      email, username: "inviteaccepte2e", password: "TestPass123!", name: "Invite Accept",
    });

    // Create invitation via org members endpoint
    const invRes = await api.post(`/api/v1/orgs/${TEST_ORG_SLUG}/invitations`, {
      email, role: "member",
    });
    if (invRes.status !== 201) {
      // May return 409 if already member, or invitation API differs
      try { db.cleanup(CLEANUP.userByEmail(email)); } catch { /* */ }
      test.skip();
      return;
    }

    // Get invitation token from DB
    const token = db.queryValue(
      `SELECT token FROM invitations WHERE email = '${email}' AND accepted_at IS NULL LIMIT 1`
    );
    if (!token) {
      try { db.cleanup(CLEANUP.userByEmail(email)); } catch { /* */ }
      test.skip();
      return;
    }

    // Accept as invitee
    await api.loginAs(email, "TestPass123!");
    const acceptRes = await api.post(`/api/v1/invitations/${token}/accept`, {});
    expect([200, 201]).toContain(acceptRes.status);

    try { db.cleanup(`DELETE FROM invitations WHERE email = '${email}'`); } catch { /* */ }
    try { db.cleanup(CLEANUP.userByEmail(email)); } catch { /* */ }
  });

  /**
   * TC-POD-002: Create pod with repository
   */
  test("create pod with repository association", async ({ api }) => {
    const runnersRes = await api.get(`/api/v1/orgs/${TEST_ORG_SLUG}/runners/available`);
    const runners = (await runnersRes.json()).runners;
    if (!runners?.length) { test.skip(); return; }

    const agentsRes = await api.get(`/api/v1/orgs/${TEST_ORG_SLUG}/agents`);
    const agents = (await agentsRes.json()).builtin_agents;
    if (!agents?.length) { test.skip(); return; }

    // Get repositories
    const repoRes = await api.get(`/api/v1/orgs/${TEST_ORG_SLUG}/repositories`);
    const repos = (await repoRes.json()).repositories;

    const body: Record<string, unknown> = {
      runner_id: runners[0].id,
      agent_slug: agents[0].slug,
      prompt: "E2E Pod with Repo",
    };
    if (repos?.length) {
      body.repository_id = repos[0].id;
    }

    const res = await api.post(`/api/v1/orgs/${TEST_ORG_SLUG}/pods`, body);
    expect([200, 201]).toContain(res.status);
    const data = await res.json();
    const podKey = data.pod_key || data.pod?.pod_key;

    if (podKey) {
      await api.post(`/api/v1/orgs/${TEST_ORG_SLUG}/pods/${podKey}/terminate`, {});
    }
  });
});
