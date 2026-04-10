import { test, expect } from "../../fixtures/index";
import { RegisterPage } from "../../pages/register.page";
import { CLEANUP } from "../../helpers/test-data";
import { clearAuthRateLimit } from "../../helpers/redis";

test.describe("Registration Flow", () => {
  test.use({ storageState: { cookies: [], origins: [] } });

  test.beforeEach(async () => {
    clearAuthRateLimit();
  });

  /**
   * TC-REG-004: Register page UI elements
   * Maps to: e2e/account/auth/register/TC-REG-004-ui-flow.yaml
   */
  test("register page displays all required elements", async ({ page }) => {
    const registerPage = new RegisterPage(page);
    await registerPage.goto();

    await expect(registerPage.nameInput).toBeVisible();
    await expect(registerPage.emailInput).toBeVisible();
    await expect(registerPage.usernameInput).toBeVisible();
    await expect(registerPage.passwordInput).toBeVisible();
    await expect(registerPage.submitButton).toBeVisible();
    await expect(registerPage.loginLink).toBeVisible();
  });

  /**
   * TC-REG-001: Successful registration (API)
   * Maps to: e2e/account/auth/register/TC-REG-001-success.yaml
   */
  test("successful registration creates user and returns token", async ({ api, db }) => {
    const email = "newuser-e2e@test.local";
    try { db.cleanup(CLEANUP.userByEmail(email)); } catch { /* ignore */ }

    const res = await api.postPublic("/api/v1/auth/register", {
      email, username: "newusere2e", password: "TestPass123!", name: "E2E Test User",
    });

    expect(res.status).toBe(201);
    const data = await res.json();
    expect(data.token).toBeTruthy();
    expect(data.user.email).toBe(email);

    db.cleanup(CLEANUP.userByEmail(email));
  });

  /**
   * TC-REG-002: Registration fails with existing email (API)
   * Maps to: e2e/account/auth/register/TC-REG-002-email-exists.yaml
   */
  test("registration fails with existing email", async ({ api }) => {
    const res = await api.postPublic("/api/v1/auth/register", {
      email: "dev@agentsmesh.local", username: "anotheruser",
      password: "TestPass123!", name: "Another User",
    });
    expect(res.status).toBe(409);
  });

  /**
   * TC-REG-003: Registration fails with weak password (API)
   * Maps to: e2e/account/auth/register/TC-REG-003-weak-password.yaml
   */
  test("registration fails with weak password", async ({ api }) => {
    const res = await api.postPublic("/api/v1/auth/register", {
      email: "weakpwd@test.local", username: "weakpwduser",
      password: "123", name: "Weak Password User",
    });
    expect(res.status).toBe(400);
  });

  /**
   * TC-REG-004: Register page UI interaction flow
   * Maps to: e2e/account/auth/register/TC-REG-004-ui-flow.yaml
   */
  test("register page UI flow", async ({ page, db }) => {
    const email = "uiregister@test.local";
    try { db.cleanup(CLEANUP.userAndOrgsByEmail(email)); } catch { /* ignore */ }

    const registerPage = new RegisterPage(page);
    await registerPage.goto();
    await registerPage.register({
      name: "UI Register Test", email,
      username: "uiregistertest", password: "TestPass123!",
      confirmPassword: "TestPass123!",
    });

    await page.waitForURL((url) => !url.pathname.includes("/register"), {
      timeout: 15_000,
    });

    try { db.cleanup(CLEANUP.userAndOrgsByEmail(email)); } catch { /* ignore */ }
  });
});
