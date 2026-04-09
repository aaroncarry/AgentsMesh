import { test, expect } from "../../fixtures/index";
import { TEST_ORG_SLUG } from "../../helpers/env";
import { clearAuthRateLimit } from "../../helpers/redis";

const REPO_ID = "1"; // demo-webapp from seed
const SKILLS_BASE = `/api/v1/orgs/${TEST_ORG_SLUG}/repositories/${REPO_ID}/skills`;
const MARKET = `/api/v1/orgs/${TEST_ORG_SLUG}/market/skills`;

/**
 * Extensions Skills comprehensive tests.
 * Maps to: TC-SKILL-001~007, TC-EXTSET-001~002
 */
test.describe("Extensions Skills", () => {
  test.beforeEach(async () => { clearAuthRateLimit(); });

  /** TC-SKILL-007: List user-installed skills */
  test("list user skills for repository", async ({ api }) => {
    const res = await api.get(`${SKILLS_BASE}?scope=user`);
    expect(res.status).toBe(200);
    const data = await res.json();
    expect(data.skills).toBeDefined();
  });

  /** TC-SKILL-007: List org-installed skills */
  test("list org skills for repository", async ({ api }) => {
    const res = await api.get(`${SKILLS_BASE}?scope=org`);
    expect(res.status).toBe(200);
  });

  /** TC-SKILL-007: List marketplace skills */
  test("marketplace skills endpoint works", async ({ api }) => {
    const res = await api.get(MARKET);
    expect(res.status).toBe(200);
  });

  /** TC-SKILL-001: Skills tab UI display */
  test("extensions settings page shows skills section", async ({ page }) => {
    await page.goto(`/${TEST_ORG_SLUG}/settings?scope=organization&tab=extensions`);
    await page.waitForLoadState("networkidle");
    const body = await page.textContent("body");
    expect(body).toMatch(/skill|extension|扩展|技能/i);
  });

  /** TC-EXTSET-001: Extensions settings page */
  test("extensions page shows registries and templates tabs", async ({ page }) => {
    await page.goto(`/${TEST_ORG_SLUG}/settings?scope=organization&tab=extensions`);
    await page.waitForLoadState("networkidle");
    const body = await page.textContent("body");
    // Should have skill registries or MCP templates section
    expect(body).toMatch(/registr|template|MCP|注册|模板/i);
  });

  /** TC-EXTSET-002: Skill registries management */
  test("skill registries list endpoint works", async ({ api }) => {
    const res = await api.get(`/api/v1/orgs/${TEST_ORG_SLUG}/skill-registries`);
    expect(res.status).toBe(200);
  });
});
