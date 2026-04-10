import { test, expect } from "../../fixtures/index";
import { OrgGeneralPage } from "../../pages/settings/org-general.page";
import { TEST_ORG_SLUG } from "../../helpers/env";

test.describe("Organization General Settings", () => {
  let orgPage: OrgGeneralPage;

  test.beforeEach(async ({ page }) => {
    orgPage = new OrgGeneralPage(page, TEST_ORG_SLUG);
    await orgPage.goto();
  });

  /**
   * TC-ORGSET-001: Get org details (API)
   * Maps to: e2e/settings/org-general/TC-ORGSET-001-get-org.yaml
   */
  test("get org details via API", async ({ api }) => {
    const res = await api.get(`/api/v1/orgs/${TEST_ORG_SLUG}`);
    expect(res.status).toBe(200);
    const data = await res.json();
    expect(data.organization.slug).toBe(TEST_ORG_SLUG);
    expect(data.organization.name).toBeTruthy();
  });

  /**
   * TC-ORGSET-002: Settings page elements
   * Maps to: e2e/settings/org-general/TC-ORGSET-002-page-elements.yaml
   */
  test("settings page displays required elements", async () => {
    await expect(orgPage.nameInput).toBeVisible();
    await expect(orgPage.slugInput).toBeVisible();
    await expect(orgPage.saveButton).toBeVisible();
  });

  /**
   * TC-ORGSET-004: Slug is disabled
   * Maps to: e2e/settings/org-general/TC-ORGSET-004-slug-disabled.yaml
   */
  test("org slug field is disabled", async () => {
    await expect(orgPage.slugInput).toBeDisabled();
    const slugValue = await orgPage.slugInput.inputValue();
    expect(slugValue).toBe(TEST_ORG_SLUG);
  });

  /**
   * TC-ORGSET-003: Update org name
   * Maps to: e2e/settings/org-general/TC-ORGSET-003-update-name.yaml
   */
  test("update org name and verify persistence", async ({ page, db }) => {
    // Save original name
    const originalName = await orgPage.nameInput.inputValue();

    // Update
    await orgPage.updateName("E2E Test Organization Name");
    await page.waitForTimeout(1000);

    // Refresh and verify
    await page.reload();
    await orgPage.nameInput.waitFor({ state: "visible" });
    const updatedValue = await orgPage.nameInput.inputValue();
    expect(updatedValue).toBe("E2E Test Organization Name");

    // Restore original
    db.cleanup(
      `UPDATE organizations SET name = '${originalName}' WHERE slug = '${TEST_ORG_SLUG}'`
    );
  });
});
