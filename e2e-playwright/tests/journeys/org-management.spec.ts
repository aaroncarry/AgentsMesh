import { test, expect } from "../../fixtures/index";
import { TEST_ORG_SLUG } from "../../helpers/env";
import { clearAuthRateLimit } from "../../helpers/redis";
import { CLEANUP } from "../../helpers/test-data";

/**
 * Journey: Organization Management Lifecycle
 * Create org → Invite member → Accept → Change role → Remove → Cleanup
 */
test.describe("Journey: Organization Management", () => {
  test.beforeEach(async () => { clearAuthRateLimit(); });

  const MEMBER_EMAIL = "org-journey-member@test.local";

  test("full org member lifecycle: invite → accept → role → remove", async ({ api, db }) => {
    // ── Pre-clean ──
    try { db.cleanup(`DELETE FROM invitations WHERE email = '${MEMBER_EMAIL}'`); } catch { /* */ }
    try { db.cleanup(CLEANUP.userByEmail(MEMBER_EMAIL)); } catch { /* */ }

    // ── Step 1: Register a user to be invited ──
    const regRes = await api.postPublic("/api/v1/auth/register", {
      email: MEMBER_EMAIL,
      username: "orgjourneymbr",
      password: "JourneyPass123!",
      name: "Org Journey Member",
    });
    expect([200, 201]).toContain(regRes.status);

    const userId = db.queryValue(
      `SELECT id FROM users WHERE email = '${MEMBER_EMAIL}'`
    );
    expect(userId).toBeTruthy();

    // ── Step 2: Invite member to org ──
    const invRes = await api.post(
      `/api/v1/orgs/${TEST_ORG_SLUG}/invitations`,
      { email: MEMBER_EMAIL, role: "member" }
    );
    expect([200, 201]).toContain(invRes.status);

    // ── Step 3: Get invitation token from DB ──
    const invToken = db.queryValue(
      `SELECT token FROM invitations WHERE email = '${MEMBER_EMAIL}' AND accepted_at IS NULL LIMIT 1`
    );
    expect(invToken).toBeTruthy();

    // ── Step 4: Accept invitation as the invited user ──
    await api.loginAs(MEMBER_EMAIL, "JourneyPass123!");
    const acceptRes = await api.post(
      `/api/v1/invitations/${invToken}/accept`, {}
    );
    expect([200, 201]).toContain(acceptRes.status);

    // ── Step 5: Verify member appears in org member list ──
    await api.login(); // back to dev user
    const membersRes = await api.get(
      `/api/v1/orgs/${TEST_ORG_SLUG}/members`
    );
    const members = (await membersRes.json()).members || [];
    const member = members.find(
      (m: { user?: { email?: string } }) => m.user?.email === MEMBER_EMAIL
    );
    expect(member).toBeTruthy();
    expect(member.role).toBe("member");

    // ── Step 6: Change member role to admin ──
    const roleRes = await api.put(
      `/api/v1/orgs/${TEST_ORG_SLUG}/members/${userId}`,
      { role: "admin" }
    );
    expect([200, 204]).toContain(roleRes.status);

    // Verify role changed
    const updatedRes = await api.get(
      `/api/v1/orgs/${TEST_ORG_SLUG}/members`
    );
    const updated = (await updatedRes.json()).members || [];
    const updatedMember = updated.find(
      (m: { user?: { email?: string } }) => m.user?.email === MEMBER_EMAIL
    );
    expect(updatedMember?.role).toBe("admin");

    // ── Step 7: Remove member from org ──
    const removeRes = await api.delete(
      `/api/v1/orgs/${TEST_ORG_SLUG}/members/${userId}`
    );
    expect([200, 204]).toContain(removeRes.status);

    // ── Step 8: Verify member no longer in list ──
    const finalRes = await api.get(
      `/api/v1/orgs/${TEST_ORG_SLUG}/members`
    );
    const finalMembers = (await finalRes.json()).members || [];
    const gone = finalMembers.find(
      (m: { user?: { email?: string } }) => m.user?.email === MEMBER_EMAIL
    );
    expect(gone).toBeFalsy();

    // ── Cleanup ──
    try { db.cleanup(`DELETE FROM invitations WHERE email = '${MEMBER_EMAIL}'`); } catch { /* */ }
    try { db.cleanup(CLEANUP.userByEmail(MEMBER_EMAIL)); } catch { /* */ }
  });

  test("org settings visible in UI", async ({ page }) => {
    // Verify org general settings
    await page.goto(
      `/${TEST_ORG_SLUG}/settings?scope=organization&tab=general`
    );
    await page.waitForLoadState("networkidle");
    const body = await page.textContent("body");
    expect(body).toMatch(/Dev Organization|dev-org|organization|组织/i);

    // Verify members page
    await page.goto(
      `/${TEST_ORG_SLUG}/settings?scope=organization&tab=members`
    );
    await page.waitForLoadState("networkidle");
    const membersBody = await page.textContent("body");
    expect(membersBody).toMatch(/member|成员|invite|邀请/i);
  });
});
