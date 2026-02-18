package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ─── Mock APIKeyValidator ──────────────────────────────────────────

// mockAPIKeyValidator implements APIKeyValidator for testing
type mockAPIKeyValidator struct {
	result *APIKeyValidateResult
	err    error

	lastUsedCalled bool
	lastUsedID     int64
	lastUsedErr    error
}

func (m *mockAPIKeyValidator) ValidateKey(_ context.Context, _ string) (*APIKeyValidateResult, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.result, nil
}

func (m *mockAPIKeyValidator) UpdateLastUsed(_ context.Context, id int64) error {
	m.lastUsedCalled = true
	m.lastUsedID = id
	return m.lastUsedErr
}

// ─── Helper ────────────────────────────────────────────────────────

// setupAPIKeyMiddlewareTest creates a gin engine with the API key middleware
// mounted on a route that expects :slug.
func setupAPIKeyMiddlewareTest(
	validator APIKeyValidator,
	orgSvc OrganizationService,
) *gin.Engine {
	r := gin.New()
	r.GET("/orgs/:slug/test", APIKeyAuthMiddleware(validator, orgSvc), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	return r
}

// ─── APIKeyAuthMiddleware ──────────────────────────────────────────

func TestAPIKeyAuthMiddleware(t *testing.T) {
	defaultOrg := &mockOrg{id: 1, slug: "my-org", name: "My Org"}
	defaultResult := &APIKeyValidateResult{
		APIKeyID:       42,
		OrganizationID: 1,
		CreatedBy:      10,
		Scopes:         []string{"pods:read", "tickets:write"},
		KeyName:        "test-key",
	}

	t.Run("no API key header returns 401", func(t *testing.T) {
		r := setupAPIKeyMiddlewareTest(
			&mockAPIKeyValidator{result: defaultResult},
			&mockOrgService{org: defaultOrg, isMember: true, role: "member"},
		)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/orgs/my-org/test", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "API key is required")
	})

	t.Run("X-API-Key header — valid key sets TenantContext and APIKeyContext", func(t *testing.T) {
		validator := &mockAPIKeyValidator{result: defaultResult}
		orgSvc := &mockOrgService{org: defaultOrg, isMember: true, role: "member"}

		r := gin.New()

		var capturedTenant *TenantContext
		var capturedAKCtx *APIKeyContext
		var capturedAuthType string

		r.GET("/orgs/:slug/test",
			APIKeyAuthMiddleware(validator, orgSvc),
			func(c *gin.Context) {
				capturedTenant = GetTenant(c)
				capturedAKCtx = GetAPIKeyContext(c)
				if val, exists := c.Get("auth_type"); exists {
					capturedAuthType, _ = val.(string)
				}
				c.JSON(http.StatusOK, gin.H{"ok": true})
			},
		)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/orgs/my-org/test", nil)
		req.Header.Set("X-API-Key", "amk_testkey123")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// TenantContext
		require.NotNil(t, capturedTenant)
		assert.Equal(t, int64(1), capturedTenant.OrganizationID)
		assert.Equal(t, "my-org", capturedTenant.OrganizationSlug)
		assert.Equal(t, int64(10), capturedTenant.UserID)
		assert.Equal(t, "apikey", capturedTenant.UserRole)

		// APIKeyContext
		require.NotNil(t, capturedAKCtx)
		assert.Equal(t, int64(42), capturedAKCtx.APIKeyID)
		assert.Equal(t, "test-key", capturedAKCtx.KeyName)
		assert.ElementsMatch(t, []string{"pods:read", "tickets:write"}, capturedAKCtx.Scopes)

		// auth_type
		assert.Equal(t, "apikey", capturedAuthType)
	})

	t.Run("Authorization Bearer amk_ header works", func(t *testing.T) {
		r := setupAPIKeyMiddlewareTest(
			&mockAPIKeyValidator{result: defaultResult},
			&mockOrgService{org: defaultOrg, isMember: true, role: "member"},
		)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/orgs/my-org/test", nil)
		req.Header.Set("Authorization", "Bearer amk_testkey123")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("invalid key — validator returns ErrAPIKeyNotFound", func(t *testing.T) {
		r := setupAPIKeyMiddlewareTest(
			&mockAPIKeyValidator{err: ErrAPIKeyNotFound},
			&mockOrgService{org: defaultOrg, isMember: true, role: "member"},
		)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/orgs/my-org/test", nil)
		req.Header.Set("X-API-Key", "amk_badkey")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid API key")
	})

	t.Run("disabled key — validator returns ErrAPIKeyDisabled", func(t *testing.T) {
		r := setupAPIKeyMiddlewareTest(
			&mockAPIKeyValidator{err: ErrAPIKeyDisabled},
			&mockOrgService{org: defaultOrg, isMember: true, role: "member"},
		)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/orgs/my-org/test", nil)
		req.Header.Set("X-API-Key", "amk_disabled")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
		assert.Contains(t, w.Body.String(), "API key is disabled")
	})

	t.Run("expired key — validator returns ErrAPIKeyExpired", func(t *testing.T) {
		r := setupAPIKeyMiddlewareTest(
			&mockAPIKeyValidator{err: ErrAPIKeyExpired},
			&mockOrgService{org: defaultOrg, isMember: true, role: "member"},
		)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/orgs/my-org/test", nil)
		req.Header.Set("X-API-Key", "amk_expired")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "API key has expired")
	})

	t.Run("unknown validator error returns 500", func(t *testing.T) {
		r := setupAPIKeyMiddlewareTest(
			&mockAPIKeyValidator{err: errors.New("database connection lost")},
			&mockOrgService{org: defaultOrg, isMember: true, role: "member"},
		)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/orgs/my-org/test", nil)
		req.Header.Set("X-API-Key", "amk_something")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Failed to validate API key")
	})

	t.Run("org slug mismatch returns 403", func(t *testing.T) {
		differentOrgResult := &APIKeyValidateResult{
			APIKeyID:       42,
			OrganizationID: 999, // different from the org looked up by slug
			CreatedBy:      10,
			Scopes:         []string{"pods:read"},
			KeyName:        "test-key",
		}
		r := setupAPIKeyMiddlewareTest(
			&mockAPIKeyValidator{result: differentOrgResult},
			&mockOrgService{org: defaultOrg, isMember: true, role: "member"}, // org.GetID() == 1
		)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/orgs/my-org/test", nil)
		req.Header.Set("X-API-Key", "amk_mismatch")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
		assert.Contains(t, w.Body.String(), "API key does not belong to this organization")
	})

	t.Run("missing slug returns 400", func(t *testing.T) {
		rr := gin.New()
		// Route without :slug parameter
		rr.GET("/test",
			APIKeyAuthMiddleware(
				&mockAPIKeyValidator{result: defaultResult},
				&mockOrgService{org: defaultOrg, isMember: true, role: "member"},
			),
			func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"ok": true})
			},
		)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-API-Key", "amk_test")
		rr.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Organization slug is required")
	})

	t.Run("org not found returns 404", func(t *testing.T) {
		r := setupAPIKeyMiddlewareTest(
			&mockAPIKeyValidator{result: defaultResult},
			&mockOrgService{org: defaultOrg, getErr: errors.New("not found"), isMember: true, role: "member"},
		)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/orgs/nonexistent/test", nil)
		req.Header.Set("X-API-Key", "amk_test")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "Organization not found")
	})
}

// ─── RequireScope ──────────────────────────────────────────────────

func TestRequireScope(t *testing.T) {
	t.Run("has required scope — passes through", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/test", nil)

		c.Set("apikey_context", &APIKeyContext{
			APIKeyID: 1,
			Scopes:   []string{"pods:read", "tickets:write"},
		})

		handler := RequireScope("pods:read")
		handler(c)

		assert.False(t, c.IsAborted())
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("has one of multiple required scopes", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/test", nil)

		c.Set("apikey_context", &APIKeyContext{
			APIKeyID: 1,
			Scopes:   []string{"tickets:write"},
		})

		handler := RequireScope("pods:read", "tickets:write")
		handler(c)

		assert.False(t, c.IsAborted())
	})

	t.Run("missing scope returns 403", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/test", nil)

		c.Set("apikey_context", &APIKeyContext{
			APIKeyID: 1,
			Scopes:   []string{"pods:read"},
		})

		handler := RequireScope("tickets:write")
		handler(c)

		assert.True(t, c.IsAborted())
		assert.Equal(t, http.StatusForbidden, w.Code)
		assert.Contains(t, w.Body.String(), "Insufficient scope")
	})

	t.Run("no APIKeyContext (user auth) passes through", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/test", nil)
		// No apikey_context set — simulating regular user auth

		handler := RequireScope("pods:read")
		handler(c)

		assert.False(t, c.IsAborted())
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("invalid APIKeyContext type returns 500", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/test", nil)
		c.Set("apikey_context", "not-a-valid-context")

		handler := RequireScope("pods:read")
		handler(c)

		assert.True(t, c.IsAborted())
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

// ─── extractAPIKey ─────────────────────────────────────────────────

func TestExtractAPIKey(t *testing.T) {
	t.Run("X-API-Key header", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/test", nil)
		c.Request.Header.Set("X-API-Key", "amk_abc123")

		result := extractAPIKey(c)
		assert.Equal(t, "amk_abc123", result)
	})

	t.Run("Bearer amk_ header", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/test", nil)
		c.Request.Header.Set("Authorization", "Bearer amk_xyz789")

		result := extractAPIKey(c)
		assert.Equal(t, "amk_xyz789", result)
	})

	t.Run("Bearer non-amk_ header returns empty", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/test", nil)
		c.Request.Header.Set("Authorization", "Bearer jwt_token_here")

		result := extractAPIKey(c)
		assert.Equal(t, "", result)
	})

	t.Run("no header returns empty", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/test", nil)

		result := extractAPIKey(c)
		assert.Equal(t, "", result)
	})

	t.Run("X-API-Key takes priority over Authorization", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/test", nil)
		c.Request.Header.Set("X-API-Key", "amk_from_xapikey")
		c.Request.Header.Set("Authorization", "Bearer amk_from_bearer")

		result := extractAPIKey(c)
		assert.Equal(t, "amk_from_xapikey", result)
	})

	t.Run("Authorization without Bearer prefix returns empty", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/test", nil)
		c.Request.Header.Set("Authorization", "amk_nobearer")

		result := extractAPIKey(c)
		assert.Equal(t, "", result)
	})

	t.Run("X-API-Key without amk_ prefix returns empty", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/test", nil)
		c.Request.Header.Set("X-API-Key", "invalid_prefix_key")

		result := extractAPIKey(c)
		assert.Equal(t, "", result)
	})
}

// ─── GetAPIKeyContext ──────────────────────────────────────────────

func TestGetAPIKeyContext(t *testing.T) {
	t.Run("present returns context", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		expected := &APIKeyContext{
			APIKeyID: 42,
			KeyName:  "my-key",
			Scopes:   []string{"pods:read"},
		}
		c.Set("apikey_context", expected)

		result := GetAPIKeyContext(c)
		require.NotNil(t, result)
		assert.Equal(t, int64(42), result.APIKeyID)
		assert.Equal(t, "my-key", result.KeyName)
		assert.Equal(t, []string{"pods:read"}, result.Scopes)
	})

	t.Run("not present returns nil", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		result := GetAPIKeyContext(c)
		assert.Nil(t, result)
	})

	t.Run("wrong type returns nil", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("apikey_context", "not the right type")

		result := GetAPIKeyContext(c)
		assert.Nil(t, result)
	})
}

// ─── handleAPIKeyError ─────────────────────────────────────────────

func TestHandleAPIKeyError(t *testing.T) {
	t.Run("ErrAPIKeyNotFound returns 401", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/test", nil)

		handleAPIKeyError(c, ErrAPIKeyNotFound)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("ErrAPIKeyDisabled returns 403", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/test", nil)

		handleAPIKeyError(c, ErrAPIKeyDisabled)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("ErrAPIKeyExpired returns 401", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/test", nil)

		handleAPIKeyError(c, ErrAPIKeyExpired)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("unknown error returns 500", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/test", nil)

		handleAPIKeyError(c, errors.New("something unexpected"))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}
