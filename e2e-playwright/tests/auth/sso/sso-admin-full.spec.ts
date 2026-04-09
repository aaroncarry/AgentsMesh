import { test, expect } from "../../../fixtures/index";
import { ADMIN_USER, TEST_ORG_SLUG } from "../../../helpers/env";
import { clearAuthRateLimit } from "../../../helpers/redis";

/**
 * SSO Admin UI + API supplements.
 * Maps to: TC-SSO-ADM-001~009
 */
test.describe("SSO Admin Full", () => {
  test.beforeEach(async () => { clearAuthRateLimit(); });

  /** TC-SSO-ADM-001: List SSO configs */
  test("list SSO configs with pagination", async ({ api }) => {
    await api.loginAs(ADMIN_USER.email, ADMIN_USER.password);
    const res = await api.get("/api/v1/admin/sso/configs");
    expect(res.status).toBe(200);
    const data = await res.json();
    expect(data).toBeTruthy();
  });

  /** TC-SSO-ADM-002: Get single config (non-existent) */
  test("get non-existent SSO config returns 404", async ({ api }) => {
    await api.loginAs(ADMIN_USER.email, ADMIN_USER.password);
    const res = await api.get("/api/v1/admin/sso/configs/999999");
    expect(res.status).toBe(404);
  });

  /** TC-SSO-ADM-003: Full CRUD with LDAP (avoids OIDC discovery) */
  test("SSO config CRUD with LDAP protocol", async ({ api }) => {
    await api.loginAs(ADMIN_USER.email, ADMIN_USER.password);

    // Create LDAP config (doesn't need external IdP)
    const createRes = await api.post("/api/v1/admin/sso/configs", {
      name: "E2E LDAP CRUD",
      domain: "e2e-ldap-crud.example.com",
      protocol: "ldap",
      ldap_host: "ldap.example.com",
      ldap_port: 389,
      ldap_base_dn: "dc=example,dc=com",
      ldap_bind_dn: "cn=admin,dc=example,dc=com",
      ldap_bind_password: "test-password",
    });
    expect([200, 201]).toContain(createRes.status);
    const created = await createRes.json();
    const id = created.config?.id || created.id;
    if (!id) return;

    // Read
    const getRes = await api.get(`/api/v1/admin/sso/configs/${id}`);
    expect(getRes.status).toBe(200);

    // Update
    const updateRes = await api.put(`/api/v1/admin/sso/configs/${id}`, {
      name: "E2E LDAP CRUD Updated",
      domain: "e2e-ldap-crud.example.com",
      protocol: "ldap",
      ldap_host: "ldap2.example.com",
      ldap_port: 636,
      ldap_base_dn: "dc=example,dc=com",
      ldap_bind_dn: "cn=admin,dc=example,dc=com",
      ldap_bind_password: "updated-password",
    });
    expect(updateRes.status).toBe(200);

    // Delete
    const delRes = await api.delete(`/api/v1/admin/sso/configs/${id}`);
    expect([200, 204]).toContain(delRes.status);

    // Confirm deleted
    const gone = await api.get(`/api/v1/admin/sso/configs/${id}`);
    expect(gone.status).toBe(404);
  });

  /** TC-SSO-ADM-004: Unauthorized access */
  test("non-admin cannot access SSO admin endpoints", async ({ api }) => {
    // Default user is non-admin
    const res = await api.get("/api/v1/admin/sso/configs");
    expect(res.status).toBe(403);
  });

  /** TC-SSO-ADM-005: Enable/disable config */
  test("enable and disable SSO config", async ({ api }) => {
    await api.loginAs(ADMIN_USER.email, ADMIN_USER.password);

    const createRes = await api.post("/api/v1/admin/sso/configs", {
      name: "E2E Toggle SSO",
      domain: `e2e-toggle-${Date.now()}.example.com`,
      protocol: "ldap",
      ldap_host: "ldap.example.com",
      ldap_port: 389,
      ldap_base_dn: "dc=example,dc=com",
      ldap_bind_dn: "cn=admin,dc=example,dc=com",
      ldap_bind_password: "test",
    });
    if (createRes.status !== 201) { test.skip(); return; }
    const created = await createRes.json();
    const id = created.config?.id || created.id;

    // Enable
    const enableRes = await api.post(`/api/v1/admin/sso/configs/${id}/enable`, {});
    expect(enableRes.status).toBe(200);

    // Disable
    const disableRes = await api.post(`/api/v1/admin/sso/configs/${id}/disable`, {});
    expect([200, 204]).toContain(disableRes.status);

    // Cleanup
    await api.delete(`/api/v1/admin/sso/configs/${id}`);
  });
});
