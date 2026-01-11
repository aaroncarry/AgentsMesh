-- builtin/codex_cli.lua
-- OpenAI Codex CLI Agent configuration plugin

plugin = {
    name = "codex-cli",
    version = "1.0.0",
    description = "OpenAI Codex CLI Agent configuration",
    supported_agents = {"codex-cli"},
    executable = "codex",  -- Check if 'codex' CLI is available
    order = 50,
    critical = true,
    ui = {
        configurable = true,
        fields = {
            { name = "mcp_enabled", type = "boolean", label = "Enable MCP collaboration", default = true },
            { name = "model", type = "string", label = "Model", default = "o3" },
            { name = "approval_mode", type = "select", label = "Approval mode", default = "suggest",
              options = {
                  { value = "suggest", label = "Suggest only" },
                  { value = "auto-edit", label = "Auto edit" },
                  { value = "full-auto", label = "Full auto" },
              }
            },
        },
    },
}

function setup(ctx)
    setup_mcp(ctx)
    setup_model(ctx)
    setup_approval(ctx)
end

-- MCP configuration: via -c command line parameter (no file needed)
-- Uses shared mcp_utils module for common logic
function setup_mcp(ctx)
    mcp_utils.setup_via_cli_args(ctx)
end

function setup_model(ctx)
    if ctx.config.model and ctx.config.model ~= "" then
        ctx.add_args("-m", ctx.config.model)
    end
end

function setup_approval(ctx)
    if ctx.config.approval_mode == "full-auto" then
        ctx.add_args("--full-auto")
    elseif ctx.config.approval_mode and ctx.config.approval_mode ~= "" then
        ctx.add_args("--approval-mode", ctx.config.approval_mode)
    end
end

function teardown(ctx)
    -- No cleanup needed, config passed via command line arguments
end
