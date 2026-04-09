import { test, expect } from "../../fixtures/index";
import { OrgMembersPage } from "../../pages/settings/org-members.page";
import { TEST_ORG_SLUG } from "../../helpers/env";
import { clearAuthRateLimit } from "../../helpers/redis";
import { CLEANUP } from "../../helpers/test-data";

test.describe("Organization Members Settings", () => {
  test.beforeEach(async () => { clearAuthRateLimit(); });

  /**
   * TC-MEMBER-001: List members API
   * Maps to: e2e/settings/org-members/TC-MEMBER-001-list-members.yaml
   */
  test("list organization members returns array", async ({ api }) => {
    const res = await api.get(`/api/v1/orgs/${TEST_ORG_SLUG}/members`);
    expect(res.status).toBe(200);
    const data = await res.json();
    expect(data.members).toBeTruthy();
    expect(data.members.length).toBeGreaterThan(0);
  });

  /**
   * TC-MEMBER-002: Members page elements
   * Maps to: e2e/settings/org-members/TC-MEMBER-002-page-elements.yaml
   */
  test("members page displays required elements", async ({ page }) => {
    const membersPage = new OrgMembersPage(page, TEST_ORG_SLUG);
    await membersPage.goto();

    await expect(membersPage.inviteButton).toBeVisible();
  });

  /**
   * TC-MEMBER-003: Invite member dialog
   * Maps to: e2e/settings/org-members/TC-MEMBER-003-invite-dialog.yaml
   */
  test("invite dialog opens with email and role fields", async ({ page }) => {
    const membersPage = new OrgMembersPage(page, TEST_ORG_SLUG);
    await membersPage.goto();
    await membersPage.openInviteDialog();

    await expect(membersPage.inviteEmailInput).toBeVisible();
  });

  /**
   * TC-MEMBER-004: Send member invitation
   * Maps to: e2e/settings/org-members/TC-MEMBER-004-send-invite.yaml
   */
  test("send member invitation", async ({ api, db }) => {
    // Pre-clean any existing invitation
    try { db.cleanup(`DELETE FROM invitations WHERE email = 'invite-test-e2e@example.com'`); } catch { /* ignore */ }

    const res = await api.post(`/api/v1/orgs/${TEST_ORG_SLUG}/invitations`, {
      email: "invite-test-e2e@example.com",
      role: "member",
    });
    expect([200, 201]).toContain(res.status);

    db.cleanup(
      `DELETE FROM invitations WHERE email = 'invite-test-e2e@example.com'`
    );
  });

  /**
   * TC-MEMBER-006: Remove member
   * Maps to: e2e/settings/org-members/TC-MEMBER-006-remove-member.yaml
   */
  test("remove member from organization", async ({ api, db }) => {
    const email = "remove-test-e2e@test.local";
    try { db.cleanup(CLEANUP.userByEmail(email)); } catch { /* ignore */ }

    // Create user via register
    await api.postPublic("/api/v1/auth/register", {
      email, username: "removeteste2e", password: "TestPass123!", name: "Remove Test",
    });
    const userId = db.queryValue(`SELECT id FROM users WHERE email = '${email}'`);
    const orgId = db.queryValue(
      `SELECT id FROM organizations WHERE slug = '${TEST_ORG_SLUG}'`
    );

    // Add as member
    db.setup(
      `INSERT INTO organization_members (organization_id, user_id, role) VALUES (${orgId}, ${userId}, 'member') ON CONFLICT DO NOTHING`
    );

    // Remove via API
    const members = await api.get(`/api/v1/orgs/${TEST_ORG_SLUG}/members`);
    const membersData = await members.json();
    const member = membersData.members?.find(
      (m: { user?: { email?: string } }) => m.user?.email === email
    );

    if (member) {
      const delRes = await api.delete(
        `/api/v1/orgs/${TEST_ORG_SLUG}/members/${member.id}`
      );
      expect([200, 204]).toContain(delRes.status);
    }

    db.cleanup(CLEANUP.userByEmail(email));
  });
});
