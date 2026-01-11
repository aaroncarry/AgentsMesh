-- builtin/opencode.lua
-- OpenCode CLI Agent configuration plugin

plugin = {
    name = "opencode",
    version = "1.0.0",
    description = "OpenCode CLI Agent configuration",
    supported_agents = {"opencode"},
    executable = "opencode",  -- Check if 'opencode' CLI is available
    order = 50,
    critical = true,
    ui = {
        configurable = true,
        fields = {
            { name = "mcp_enabled", type = "boolean", label = "Enable MCP collaboration", default = true },
        },
    },
}

function setup(ctx)
    setup_mcp(ctx)
end

-- MCP configuration: write to project-level opencode.json
-- Uses shared mcp_utils module for common logic
function setup_mcp(ctx)
    mcp_utils.setup_via_opencode_config(ctx)
end

function teardown(ctx)
    -- Project-level config is cleaned up with sandbox
end
