package runner

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/anthropics/agentmesh/runner/internal/sandbox"
	"github.com/anthropics/agentmesh/runner/internal/terminal"
	"github.com/anthropics/agentmesh/runner/internal/workspace"
)

// PodBuilder builds pods using the Builder pattern.
// It provides a fluent API for configuring and creating pods.
type PodBuilder struct {
	runner *Runner

	// Pod configuration
	podKey           string
	agentType        string
	launchCommand    string
	launchArgs       []string
	envVars          map[string]string
	rows             int
	cols             int
	initialPrompt    string
	repositoryURL    string
	branch           string
	ticketIdentifier string
	useWorktree      bool
	prepScript       string
	prepTimeout      int

	// MCP configuration (legacy - now handled by sandbox plugin)
	mcpEnabled bool
	mcpServers []string

	// Sandbox mode - use new plugin-based sandbox system
	useSandbox   bool
	pluginConfig map[string]interface{}
}

// NewPodBuilder creates a new pod builder.
func NewPodBuilder(runner *Runner) *PodBuilder {
	return &PodBuilder{
		runner:  runner,
		envVars: make(map[string]string),
		rows:    24,
		cols:    80,
	}
}

// WithPodKey sets the pod key.
func (b *PodBuilder) WithPodKey(key string) *PodBuilder {
	b.podKey = key
	return b
}

// WithAgentType sets the agent type.
func (b *PodBuilder) WithAgentType(agentType string) *PodBuilder {
	b.agentType = agentType
	return b
}

// WithLaunchCommand sets the command to launch.
func (b *PodBuilder) WithLaunchCommand(command string, args []string) *PodBuilder {
	b.launchCommand = command
	b.launchArgs = args
	return b
}

// WithEnvVars sets environment variables.
func (b *PodBuilder) WithEnvVars(envVars map[string]string) *PodBuilder {
	for k, v := range envVars {
		b.envVars[k] = v
	}
	return b
}

// WithEnvVar adds a single environment variable.
func (b *PodBuilder) WithEnvVar(key, value string) *PodBuilder {
	b.envVars[key] = value
	return b
}

// WithTerminalSize sets terminal dimensions.
func (b *PodBuilder) WithTerminalSize(rows, cols int) *PodBuilder {
	if rows > 0 {
		b.rows = rows
	}
	if cols > 0 {
		b.cols = cols
	}
	return b
}

// WithInitialPrompt sets the initial prompt to send.
func (b *PodBuilder) WithInitialPrompt(prompt string) *PodBuilder {
	b.initialPrompt = prompt
	return b
}

// WithRepository configures repository URL and branch.
func (b *PodBuilder) WithRepository(url, branch string) *PodBuilder {
	b.repositoryURL = url
	b.branch = branch
	return b
}

// WithWorktree enables worktree mode for the given ticket.
func (b *PodBuilder) WithWorktree(ticketIdentifier string) *PodBuilder {
	b.ticketIdentifier = ticketIdentifier
	b.useWorktree = true
	return b
}

// WithPreparationScript sets a script to run before pod starts.
func (b *PodBuilder) WithPreparationScript(script string, timeoutSeconds int) *PodBuilder {
	b.prepScript = script
	b.prepTimeout = timeoutSeconds
	return b
}

// WithMCP enables MCP with specified servers (legacy - use WithSandbox instead).
func (b *PodBuilder) WithMCP(serverNames ...string) *PodBuilder {
	b.mcpEnabled = true
	b.mcpServers = serverNames
	return b
}

// WithSandbox enables sandbox mode with plugin configuration.
// This is the recommended way to configure pod environments.
func (b *PodBuilder) WithSandbox(pluginConfig map[string]interface{}) *PodBuilder {
	b.useSandbox = true
	b.pluginConfig = pluginConfig
	return b
}

// WithPluginConfig adds or updates plugin configuration.
func (b *PodBuilder) WithPluginConfig(key string, value interface{}) *PodBuilder {
	if b.pluginConfig == nil {
		b.pluginConfig = make(map[string]interface{})
	}
	b.pluginConfig[key] = value
	return b
}

// Build creates the pod.
func (b *PodBuilder) Build(ctx context.Context) (*Pod, error) {
	if b.podKey == "" {
		return nil, fmt.Errorf("pod key is required")
	}

	log.Printf("[pod_builder] Building pod: pod_key=%s, agent=%s, use_sandbox=%v",
		b.podKey, b.agentType, b.useSandbox)

	var workingDir, worktreePath, branchName string
	var sb *sandbox.Sandbox
	var launchArgs []string
	var err error

	// Use sandbox system if enabled and sandbox manager is available
	if b.useSandbox && b.runner.sandboxManager != nil {
		sb, err = b.buildWithSandbox(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to create sandbox: %w", err)
		}
		workingDir = sb.WorkDir
		if wt, ok := sb.Metadata["worktree_path"].(string); ok {
			worktreePath = wt
		}
		if bn, ok := sb.Metadata["branch_name"].(string); ok {
			branchName = bn
		}
		launchArgs = append(b.launchArgs, sb.LaunchArgs...)
	} else {
		// Legacy mode: use old working directory resolution
		workingDir, worktreePath, branchName, err = b.resolveWorkingDirectory(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve working directory: %w", err)
		}

		// Run preparation script if specified (legacy mode)
		if b.prepScript != "" {
			if err := b.runPreparation(ctx, workingDir, worktreePath, branchName); err != nil {
				return nil, fmt.Errorf("preparation failed: %w", err)
			}
		}
		launchArgs = b.launchArgs
	}

	// Merge environment variables
	envVars := b.mergeEnvVars()

	// Add sandbox env vars if available
	if sb != nil {
		for k, v := range sb.EnvVars {
			envVars[k] = v
		}
	}

	// Create terminal
	term, err := terminal.New(terminal.Options{
		Command:  b.launchCommand,
		Args:     launchArgs,
		WorkDir:  workingDir,
		Env:      envVars,
		Rows:     b.rows,
		Cols:     b.cols,
		OnOutput: nil, // Will be set by caller
		OnExit:   nil, // Will be set by caller
	})
	if err != nil {
		// Cleanup sandbox on failure
		if sb != nil && b.runner.sandboxManager != nil {
			b.runner.sandboxManager.Cleanup(b.podKey)
		}
		return nil, fmt.Errorf("failed to create terminal: %w", err)
	}

	// Create pod
	pod := &Pod{
		ID:               b.podKey,
		PodKey:           b.podKey,
		AgentType:        b.agentType,
		RepositoryURL:    b.repositoryURL,
		Branch:           branchName,
		WorktreePath:     worktreePath,
		InitialPrompt:    b.initialPrompt,
		Terminal:         term,
		StartedAt:        time.Now(),
		Status:           PodStatusInitializing,
		TicketIdentifier: b.ticketIdentifier,
	}

	log.Printf("[pod_builder] Pod built: pod_key=%s, working_dir=%s, sandbox=%v",
		b.podKey, workingDir, sb != nil)

	return pod, nil
}

// buildWithSandbox creates a sandbox using the plugin system.
func (b *PodBuilder) buildWithSandbox(ctx context.Context) (*sandbox.Sandbox, error) {
	// Build plugin config from builder fields
	config := make(map[string]interface{})

	// Copy explicit builder fields
	if b.repositoryURL != "" {
		config["repository_url"] = b.repositoryURL
	}
	if b.branch != "" {
		config["branch"] = b.branch
	}
	if b.ticketIdentifier != "" {
		config["ticket_identifier"] = b.ticketIdentifier
	}
	if b.prepScript != "" {
		config["init_script"] = b.prepScript
	}
	if b.prepTimeout > 0 {
		config["init_timeout"] = b.prepTimeout
	}
	if len(b.envVars) > 0 {
		envMap := make(map[string]interface{})
		for k, v := range b.envVars {
			envMap[k] = v
		}
		config["env_vars"] = envMap
	}

	// Merge explicit plugin config (can override above values)
	for k, v := range b.pluginConfig {
		config[k] = v
	}

	// Create sandbox
	return b.runner.sandboxManager.Create(ctx, b.podKey, config)
}

// resolveWorkingDirectory determines the working directory for the pod.
// Returns (workingDir, worktreePath, branchName, error).
// Note: This method is only used in non-sandbox mode. For sandbox mode,
// the working directory is set by the SandboxManager plugins.
func (b *PodBuilder) resolveWorkingDirectory(ctx context.Context) (string, string, string, error) {
	// Priority 1: Use workspace manager with repository URL
	if b.repositoryURL != "" && b.runner.workspace != nil {
		worktreePath, err := b.runner.workspace.CreateWorktree(ctx, b.repositoryURL, b.branch, b.podKey)
		if err != nil {
			return "", "", "", fmt.Errorf("failed to create repository worktree: %w", err)
		}
		return worktreePath, worktreePath, b.branch, nil
	}

	// Priority 2: Use temporary workspace
	if b.runner.workspace != nil {
		tempPath := b.runner.workspace.TempWorkspace(b.podKey)
		return tempPath, "", "", nil
	}

	// Priority 3: Use workspace root from config
	return b.runner.cfg.WorkspaceRoot, "", "", nil
}

// runPreparation runs the preparation script.
func (b *PodBuilder) runPreparation(ctx context.Context, workingDir, worktreePath, branchName string) error {
	preparer := workspace.NewPreparerFromScript(b.prepScript, b.prepTimeout)
	if preparer == nil {
		return nil
	}

	prepCtx := &workspace.PreparationContext{
		PodID:            b.podKey,
		TicketIdentifier: b.ticketIdentifier,
		BranchName:       branchName,
		WorkingDir:       workingDir,
		WorktreeDir:      worktreePath,
		BaseEnvVars:      b.envVars,
	}

	log.Printf("[pod_builder] Running preparation script: pod_key=%s", b.podKey)

	if err := preparer.Prepare(ctx, prepCtx); err != nil {
		return fmt.Errorf("preparation script failed: %w", err)
	}

	log.Printf("[pod_builder] Preparation completed: pod_key=%s", b.podKey)
	return nil
}

// mergeEnvVars merges all environment variable sources.
func (b *PodBuilder) mergeEnvVars() map[string]string {
	result := make(map[string]string)

	// Add config env vars first (lowest priority)
	if b.runner.cfg != nil {
		for k, v := range b.runner.cfg.AgentEnvVars {
			result[k] = v
		}
	}

	// Add builder env vars (highest priority)
	for k, v := range b.envVars {
		result[k] = v
	}

	return result
}

// ExtendedPod adds additional fields to Pod for enhanced functionality.
type ExtendedPod struct {
	*Pod

	// Output/exit callbacks
	OnOutput func([]byte)
	OnExit   func(int)

	// Additional metadata
	TicketIdentifier       string
	ManagedTerminalSession *terminal.Session // Reference to managed PTY terminal session
}

// init initializes extended Pod functionality.
func init() {
	// The Pod struct in runner.go will be extended
}
