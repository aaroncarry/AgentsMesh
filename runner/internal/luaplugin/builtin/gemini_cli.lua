-- builtin/gemini_cli.lua
-- Gemini CLI Agent configuration plugin

plugin = {
    name = "gemini-cli",
    version = "1.0.0",
    description = "Gemini CLI Agent configuration",
    supported_agents = {"gemini-cli"},
    executable = "gemini",  -- Check if 'gemini' CLI is available
    order = 50,
    critical = true,
    ui = {
        configurable = true,
        fields = {
            { name = "mcp_enabled", type = "boolean", label = "Enable MCP collaboration", default = true },
            { name = "model", type = "string", label = "Model", default = "gemini-2.0-flash" },
            { name = "approval_mode", type = "select", label = "Approval mode", default = "default",
              options = {
                  { value = "default", label = "Default (requires confirmation)" },
                  { value = "auto_edit", label = "Auto edit" },
                  { value = "yolo", label = "Full auto" },
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

-- MCP configuration: write to project-level .gemini/settings.json
-- Uses shared mcp_utils module for common logic
function setup_mcp(ctx)
    mcp_utils.setup_via_gemini_settings(ctx)
end

function setup_model(ctx)
    if ctx.config.model and ctx.config.model ~= "" then
        ctx.add_args("--model", ctx.config.model)
    end
end

function setup_approval(ctx)
    if ctx.config.approval_mode and ctx.config.approval_mode ~= "default" then
        if ctx.config.approval_mode == "yolo" then
            ctx.add_args("--yolo")
        else
            ctx.add_args("--approval-mode", ctx.config.approval_mode)
        end
    end
end

function teardown(ctx)
    -- Project-level config is cleaned up with sandbox
end
