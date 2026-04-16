package v1

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/anthropics/agentsmesh/backend/internal/infra"
	"github.com/anthropics/agentsmesh/backend/internal/middleware"
	runnerservice "github.com/anthropics/agentsmesh/backend/internal/service/runner"
	"github.com/anthropics/agentsmesh/backend/internal/testkit"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func init() {
	gin.SetMode(gin.TestMode)
}

type mockUpgradeCommandSender struct {
	connected   map[int64]bool
	sendErr     error
	sentCalls   []upgradeSendCall
}

type upgradeSendCall struct {
	RunnerID, RequestID, TargetVersion string
	Force                              bool
}

func (m *mockUpgradeCommandSender) IsConnected(runnerID int64) bool {
	return m.connected[runnerID]
}

func (m *mockUpgradeCommandSender) SendUpgradeRunner(runnerID int64, requestID, targetVersion string, force bool) error {
	m.sentCalls = append(m.sentCalls, upgradeSendCall{
		RunnerID:      fmt.Sprintf("%d", runnerID),
		RequestID:     requestID,
		TargetVersion: targetVersion,
		Force:         force,
	})
	return m.sendErr
}

type upgradeTestEnv struct {
	handler  *RunnerHandler
	sender   *mockUpgradeCommandSender
	router   *gin.Engine
	runnerID int64
}

func setupUpgradeTest(t *testing.T) *upgradeTestEnv {
	t.Helper()

	db := testkit.SetupTestDB(t)
	db.Exec(`INSERT INTO organizations (id, name, slug) VALUES (1, 'Test Org', 'test-org')`)
	db.Exec(`INSERT INTO runners (id, organization_id, node_id, status, current_pods, max_concurrent_pods, is_enabled, visibility, registered_by_user_id)
		VALUES (10, 1, 'node-10', 'online', 3, 10, 1, 'organization', 100)`)

	repo := infra.NewRunnerRepository(db)
	svc := runnerservice.NewService(repo)

	sender := &mockUpgradeCommandSender{connected: map[int64]bool{10: true}}
	handler := NewRunnerHandler(svc, WithUpgradeCommandSender(sender))

	router := gin.New()
	router.POST("/runners/:id/upgrade", func(c *gin.Context) {
		c.Set("tenant", &middleware.TenantContext{
			OrganizationID: 1, UserID: 100, UserRole: "owner",
		})
		handler.UpgradeRunner(c)
	})

	return &upgradeTestEnv{handler: handler, sender: sender, router: router, runnerID: 10}
}

func TestUpgradeRunner_Success(t *testing.T) {
	env := setupUpgradeTest(t)
	body := `{"target_version":"1.2.3"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/runners/10/upgrade", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)

	if w.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d: %s", w.Code, w.Body.String())
	}
	if len(env.sender.sentCalls) != 1 {
		t.Fatalf("expected 1 send call, got %d", len(env.sender.sentCalls))
	}
	if env.sender.sentCalls[0].TargetVersion != "1.2.3" {
		t.Errorf("expected target_version=1.2.3, got %q", env.sender.sentCalls[0].TargetVersion)
	}
	if !env.sender.sentCalls[0].Force {
		t.Error("force should always be true")
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp["request_id"] == nil || resp["request_id"] == "" {
		t.Error("response should contain a non-empty request_id")
	}
}

func TestUpgradeRunner_EmptyBody(t *testing.T) {
	env := setupUpgradeTest(t)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/runners/10/upgrade", nil)
	env.router.ServeHTTP(w, req)

	if w.Code != http.StatusAccepted {
		t.Fatalf("expected 202 with empty body, got %d: %s", w.Code, w.Body.String())
	}
	if env.sender.sentCalls[0].TargetVersion != "" {
		t.Errorf("expected empty target_version, got %q", env.sender.sentCalls[0].TargetVersion)
	}
}

func TestUpgradeRunner_ServiceNotConfigured(t *testing.T) {
	db := testkit.SetupTestDB(t)
	svc := runnerservice.NewService(infra.NewRunnerRepository(db))
	handler := NewRunnerHandler(svc) // no upgrade sender

	router := gin.New()
	router.POST("/runners/:id/upgrade", func(c *gin.Context) {
		c.Set("tenant", &middleware.TenantContext{OrganizationID: 1, UserID: 1, UserRole: "owner"})
		handler.UpgradeRunner(c)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/runners/10/upgrade", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503, got %d", w.Code)
	}
}

func TestUpgradeRunner_InvalidID(t *testing.T) {
	env := setupUpgradeTest(t)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/runners/abc/upgrade", nil)
	env.router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestUpgradeRunner_NotFound(t *testing.T) {
	env := setupUpgradeTest(t)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/runners/999/upgrade", nil)
	env.router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
}

func TestUpgradeRunner_Forbidden(t *testing.T) {
	env := setupUpgradeTest(t)

	router := gin.New()
	router.POST("/runners/:id/upgrade", func(c *gin.Context) {
		c.Set("tenant", &middleware.TenantContext{
			OrganizationID: 1, UserID: 999, UserRole: "member",
		})
		env.handler.UpgradeRunner(c)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/runners/10/upgrade", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d: %s", w.Code, w.Body.String())
	}
}

func TestUpgradeRunner_Offline(t *testing.T) {
	env := setupUpgradeTest(t)
	env.sender.connected[10] = false

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/runners/10/upgrade", nil)
	env.router.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503, got %d", w.Code)
	}
}

func TestUpgradeRunner_SendGRPCNotFound(t *testing.T) {
	env := setupUpgradeTest(t)
	env.sender.sendErr = status.Error(codes.NotFound, "stream gone")

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/runners/10/upgrade", nil)
	env.router.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503 for gRPC NotFound, got %d", w.Code)
	}
}

func TestUpgradeRunner_SendInternalError(t *testing.T) {
	env := setupUpgradeTest(t)
	env.sender.sendErr = fmt.Errorf("internal error")

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/runners/10/upgrade", nil)
	env.router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}
