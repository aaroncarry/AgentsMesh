package v1_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	v1 "github.com/anthropics/agentsmesh/backend/internal/api/rest/v1"
	orgDomain "github.com/anthropics/agentsmesh/backend/internal/domain/organization"
	"github.com/anthropics/agentsmesh/backend/internal/service/organization"
	"github.com/anthropics/agentsmesh/backend/internal/service/user"
)

// orgRepoStub satisfies orgDomain.Repository; every method panics. Used to
// assert that the IsReserved early-return short-circuits :slug handlers
// before any service/repository call would occur.
type orgRepoStub struct{}

func (orgRepoStub) GetByID(context.Context, int64) (*orgDomain.Organization, error) {
	panic("GetByID called")
}
func (orgRepoStub) GetBySlug(context.Context, string) (*orgDomain.Organization, error) {
	panic("GetBySlug called")
}
func (orgRepoStub) SlugExists(context.Context, string) (bool, error) {
	panic("SlugExists called")
}
func (orgRepoStub) Update(context.Context, int64, map[string]interface{}) error {
	panic("Update called")
}
func (orgRepoStub) ListByUser(context.Context, int64) ([]*orgDomain.Organization, error) {
	panic("ListByUser called")
}
func (orgRepoStub) CreateWithMember(context.Context, *orgDomain.CreateOrgParams) error {
	panic("CreateWithMember called")
}
func (orgRepoStub) DeleteWithCleanup(context.Context, int64) error {
	panic("DeleteWithCleanup called")
}
func (orgRepoStub) CreateMember(context.Context, *orgDomain.Member) error {
	panic("CreateMember called")
}
func (orgRepoStub) GetMember(context.Context, int64, int64) (*orgDomain.Member, error) {
	panic("GetMember called")
}
func (orgRepoStub) DeleteMember(context.Context, int64, int64) error {
	panic("DeleteMember called")
}
func (orgRepoStub) UpdateMemberRole(context.Context, int64, int64, string) error {
	panic("UpdateMemberRole called")
}
func (orgRepoStub) ListMembers(context.Context, int64) ([]*orgDomain.Member, error) {
	panic("ListMembers called")
}
func (orgRepoStub) ListMembersWithUser(context.Context, int64) ([]*orgDomain.Member, error) {
	panic("ListMembersWithUser called")
}
func (orgRepoStub) MemberExists(context.Context, int64, int64) (bool, error) {
	panic("MemberExists called")
}

func setupReservedTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	// Silent recovery: stub-repo panics surface as 500 (assertions require 404,
	// so regression is still detected) without polluting test stderr.
	r.Use(gin.RecoveryWithWriter(io.Discard))

	orgSvc := organization.NewService(orgRepoStub{})
	userSvc := user.NewService(nil)
	handler := v1.NewOrganizationHandler(orgSvc, userSvc)

	g := r.Group("/api/v1/orgs")
	g.Use(func(c *gin.Context) { c.Set("user_id", int64(1)); c.Next() })
	g.GET("/:slug", handler.GetOrganization)
	g.PUT("/:slug", handler.UpdateOrganization)
	g.DELETE("/:slug", handler.DeleteOrganization)
	return r
}

// Reserved slugs MUST short-circuit to 404 before any service call. If the
// IsReserved early-return is broken, orgRepoStub panics; gin.Recovery turns
// that into 500, which the assertion catches (404 ≠ 500).
func TestOrganizationHandler_GetReservedSlugReturns404(t *testing.T) {
	r := setupReservedTestRouter()
	for _, slug := range []string{"admin", "api", "personal", "onboarding"} {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/orgs/"+slug, nil)
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusNotFound, w.Code, "GET /orgs/%s should be 404", slug)
	}
}

func TestOrganizationHandler_DeleteReservedSlugReturns404(t *testing.T) {
	r := setupReservedTestRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/v1/orgs/admin", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestOrganizationHandler_PutReservedSlugReturns404(t *testing.T) {
	r := setupReservedTestRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/api/v1/orgs/admin", nil)
	r.ServeHTTP(w, req)
	// IsReserved must short-circuit BEFORE ShouldBindJSON. If a future
	// refactor moves the early-return below body parsing, an empty PUT body
	// would 400 instead of 404 — fail loud here so that regression cannot
	// pass silently.
	assert.Equal(t, http.StatusNotFound, w.Code,
		"PUT /orgs/admin must 404 from IsReserved early-return, not 400 from body validation")
}
