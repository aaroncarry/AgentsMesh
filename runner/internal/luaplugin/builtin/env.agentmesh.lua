-- builtin/env.lua
-- Universal environment variable injection plugin (supports all agents)

plugin = {
    name = "env",
    version = "1.0.0",
    description = "Universal environment variable injection",
    supported_agents = {},  -- Empty array means supports all agents
    order = 10,  -- Execute first
    critical = true,
    ui = {
        configurable = false,  -- Not shown in UI
    },
}

function setup(ctx)
    -- Inject environment variables from CreatePodCommand
    if ctx.config.env_vars then
        local count = 0
        for key, value in pairs(ctx.config.env_vars) do
            ctx.add_env(key, value)
            count = count + 1
        end
        if count > 0 then
            ctx.log("Injected " .. count .. " environment variables")
        end
    end
end

function teardown(ctx)
    -- Environment variables are cleaned up with process termination
end
