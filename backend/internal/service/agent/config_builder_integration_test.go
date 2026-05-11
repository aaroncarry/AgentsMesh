package agent

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/anthropics/agentsmesh/backend/internal/domain/extension"
	"github.com/anthropics/agentsmesh/backend/internal/infra"
	extensionservice "github.com/anthropics/agentsmesh/backend/internal/service/extension"
	"github.com/anthropics/agentsmesh/backend/internal/testkit"
	runnerv1 "github.com/anthropics/agentsmesh/proto/gen/go/runner/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// integrationProvider wraps the testutil DB and creates a ConfigBuilder provider.
type integrationProvider struct {
	db *gorm.DB
}

func newIntegrationProvider(t *testing.T) (*integrationProvider, *gorm.DB) {
	t.Helper()
	db := testkit.SetupTestDB(t)
	return &integrationProvider{db: db}, db
}

func (ip *integrationProvider) configProvider() AgentConfigProvider {
	agentSvc := NewAgentService(infra.NewAgentRepository(ip.db))
	credentialSvc := NewCredentialProfileService(
		infra.NewCredentialProfileRepository(ip.db),
		agentSvc,
		testEncryptor(),
	)
	return &testCompositeProvider{
		agentSvc:      agentSvc,
		credentialSvc: credentialSvc,
	}
}

func seedAgent(t *testing.T, db *gorm.DB, slug, agentfileSrc string) {
	t.Helper()
	testkit.CreateAgent(t, db, slug, "Agent "+slug, agentfileSrc)
}

type testExtensionProvider struct {
	mcpServers []*extension.InstalledMcpServer
	skills     []*extensionservice.ResolvedSkill
}

func (p *testExtensionProvider) GetEffectiveMcpServers(context.Context, int64, int64, int64, string) ([]*extension.InstalledMcpServer, error) {
	return p.mcpServers, nil
}

func (p *testExtensionProvider) GetEffectiveSkills(context.Context, int64, int64, int64, string) ([]*extensionservice.ResolvedSkill, error) {
	return p.skills, nil
}

const factoryDroidAgentFile = `# === Identity ===
AGENT droid
EXECUTABLE droid

# === Mode ===
MODE pty
MODE acp "exec" "--output-format" "acp"

# === Configuration ===
CONFIG autonomy_level SELECT("", "off", "low", "medium", "high") = ""
CONFIG interaction_mode SELECT("", "auto", "spec") = ""
CONFIG skip_permissions_unsafe BOOL = false

# === Environment ===
ENV FACTORY_API_KEY SECRET OPTIONAL
ENV FACTORY_BASE_URL TEXT OPTIONAL

# === Prompt ===
PROMPT_POSITION append

# === Capabilities ===
MCP ON
SKILLS am-delegate, am-channel

# === Build Logic ===
arg "--resume" when config.resume_enabled and mode != "acp"
arg "--skip-permissions-unsafe" when config.skip_permissions_unsafe and mode == "acp"

factory_settings_path = sandbox.root + "/factory-runtime-settings.json"

if config.autonomy_level != "" and config.interaction_mode != "" {
  file factory_settings_path json({ sessionDefaultSettings: { autonomyLevel: config.autonomy_level, interactionMode: config.interaction_mode } })
  arg "--settings" factory_settings_path
}
if config.autonomy_level != "" and config.interaction_mode == "" {
  file factory_settings_path json({ sessionDefaultSettings: { autonomyLevel: config.autonomy_level } })
  arg "--settings" factory_settings_path
}
if config.autonomy_level == "" and config.interaction_mode != "" {
  file factory_settings_path json({ sessionDefaultSettings: { interactionMode: config.interaction_mode } })
  arg "--settings" factory_settings_path
}

if mcp.enabled {
  mkdir sandbox.work_dir + "/.factory"
  file sandbox.work_dir + "/.factory/mcp.json" json({ mcpServers: mcp.servers })
}
`

func TestConfigBuilder_BuildBasicCommand(t *testing.T) {
	ip, db := newIntegrationProvider(t)

	agentfile := "AGENT test-agent\nEXECUTABLE test-agent\nMODE pty"
	seedAgent(t, db, "test-agent", agentfile)

	builder := NewConfigBuilder(ip.configProvider())

	cmd, err := builder.BuildPodCommand(context.Background(), &ConfigBuildRequest{
		AgentSlug:             "test-agent",
		PodKey:                "pod-basic-1",
		MergedAgentfileSource: agentfile,
		MCPPort:               19000,
		Cols:                  120,
		Rows:                  40,
	})
	require.NoError(t, err)
	require.NotNil(t, cmd)

	assert.Equal(t, "pod-basic-1", cmd.PodKey)
	assert.Equal(t, "test-agent", cmd.LaunchCommand)
	assert.Equal(t, "pty", cmd.InteractionMode)
	assert.Equal(t, int32(120), cmd.Cols)
	assert.Equal(t, int32(40), cmd.Rows)
	assert.Nil(t, cmd.Credentials)
	assert.Nil(t, cmd.SandboxConfig)
}

func TestConfigBuilder_WithCredentials(t *testing.T) {
	ip, db := newIntegrationProvider(t)

	agentfile := "AGENT cred-agent\nEXECUTABLE cred-agent\nMODE pty\nENV API_KEY SECRET"
	seedAgent(t, db, "cred-agent", agentfile)

	userID := testkit.CreateUser(t, db, "cred-user@test.com", "creduser")

	agentSvc := NewAgentService(infra.NewAgentRepository(db))
	credSvc := NewCredentialProfileService(
		infra.NewCredentialProfileRepository(db),
		agentSvc,
		testEncryptor(),
	)
	profile, err := credSvc.CreateCredentialProfile(context.Background(), userID, &CreateCredentialProfileParams{
		AgentSlug:   "cred-agent",
		Name:        "test-profile",
		IsDefault:   true,
		Credentials: map[string]string{"API_KEY": "sk-test-123"},
	})
	require.NoError(t, err)
	require.NotNil(t, profile)

	builder := NewConfigBuilder(ip.configProvider())

	cmd, err := builder.BuildPodCommand(context.Background(), &ConfigBuildRequest{
		AgentSlug:             "cred-agent",
		UserID:                userID,
		PodKey:                "pod-cred-1",
		MergedAgentfileSource: agentfile,
		MCPPort:               19000,
		Cols:                  80,
		Rows:                  24,
	})
	require.NoError(t, err)
	require.NotNil(t, cmd)

	// Credentials injected into env_vars via AgentFile eval (ENV API_KEY SECRET)
	require.NotNil(t, cmd.Credentials)
	assert.Contains(t, cmd.Credentials, "API_KEY")
}

func TestConfigBuilder_EvalProducesCorrectOutput(t *testing.T) {
	ip, db := newIntegrationProvider(t)

	agentfile := `AGENT merge-agent
EXECUTABLE merge-agent
MODE acp
CONFIG model STRING = "opus"
arg "--model" config.model when config.model != ""
PROMPT_POSITION prepend
`
	seedAgent(t, db, "merge-agent", agentfile)

	builder := NewConfigBuilder(ip.configProvider())

	cmd, err := builder.BuildPodCommand(context.Background(), &ConfigBuildRequest{
		AgentSlug:             "merge-agent",
		PodKey:                "pod-eval-1",
		MergedAgentfileSource: agentfile,
		MCPPort:               19000,
		Cols:                  80,
		Rows:                  24,
	})
	require.NoError(t, err)
	require.NotNil(t, cmd)

	// Eval produces launch_command from AGENT declaration
	assert.Equal(t, "merge-agent", cmd.LaunchCommand)
	// Eval produces interaction_mode from MODE declaration
	assert.Equal(t, "acp", cmd.InteractionMode)
	// Eval produces launch_args from arg statements (config.model = "opus")
	assert.Contains(t, cmd.LaunchArgs, "--model")
	assert.Contains(t, cmd.LaunchArgs, "opus")
	// Eval produces prompt_position from PROMPT_POSITION declaration
	assert.Equal(t, "prepend", cmd.PromptPosition)

	// Empty MergedAgentfileSource → error (orchestrator must always resolve it)
	seedAgent(t, db, "resume-agent", "AGENT resume-agent\nEXECUTABLE resume-agent\nMODE pty")
	_, err = builder.BuildPodCommand(context.Background(), &ConfigBuildRequest{
		AgentSlug: "resume-agent",
		PodKey:    "pod-resume-1",
		MCPPort:   19000,
		Cols:      80,
		Rows:      24,
	})
	require.Error(t, err, "empty MergedAgentfileSource should error")
}

func TestConfigBuilder_FactoryDroidMaterializesMCPAndSkills(t *testing.T) {
	ip, db := newIntegrationProvider(t)
	seedAgent(t, db, "factory-cli", factoryDroidAgentFile)

	repoID := int64(42)
	skillSha := strings.Repeat("a", 64)
	builder := NewConfigBuilder(ip.configProvider())
	builder.SetExtensionProvider(&testExtensionProvider{
		skills: []*extensionservice.ResolvedSkill{
			{
				Slug:        "repo-skill",
				ContentSha:  skillSha,
				DownloadURL: "https://cdn.example.test/repo-skill.tar.gz",
				PackageSize: 1234,
				TargetDir:   "skills/repo-skill",
			},
		},
	})

	cmd, err := builder.BuildPodCommand(context.Background(), &ConfigBuildRequest{
		AgentSlug:             "factory-cli",
		OrganizationID:        1,
		UserID:                2,
		RepositoryID:          &repoID,
		PodKey:                "pod-factory-1",
		MergedAgentfileSource: factoryDroidAgentFile,
		Prompt:                "hello",
		MCPPort:               19000,
		Cols:                  100,
		Rows:                  30,
	})
	require.NoError(t, err)
	require.NotNil(t, cmd)

	assert.Equal(t, "droid", cmd.LaunchCommand)
	assert.Equal(t, "pty", cmd.InteractionMode)
	assert.Equal(t, "append", cmd.PromptPosition)
	assert.Equal(t, "hello", cmd.Prompt)
	assert.Empty(t, cmd.LaunchArgs)

	mcpFile := requireFileToCreate(t, cmd, PlaceholderWorkDir+"/.factory/mcp.json")
	var mcpConfig map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(mcpFile.Content), &mcpConfig))
	servers, ok := mcpConfig["mcpServers"].(map[string]interface{})
	require.True(t, ok)
	agentsMesh, ok := servers["agentsmesh"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "http", agentsMesh["type"])
	assert.Equal(t, "http://127.0.0.1:19000/mcp", agentsMesh["url"])
	headers, ok := agentsMesh["headers"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "pod-factory-1", headers["X-Pod-Key"])

	delegateSkill := requireFileToCreate(t, cmd, PlaceholderWorkDir+"/.factory/skills/am-delegate/SKILL.md")
	assert.Contains(t, delegateSkill.Content, "name: am-delegate")
	channelSkill := requireFileToCreate(t, cmd, PlaceholderWorkDir+"/.factory/skills/am-channel/SKILL.md")
	assert.Contains(t, channelSkill.Content, "name: am-channel")

	require.Len(t, cmd.ResourcesToDownload, 1)
	res := cmd.ResourcesToDownload[0]
	assert.Equal(t, skillSha, res.Sha)
	assert.Equal(t, "https://cdn.example.test/repo-skill.tar.gz", res.DownloadUrl)
	assert.Equal(t, "{{.sandbox.work_dir}}/.factory/skills/repo-skill", res.TargetPath)
	assert.Equal(t, "skill_package", res.ResourceType)
	assert.Equal(t, int64(1234), res.SizeBytes)
}

func TestConfigBuilder_FactoryDroidRuntimeSettings(t *testing.T) {
	ip, db := newIntegrationProvider(t)
	seedAgent(t, db, "factory-cli", factoryDroidAgentFile)

	source := factoryDroidAgentFileWithRuntimeConfig("medium", "spec")
	builder := NewConfigBuilder(ip.configProvider())
	cmd, err := builder.BuildPodCommand(context.Background(), &ConfigBuildRequest{
		AgentSlug:             "factory-cli",
		PodKey:                "pod-factory-settings",
		MergedAgentfileSource: source,
		MCPPort:               19000,
		Cols:                  100,
		Rows:                  30,
	})
	require.NoError(t, err)
	require.NotNil(t, cmd)

	settingsPath := PlaceholderSandboxRoot + "/factory-runtime-settings.json"
	assert.Equal(t, []string{"--settings", settingsPath}, cmd.LaunchArgs)

	settingsFile := requireFileToCreate(t, cmd, settingsPath)
	var settings map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(settingsFile.Content), &settings))
	sessionDefaults, ok := settings["sessionDefaultSettings"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "medium", sessionDefaults["autonomyLevel"])
	assert.Equal(t, "spec", sessionDefaults["interactionMode"])
}

func TestConfigBuilder_FactoryDroidACPModeUsesExecACP(t *testing.T) {
	ip, db := newIntegrationProvider(t)
	seedAgent(t, db, "factory-cli", factoryDroidAgentFile)

	source := factoryDroidAgentFile + "\nMODE acp\n"
	builder := NewConfigBuilder(ip.configProvider())
	cmd, err := builder.BuildPodCommand(context.Background(), &ConfigBuildRequest{
		AgentSlug:             "factory-cli",
		PodKey:                "pod-factory-acp",
		MergedAgentfileSource: source,
		MCPPort:               19000,
		Cols:                  100,
		Rows:                  30,
	})
	require.NoError(t, err)
	require.NotNil(t, cmd)

	assert.Equal(t, "droid", cmd.LaunchCommand)
	assert.Equal(t, "acp", cmd.InteractionMode)
	assert.Equal(t, []string{"exec", "--output-format", "acp"}, cmd.LaunchArgs)
}

func TestConfigBuilder_FactoryDroidACPModeUsesSkipPermissionsUnsafe(t *testing.T) {
	ip, db := newIntegrationProvider(t)
	seedAgent(t, db, "factory-cli", factoryDroidAgentFile)

	source := factoryDroidAgentFileWithUnsafeSkipPermissions() + "\nMODE acp\n"
	builder := NewConfigBuilder(ip.configProvider())
	cmd, err := builder.BuildPodCommand(context.Background(), &ConfigBuildRequest{
		AgentSlug:             "factory-cli",
		PodKey:                "pod-factory-acp-unsafe",
		MergedAgentfileSource: source,
		MCPPort:               19000,
		Cols:                  100,
		Rows:                  30,
	})
	require.NoError(t, err)
	require.NotNil(t, cmd)

	assert.Equal(t, "acp", cmd.InteractionMode)
	assert.Equal(t, []string{"exec", "--output-format", "acp", "--skip-permissions-unsafe"}, cmd.LaunchArgs)
}

func TestConfigBuilder_FactoryDroidPTYModeIgnoresACPSkipPermissionsUnsafe(t *testing.T) {
	ip, db := newIntegrationProvider(t)
	seedAgent(t, db, "factory-cli", factoryDroidAgentFile)

	builder := NewConfigBuilder(ip.configProvider())
	cmd, err := builder.BuildPodCommand(context.Background(), &ConfigBuildRequest{
		AgentSlug:             "factory-cli",
		PodKey:                "pod-factory-pty-unsafe",
		MergedAgentfileSource: factoryDroidAgentFileWithUnsafeSkipPermissions(),
		MCPPort:               19000,
		Cols:                  100,
		Rows:                  30,
	})
	require.NoError(t, err)
	require.NotNil(t, cmd)

	assert.Equal(t, "pty", cmd.InteractionMode)
	assert.NotContains(t, cmd.LaunchArgs, "--skip-permissions-unsafe")
}

func TestConfigBuilder_FactoryDroidResumeUsesInteractiveResume(t *testing.T) {
	ip, db := newIntegrationProvider(t)
	seedAgent(t, db, "factory-cli", factoryDroidAgentFile)

	resumeAgentFile := `CONFIG resume_enabled BOOL = true
CONFIG resume_session STRING = "agentsmesh-session-id"
` + factoryDroidAgentFileWithRuntimeConfig("high", "auto")

	builder := NewConfigBuilder(ip.configProvider())
	cmd, err := builder.BuildPodCommand(context.Background(), &ConfigBuildRequest{
		AgentSlug:             "factory-cli",
		PodKey:                "pod-factory-resume",
		MergedAgentfileSource: resumeAgentFile,
		MCPPort:               19000,
		Cols:                  100,
		Rows:                  30,
	})
	require.NoError(t, err)
	require.NotNil(t, cmd)

	assert.Equal(t, "droid", cmd.LaunchCommand)
	settingsPath := PlaceholderSandboxRoot + "/factory-runtime-settings.json"
	assert.Equal(t, []string{"--resume", "--settings", settingsPath}, cmd.LaunchArgs)

	settingsFile := requireFileToCreate(t, cmd, settingsPath)
	var settings map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(settingsFile.Content), &settings))
	sessionDefaults, ok := settings["sessionDefaultSettings"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "high", sessionDefaults["autonomyLevel"])
	assert.Equal(t, "auto", sessionDefaults["interactionMode"])
}

func TestConfigBuilder_FactoryDroidACPModeDoesNotUseInteractiveResumeArg(t *testing.T) {
	ip, db := newIntegrationProvider(t)
	seedAgent(t, db, "factory-cli", factoryDroidAgentFile)

	resumeACPAgentFile := `CONFIG resume_enabled BOOL = true
CONFIG resume_session STRING = "agentsmesh-session-id"
` + factoryDroidAgentFileWithRuntimeConfig("high", "auto") + "\nMODE acp\n"

	builder := NewConfigBuilder(ip.configProvider())
	cmd, err := builder.BuildPodCommand(context.Background(), &ConfigBuildRequest{
		AgentSlug:             "factory-cli",
		PodKey:                "pod-factory-acp-resume",
		MergedAgentfileSource: resumeACPAgentFile,
		MCPPort:               19000,
		Cols:                  100,
		Rows:                  30,
	})
	require.NoError(t, err)
	require.NotNil(t, cmd)

	settingsPath := PlaceholderSandboxRoot + "/factory-runtime-settings.json"
	assert.Equal(t, "acp", cmd.InteractionMode)
	assert.Equal(t, []string{"exec", "--output-format", "acp", "--settings", settingsPath}, cmd.LaunchArgs)
	assert.NotContains(t, cmd.LaunchArgs, "--resume")
}

func TestConfigBuilder_FactoryDroidLegacySlugMaterializesSkills(t *testing.T) {
	ip, db := newIntegrationProvider(t)
	seedAgent(t, db, "factory-droid", factoryDroidAgentFile)

	builder := NewConfigBuilder(ip.configProvider())
	cmd, err := builder.BuildPodCommand(context.Background(), &ConfigBuildRequest{
		AgentSlug:             "factory-droid",
		PodKey:                "pod-factory-legacy",
		MergedAgentfileSource: factoryDroidAgentFile,
		Prompt:                "hello",
		MCPPort:               19000,
		Cols:                  100,
		Rows:                  30,
	})
	require.NoError(t, err)
	require.NotNil(t, cmd)

	delegateSkill := requireFileToCreate(t, cmd, PlaceholderWorkDir+"/.factory/skills/am-delegate/SKILL.md")
	assert.Contains(t, delegateSkill.Content, "name: am-delegate")
	channelSkill := requireFileToCreate(t, cmd, PlaceholderWorkDir+"/.factory/skills/am-channel/SKILL.md")
	assert.Contains(t, channelSkill.Content, "name: am-channel")
}

func requireFileToCreate(t *testing.T, cmd *runnerv1.CreatePodCommand, path string) *runnerv1.FileToCreate {
	t.Helper()
	for _, f := range cmd.FilesToCreate {
		if f.Path == path {
			return f
		}
	}
	require.Failf(t, "file not found", "expected file path %s", path)
	return nil
}

func factoryDroidAgentFileWithRuntimeConfig(autonomyLevel, interactionMode string) string {
	src := factoryDroidAgentFile
	src = strings.Replace(src,
		`CONFIG autonomy_level SELECT("", "off", "low", "medium", "high") = ""`,
		`CONFIG autonomy_level SELECT("", "off", "low", "medium", "high") = "`+autonomyLevel+`"`,
		1,
	)
	src = strings.Replace(src,
		`CONFIG interaction_mode SELECT("", "auto", "spec") = ""`,
		`CONFIG interaction_mode SELECT("", "auto", "spec") = "`+interactionMode+`"`,
		1,
	)
	return src
}

func factoryDroidAgentFileWithUnsafeSkipPermissions() string {
	return strings.Replace(
		factoryDroidAgentFile,
		`CONFIG skip_permissions_unsafe BOOL = false`,
		`CONFIG skip_permissions_unsafe BOOL = true`,
		1,
	)
}
