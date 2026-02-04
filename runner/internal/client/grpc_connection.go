// Package client provides gRPC connection management for Runner.
package client

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	runnerv1 "github.com/anthropics/agentsmesh/proto/gen/go/runner/v1"
	"github.com/anthropics/agentsmesh/runner/internal/logger"
	"golang.org/x/time/rate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/tls/certprovider"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// GRPCProtocolVersion is the current gRPC protocol version.
const GRPCProtocolVersion = 2

// GRPCConnection manages the gRPC connection to the server with mTLS.
// Responsibilities: mTLS setup, bidirectional streaming, reconnection, message routing.
type GRPCConnection struct {
	// Connection configuration
	endpoint  string
	serverURL string // HTTP server URL for REST API calls (certificate renewal)
	nodeID    string
	orgSlug   string

	// mTLS certificate paths
	certFile string
	keyFile  string
	caFile   string

	// gRPC components
	conn   *grpc.ClientConn
	creds  credentials.TransportCredentials                                            // advancedtls credentials for hot-reload
	client runnerv1.RunnerServiceClient                                                // gRPC service client
	stream grpc.BidiStreamingClient[runnerv1.RunnerMessage, runnerv1.ServerMessage] // Bidirectional stream
	mu     sync.Mutex

	// Certificate providers for cleanup (prevent goroutine leaks)
	identityProvider certprovider.Provider
	rootProvider     certprovider.Provider

	// Message handling
	handler MessageHandler

	// Reconnection strategy
	reconnectStrategy *ReconnectStrategy

	// Heartbeat
	heartbeatInterval time.Duration

	// Initialization
	initTimeout     time.Duration
	initialized     bool
	availableAgents []string
	initResultCh    chan *runnerv1.InitializeResult

	// Runner info
	runnerVersion string
	mcpPort       int

	// Lifecycle - Priority-based channels for message sending
	// Control messages (heartbeat, pod events, OSC) have higher priority than agent status
	controlCh     chan *runnerv1.RunnerMessage // High priority: heartbeat, pod_created, pod_terminated, OSC, etc.
	terminalCh    chan *runnerv1.RunnerMessage // Low priority: agent_status (terminal output via Relay)
	stopCh        chan struct{}
	stopOnce      sync.Once
	reconnectOnce sync.Once     // Ensures only one reconnection attempt
	reconnectCh   chan struct{} // Signal to trigger reconnection

	// Stuck detection for writeLoop
	lastSendTime atomic.Int64

	// Rate limiting for terminal output (bytes per second)
	// Default: 100KB/s to avoid overwhelming slow server connections
	terminalRateLimiter *rate.Limiter
	terminalRateLimit   int // bytes per second

	// Certificate renewal
	certRenewalCheckInterval time.Duration
	certExpiryWarningDays    int
	certRenewalDays          int // Days before expiry to trigger renewal (default 30)
	certUrgentDays           int // Days before expiry for urgent reconnection (default 7)
}

// GRPCConnectionOption is a functional option for GRPCConnection.
type GRPCConnectionOption func(*GRPCConnection)

// WithGRPCHeartbeatInterval sets the heartbeat interval.
func WithGRPCHeartbeatInterval(d time.Duration) GRPCConnectionOption {
	return func(c *GRPCConnection) {
		c.heartbeatInterval = d
	}
}

// WithGRPCInitTimeout sets the initialization timeout.
func WithGRPCInitTimeout(d time.Duration) GRPCConnectionOption {
	return func(c *GRPCConnection) {
		c.initTimeout = d
	}
}

// WithGRPCRunnerVersion sets the runner version.
func WithGRPCRunnerVersion(version string) GRPCConnectionOption {
	return func(c *GRPCConnection) {
		c.runnerVersion = version
	}
}

// WithGRPCMCPPort sets the MCP port.
func WithGRPCMCPPort(port int) GRPCConnectionOption {
	return func(c *GRPCConnection) {
		c.mcpPort = port
	}
}

// WithGRPCTerminalRateLimit sets the terminal output rate limit in bytes per second.
// Default is 100KB/s. Set to 0 to disable rate limiting.
// Recommended: Set to ~80% of server upload bandwidth to leave room for control messages.
func WithGRPCTerminalRateLimit(bytesPerSecond int) GRPCConnectionOption {
	return func(c *GRPCConnection) {
		c.terminalRateLimit = bytesPerSecond
	}
}

// WithGRPCServerURL sets the HTTP server URL for REST API calls.
func WithGRPCServerURL(serverURL string) GRPCConnectionOption {
	return func(c *GRPCConnection) {
		c.serverURL = serverURL
	}
}

// WithGRPCCertRenewalDays sets the days before expiry to trigger renewal.
func WithGRPCCertRenewalDays(days int) GRPCConnectionOption {
	return func(c *GRPCConnection) {
		c.certRenewalDays = days
	}
}

// WithGRPCCertUrgentDays sets the days before expiry for urgent reconnection.
func WithGRPCCertUrgentDays(days int) GRPCConnectionOption {
	return func(c *GRPCConnection) {
		c.certUrgentDays = days
	}
}

// NewGRPCConnection creates a new gRPC connection with mTLS.
func NewGRPCConnection(endpoint, nodeID, orgSlug, certFile, keyFile, caFile string, opts ...GRPCConnectionOption) *GRPCConnection {
	c := &GRPCConnection{
		endpoint:                 endpoint,
		nodeID:                   nodeID,
		orgSlug:                  orgSlug,
		certFile:                 certFile,
		keyFile:                  keyFile,
		caFile:                   caFile,
		heartbeatInterval:        30 * time.Second,
		initTimeout:              30 * time.Second,
		reconnectStrategy:        NewReconnectStrategy(5*time.Second, 5*time.Minute),
		controlCh:                make(chan *runnerv1.RunnerMessage, 100),  // Small buffer for control messages
		terminalCh:               make(chan *runnerv1.RunnerMessage, 1000), // Large buffer for terminal output
		stopCh:                   make(chan struct{}),
		reconnectCh:              make(chan struct{}, 1),
		initResultCh:             make(chan *runnerv1.InitializeResult, 1),
		runnerVersion:            "1.0.0",
		mcpPort:                  19000,
		certRenewalCheckInterval: 24 * time.Hour,
		certExpiryWarningDays:    30,
		certRenewalDays:          30, // Renew 30 days before expiry
		certUrgentDays:           7,  // Urgent reconnection 7 days before expiry
		terminalRateLimit:        50 * 1024, // Default: 50KB/s (conservative for shared bandwidth)
	}

	for _, opt := range opts {
		opt(c)
	}

	// Initialize rate limiter if rate limit is set
	if c.terminalRateLimit > 0 {
		// rate.Limit is tokens per second, burst allows one maxSize message
		c.terminalRateLimiter = rate.NewLimiter(rate.Limit(c.terminalRateLimit), c.terminalRateLimit)
		logger.GRPC().Info("Terminal output rate limiting enabled",
			"rate_limit", fmt.Sprintf("%dKB/s", c.terminalRateLimit/1024))
	}

	return c
}

// SetHandler sets the message handler.
func (c *GRPCConnection) SetHandler(handler MessageHandler) {
	c.handler = handler
}

// SetOrgSlug sets the organization slug.
func (c *GRPCConnection) SetOrgSlug(orgSlug string) {
	c.mu.Lock()
	c.orgSlug = orgSlug
	c.mu.Unlock()
}

// GetOrgSlug returns the organization slug.
func (c *GRPCConnection) GetOrgSlug() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.orgSlug
}

// IsInitialized returns whether the connection has completed initialization.
func (c *GRPCConnection) IsInitialized() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.initialized
}

// GetAvailableAgents returns the list of available agents.
func (c *GRPCConnection) GetAvailableAgents() []string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.availableAgents
}

// Connect establishes a gRPC connection with mTLS using advancedtls for certificate hot-reloading.
func (c *GRPCConnection) Connect() error {
	// Close existing connection if any (important for reconnection)
	// This prevents resource leaks and TLS session conflicts
	c.mu.Lock()
	if c.conn != nil {
		logger.GRPC().Debug("Closing existing gRPC connection before reconnect")
		c.conn.Close()
		c.conn = nil
	}
	c.mu.Unlock()

	// Parse endpoint to extract host:port (remove scheme like grpcs://)
	dialTarget, err := parseGRPCEndpoint(c.endpoint)
	if err != nil {
		return fmt.Errorf("failed to parse gRPC endpoint: %w", err)
	}

	// Create advancedtls credentials with file-based certificate reloading
	creds, err := c.createAdvancedTLSCredentials()
	if err != nil {
		return fmt.Errorf("failed to create TLS credentials: %w", err)
	}

	// gRPC dial options
	dialOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(creds),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                30 * time.Second,
			Timeout:             10 * time.Second,
			PermitWithoutStream: true,
		}),
	}

	// Connect to server
	conn, err := grpc.Dial(dialTarget, dialOpts...)
	if err != nil {
		return fmt.Errorf("failed to dial gRPC server: %w", err)
	}

	// Create gRPC service client
	client := runnerv1.NewRunnerServiceClient(conn)

	c.mu.Lock()
	c.conn = conn
	c.client = client
	c.creds = creds
	c.initialized = false
	c.mu.Unlock()

	logger.GRPC().Info("Connected to server with advancedtls", "endpoint", c.endpoint, "org", c.orgSlug)
	return nil
}

// Start starts the connection management loop.
func (c *GRPCConnection) Start() {
	go c.connectionLoop()
}

// Stop stops the connection and releases resources.
func (c *GRPCConnection) Stop() {
	c.stopOnce.Do(func() {
		close(c.stopCh)
		c.mu.Lock()
		if c.conn != nil {
			c.conn.Close()
			c.conn = nil
		}
		// Close certificate providers to prevent goroutine leaks
		if c.identityProvider != nil {
			c.identityProvider.Close()
			c.identityProvider = nil
		}
		if c.rootProvider != nil {
			c.rootProvider.Close()
			c.rootProvider = nil
		}
		c.mu.Unlock()
	})
}

// connectionLoop manages the connection lifecycle with auto-reconnection.
func (c *GRPCConnection) connectionLoop() {
	for {
		select {
		case <-c.stopCh:
			logger.GRPC().Info("Connection loop stopped")
			return
		default:
		}

		// Try to connect
		if err := c.Connect(); err != nil {
			delay := c.reconnectStrategy.NextDelay()
			logger.GRPC().Warn("Failed to connect, will retry",
				"attempt", c.reconnectStrategy.AttemptCount(),
				"error", err,
				"retry_in", delay)

			select {
			case <-c.stopCh:
				return
			case <-time.After(delay):
			}
			continue
		}

		// Reset reconnect strategy on successful connection
		c.reconnectStrategy.Reset()

		// Run the connection (blocks until disconnected)
		c.runConnection()

		// Check if we should stop
		select {
		case <-c.stopCh:
			return
		default:
		}

		// Wait before reconnecting
		logger.GRPC().Info("Connection closed, will attempt to reconnect")
		select {
		case <-c.stopCh:
			return
		case <-time.After(c.reconnectStrategy.CurrentInterval()):
		}
	}
}

// runConnection establishes the bidirectional stream and handles messages.
func (c *GRPCConnection) runConnection() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Clear terminal queue before establishing new connection
	// Old terminal output is stale after reconnection and would:
	// 1. Delay initialization by flooding the new stream
	// 2. Potentially cause immediate timeout if backend is slow
	// TUI frames are expendable - users will see fresh output after reconnection
	c.drainTerminalQueue()

	// Add org_slug to metadata for organization routing
	ctx = metadata.AppendToOutgoingContext(ctx, "x-org-slug", c.orgSlug)

	logger.GRPC().Debug("Establishing bidirectional stream", "org", c.orgSlug)

	// Create bidirectional stream
	stream, err := c.client.Connect(ctx)
	if err != nil {
		logger.GRPC().Error("Failed to establish stream", "error", err)
		return
	}

	c.mu.Lock()
	c.stream = stream
	c.mu.Unlock()

	logger.GRPC().Info("Bidirectional stream established")

	done := make(chan struct{})
	readLoopDone := make(chan struct{}) // Signal when readLoop exits

	// Start write loop
	go c.writeLoop(ctx, done)

	// IMPORTANT: Start read loop BEFORE initialization
	// The read loop must be running to receive the initialize_result response
	go c.readLoop(ctx, readLoopDone)

	// Perform initialization (blocks until handshake completes or times out)
	if err := c.performInitialization(ctx); err != nil {
		logger.GRPC().Error("Initialization failed", "error", err)
		close(done)
		return
	}

	// Start heartbeat loop (only after successful initialization)
	go c.heartbeatLoop(ctx, done)

	// Start certificate renewal checker
	go c.certRenewalChecker(ctx, done)

	// Monitor for reconnection signal (certificate renewal)
	go func() {
		select {
		case <-c.reconnectCh:
			logger.GRPC().Info("Reconnection requested for certificate renewal")
			cancel() // Cancel context to trigger reconnection
		case <-done:
			return
		case <-c.stopCh:
			return
		}
	}()

	// Wait for context cancellation, stop signal, or readLoop exit
	select {
	case <-ctx.Done():
		logger.GRPC().Debug("Context cancelled, closing connection")
	case <-c.stopCh:
		logger.GRPC().Debug("Stop signal received, closing connection")
	case <-readLoopDone:
		logger.GRPC().Debug("Read loop exited, closing connection")
	}

	// Clear stream to prevent sending to disconnected stream
	// This ensures sendTerminal/sendControl will reject new messages during reconnect
	c.mu.Lock()
	c.stream = nil
	c.mu.Unlock()

	// Signal other goroutines to stop
	close(done)
}

// ==================== gRPC Error Handling ====================

// isRetryableError returns true if the gRPC error is retryable.
func isRetryableError(err error) bool {
	st, ok := status.FromError(err)
	if !ok {
		return false
	}

	switch st.Code() {
	case codes.Unavailable, codes.ResourceExhausted, codes.Aborted:
		return true
	default:
		return false
	}
}

// Ensure GRPCConnection implements Connection interface.
var _ Connection = (*GRPCConnection)(nil)

// ==================== Helper Functions ====================

// parseGRPCEndpoint parses a gRPC endpoint URL and returns the host:port for dialing.
// Supports formats:
//   - grpcs://host:port -> host:port (TLS)
//   - grpc://host:port  -> host:port (plain)
//   - host:port         -> host:port (as-is)
func parseGRPCEndpoint(endpoint string) (string, error) {
	log := logger.GRPC()

	// If it doesn't contain a scheme, assume it's already host:port
	if !strings.Contains(endpoint, "://") {
		return endpoint, nil
	}

	// Parse as URL
	u, err := url.Parse(endpoint)
	if err != nil {
		log.Error("Invalid endpoint URL", "endpoint", endpoint, "error", err)
		return "", err
	}

	// Validate scheme
	switch u.Scheme {
	case "grpc", "grpcs":
		// Valid gRPC schemes
	default:
		log.Error("Unsupported gRPC scheme", "scheme", u.Scheme, "endpoint", endpoint)
		return "", fmt.Errorf("unsupported scheme %q", u.Scheme)
	}

	// Return host:port
	if u.Host == "" {
		log.Error("Missing host in endpoint URL", "endpoint", endpoint)
		return "", fmt.Errorf("missing host in endpoint")
	}

	log.Debug("Parsed gRPC endpoint", "endpoint", endpoint, "dial_target", u.Host)
	return u.Host, nil
}
