-- builtin/claude_code.lua
-- Claude Code Agent configuration plugin

plugin = {
    name = "claude-code",
    version = "1.0.0",
    description = "Claude Code Agent configuration",
    supported_agents = {"claude-code"},
    executable = "claude",  -- Check if 'claude' CLI is available
    order = 50,
    critical = true,
    ui = {
        configurable = true,
        fields = {
            { name = "mcp_enabled", type = "boolean", label = "Enable MCP collaboration", default = true },
            { name = "skills_enabled", type = "boolean", label = "Enable collaboration skills", default = true },
            { name = "model", type = "select", label = "Model", default = "opus",
              options = {
                  { value = "opus", label = "Claude Opus" },
                  { value = "sonnet", label = "Claude Sonnet" },
              }
            },
            { name = "permission_mode", type = "select", label = "Permission mode", default = "plan",
              options = {
                  { value = "plan", label = "Plan mode" },
                  { value = "default", label = "Default mode" },
              }
            },
            { name = "skip_permissions", type = "boolean", label = "Skip permission confirmation", default = false },
            { name = "think_level", type = "select", label = "Think level", default = "ultrathink",
              options = {
                  { value = "ultrathink", label = "Ultrathink" },
                  { value = "megathink", label = "Megathink" },
                  { value = "", label = "Default" },
              }
            },
        },
    },
}

-- Skills are loaded from external markdown files via ctx.read_builtin_resource()
-- This improves maintainability by separating content from code logic.

function setup(ctx)
    setup_mcp(ctx)
    setup_skills(ctx)
    setup_permission(ctx)
    setup_model(ctx)
    setup_think_level(ctx)
end

-- MCP configuration: via --mcp-config parameter
-- Uses shared mcp_utils module for common logic
function setup_mcp(ctx)
    mcp_utils.setup_via_config_file(ctx)
end

-- Skills injection: create .claude/skills directory
-- Skills content is loaded from external markdown files for better maintainability
function setup_skills(ctx)
    -- Default to enabled if not specified
    local skills_enabled = ctx.config.skills_enabled
    if skills_enabled == nil then
        skills_enabled = true
    end

    if not skills_enabled then
        ctx.log("Skills disabled, skipping")
        return
    end

    local skills_dir = ctx.sandbox.work_dir .. "/.claude/skills"

    -- Define skills to inject (name -> resource path)
    local skills = {
        ["am-delegate"] = "skills/am-delegate.md",
        ["am-channel"] = "skills/am-channel.md",
    }

    for skill_name, resource_path in pairs(skills) do
        -- Load skill content from embedded resource
        local content, err = ctx.read_builtin_resource(resource_path)
        if err then
            ctx.log("Warning: failed to load skill " .. skill_name .. ": " .. err)
        else
            -- Create skill directory and files
            local skill_dir = skills_dir .. "/" .. skill_name
            ctx.mkdir(skill_dir)
            ctx.write_file(skill_dir .. "/SKILL.md", content)
            ctx.write_file(skill_dir .. "/.gitignore", "*\n")
        end
    end

    ctx.log("Skills injected at " .. skills_dir)
end

-- Permission mode
function setup_permission(ctx)
    ctx.log("skip_permissions=" .. tostring(ctx.config.skip_permissions) .. ", permission_mode=" .. tostring(ctx.config.permission_mode))
    if ctx.config.skip_permissions then
        ctx.log("Adding --dangerously-skip-permissions")
        ctx.add_args("--dangerously-skip-permissions")
    elseif ctx.config.permission_mode and ctx.config.permission_mode ~= "" then
        ctx.log("Adding --permission-mode " .. ctx.config.permission_mode)
        ctx.add_args("--permission-mode", ctx.config.permission_mode)
    end
end

-- Model selection
function setup_model(ctx)
    if ctx.config.model and ctx.config.model ~= "" then
        ctx.add_args("--model", ctx.config.model)
    end
end

-- Think level: append to prompt
function setup_think_level(ctx)
    if ctx.config.think_level and ctx.config.think_level ~= "" then
        ctx.append_prompt("\n\n" .. ctx.config.think_level)
    end
end

function teardown(ctx)
    -- Files are cleaned up with sandbox
end
