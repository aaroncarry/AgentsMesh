package agent

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetConfigSchema_CredentialFields(t *testing.T) {
	db := setupConfigBuilderTestDB(t)

	db.Exec(`INSERT INTO agents (slug, name, launch_command, is_builtin, is_active, agentfile_source)
		VALUES ('claude-code', 'Claude Code', 'claude', 1, 1, 'AGENT claude
EXECUTABLE claude
CONFIG model SELECT("", "sonnet", "opus") = ""
ENV ANTHROPIC_API_KEY SECRET OPTIONAL
ENV ANTHROPIC_AUTH_TOKEN SECRET OPTIONAL
ENV ANTHROPIC_BASE_URL TEXT OPTIONAL
ENV FIXED_VALUE = "hello"
MCP ON')`)

	provider := createTestProvider(db)
	builder := NewConfigBuilder(provider)

	schema, err := builder.GetConfigSchema(context.Background(), "claude-code")
	require.NoError(t, err)
	require.NotNil(t, schema)

	// CONFIG fields
	require.Len(t, schema.Fields, 1)
	assert.Equal(t, "model", schema.Fields[0].Name)

	// Credential fields: only ENV with SECRET/TEXT source, not fixed-value ENV
	require.Len(t, schema.CredentialFields, 3)

	assert.Equal(t, "ANTHROPIC_API_KEY", schema.CredentialFields[0].Name)
	assert.Equal(t, "secret", schema.CredentialFields[0].Type)
	assert.True(t, schema.CredentialFields[0].Optional)

	assert.Equal(t, "ANTHROPIC_AUTH_TOKEN", schema.CredentialFields[1].Name)
	assert.Equal(t, "secret", schema.CredentialFields[1].Type)

	assert.Equal(t, "ANTHROPIC_BASE_URL", schema.CredentialFields[2].Name)
	assert.Equal(t, "text", schema.CredentialFields[2].Type)
}

func TestGetConfigSchema_NoCredentialFields(t *testing.T) {
	db := setupConfigBuilderTestDB(t)

	db.Exec(`INSERT INTO agents (slug, name, launch_command, is_builtin, is_active, agentfile_source)
		VALUES ('opencode', 'OpenCode', 'opencode', 1, 1, 'AGENT opencode
EXECUTABLE opencode
MODE pty')`)

	provider := createTestProvider(db)
	builder := NewConfigBuilder(provider)

	schema, err := builder.GetConfigSchema(context.Background(), "opencode")
	require.NoError(t, err)
	require.NotNil(t, schema)

	assert.Empty(t, schema.CredentialFields)
}

func TestGetConfigSchema_NoAgentfile(t *testing.T) {
	db := setupConfigBuilderTestDB(t)

	db.Exec(`INSERT INTO agents (slug, name, launch_command, is_builtin, is_active)
		VALUES ('custom', 'Custom Agent', 'custom', 0, 1)`)

	provider := createTestProvider(db)
	builder := NewConfigBuilder(provider)

	schema, err := builder.GetConfigSchema(context.Background(), "custom")
	require.NoError(t, err)
	require.NotNil(t, schema)

	assert.Empty(t, schema.Fields)
	assert.Empty(t, schema.CredentialFields)
}
