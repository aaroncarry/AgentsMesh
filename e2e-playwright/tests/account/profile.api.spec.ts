import { test, expect } from "../../fixtures/index";
import { TEST_USER } from "../../helpers/env";
import { CLEANUP } from "../../helpers/test-data";

test.describe("User Profile API", () => {
  /**
   * TC-PROF-001: Get current user
   * Maps to: e2e/account/profile/TC-PROF-001-get-current-user.yaml
   */
  test("get current user returns user info", async ({ api }) => {
    const res = await api.get("/api/v1/users/me");
    expect(res.status).toBe(200);

    const data = await res.json();
    expect(data.user.email).toBe(TEST_USER.email);
    expect(data.user.id).toBeTruthy();
  });

  test("get current user without auth returns 401", async ({ api }) => {
    const res = await api.getWithToken("/api/v1/users/me", "invalid-token");
    expect(res.status).toBe(401);
  });

  /**
   * TC-PROF-002: Update user name
   * Maps to: e2e/account/profile/TC-PROF-002-update-name.yaml
   */
  test("update user name", async ({ api, db }) => {
    const email = "profile-e2e@test.local";
    const password = "TestPass123!";
    db.cleanup(CLEANUP.userByEmail(email)); // pre-clean

    // Create user via API to avoid hash issues
    await api.postPublic("/api/v1/auth/register", {
      email, username: "profilee2e", password, name: "Original Name",
    });

    await api.loginAs(email, password);
    const res = await api.put("/api/v1/users/me", { name: "Updated Name E2E" });
    expect(res.status).toBe(200);

    const data = await res.json();
    expect(data.user.name).toBe("Updated Name E2E");

    db.cleanup(CLEANUP.userByEmail(email));
  });

  /**
   * TC-PROF-003: Change password
   * Maps to: e2e/account/profile/TC-PROF-003-change-password.yaml
   */
  test("change password with correct current password", async ({ api, db }) => {
    const email = "pwchange-e2e@test.local";
    const password = "TestPass123!";
    db.cleanup(CLEANUP.userByEmail(email)); // pre-clean

    await api.postPublic("/api/v1/auth/register", {
      email, username: "pwchangee2e", password, name: "PW Change User",
    });

    await api.loginAs(email, password);
    const res = await api.post("/api/v1/users/me/password", {
      current_password: password,
      new_password: "NewPassword456!",
    });
    expect(res.status).toBe(200);

    // Verify login with new password
    const loginRes = await api.postPublic("/api/v1/auth/login", {
      email, password: "NewPassword456!",
    });
    expect(loginRes.status).toBe(200);

    db.cleanup(CLEANUP.userByEmail(email));
  });

  /**
   * TC-PROF-004: Change password with wrong current password
   * Maps to: e2e/account/profile/TC-PROF-004-change-password-wrong.yaml
   */
  test("change password with wrong current password fails", async ({ api }) => {
    const res = await api.post("/api/v1/users/me/password", {
      current_password: "wrongpassword",
      new_password: "NewPassword456!",
    });
    expect(res.status).toBe(401);
  });

  /**
   * TC-PROF-005: List organizations
   * Maps to: e2e/account/profile/TC-PROF-005-list-organizations.yaml
   */
  test("list user organizations returns org list", async ({ api }) => {
    const res = await api.get("/api/v1/users/me/organizations");
    expect(res.status).toBe(200);

    const data = await res.json();
    expect(data.organizations).toBeTruthy();
    expect(data.organizations.length).toBeGreaterThan(0);
  });
});
