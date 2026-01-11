-- builtin/mcp_utils.lua
-- Shared MCP configuration utilities for all agent plugins
-- This file is loaded automatically before other plugins and provides common functions.

-- MCP utility module
mcp_utils = {}

-- Default MCP port
mcp_utils.DEFAULT_PORT = 19000

-- Default MCP server name
mcp_utils.SERVER_NAME = "agentmesh"

--- Validate that required configuration values are present and valid.
--- @param ctx table The context object with config and sandbox
--- @param required table List of required config keys (optional)
--- @return boolean, string Returns true if valid, or false with error message
function mcp_utils.validate_config(ctx, required)
    -- Validate ctx structure
    if not ctx then
        return false, "Context is nil"
    end
    if not ctx.sandbox then
        return false, "Context.sandbox is nil"
    end
    if not ctx.sandbox.pod_key or ctx.sandbox.pod_key == "" then
        return false, "Context.sandbox.pod_key is missing"
    end
    if not ctx.sandbox.root_path or ctx.sandbox.root_path == "" then
        return false, "Context.sandbox.root_path is missing"
    end
    if not ctx.sandbox.work_dir or ctx.sandbox.work_dir == "" then
        return false, "Context.sandbox.work_dir is missing"
    end

    -- Validate required config fields if specified
    if required then
        ctx.config = ctx.config or {}
        for _, key in ipairs(required) do
            if ctx.config[key] == nil then
                return false, "Required config key missing: " .. key
            end
        end
    end

    -- Validate MCP port if specified
    if ctx.config and ctx.config.mcp_port then
        local port = ctx.config.mcp_port
        if type(port) ~= "number" or port < 1 or port > 65535 then
            return false, "Invalid mcp_port: must be a number between 1 and 65535"
        end
    end

    return true
end

-- Check if MCP is enabled in config
-- Returns true if enabled (defaults to true if not specified)
function mcp_utils.is_enabled(ctx)
    local enabled = ctx.config.mcp_enabled
    if enabled == nil then
        return true
    end
    return enabled
end

-- Get MCP port from config
-- Returns the configured port or default
function mcp_utils.get_port(ctx)
    return ctx.config.mcp_port or mcp_utils.DEFAULT_PORT
end

-- Get MCP URL
function mcp_utils.get_url(ctx)
    local port = mcp_utils.get_port(ctx)
    return string.format("http://127.0.0.1:%d/mcp", port)
end

-- Get standard MCP headers with pod key
function mcp_utils.get_headers(ctx)
    return { ["X-Pod-Key"] = ctx.sandbox.pod_key }
end

-- Set common metadata after MCP setup
function mcp_utils.set_metadata(ctx, config_path)
    local port = mcp_utils.get_port(ctx)
    if config_path then
        ctx.set_metadata("mcp_config_path", config_path)
    end
    ctx.set_metadata("mcp_port", port)
end

-- Log MCP disabled message and return false
-- Returns false to indicate setup should be skipped
function mcp_utils.skip_if_disabled(ctx)
    if not mcp_utils.is_enabled(ctx) then
        ctx.log("MCP disabled, skipping")
        return true
    end
    return false
end

-- Build standard MCP server config for file-based configuration
-- format: "claude" | "gemini" | "opencode"
function mcp_utils.build_server_config(ctx, format)
    local url = mcp_utils.get_url(ctx)
    local headers = mcp_utils.get_headers(ctx)

    if format == "claude" then
        -- Claude Code format
        return {
            type = "http",
            url = url,
            headers = headers,
        }
    elseif format == "gemini" then
        -- Gemini CLI format
        return {
            httpUrl = url,
            headers = headers,
        }
    elseif format == "opencode" then
        -- OpenCode format
        return {
            type = "http",
            url = url,
            headers = headers,
            enabled = true,
        }
    else
        -- Default format
        return {
            url = url,
            headers = headers,
        }
    end
end

-- Setup MCP via file configuration (for Claude Code style)
-- Writes mcp-config.json and adds --mcp-config argument
-- Returns true on success, false and error message on failure
function mcp_utils.setup_via_config_file(ctx)
    if mcp_utils.skip_if_disabled(ctx) then
        return true
    end

    local server_config = mcp_utils.build_server_config(ctx, "claude")
    local mcp_config = {
        mcpServers = {
            [mcp_utils.SERVER_NAME] = server_config,
        },
    }

    local config_path = ctx.sandbox.root_path .. "/mcp-config.json"
    local json_content = ctx.json_encode(mcp_config)
    if json_content == nil then
        ctx.log("ERROR: Failed to encode MCP config to JSON")
        return false, "Failed to encode MCP config"
    end

    local ok, err = ctx.write_file(config_path, json_content)
    if not ok then
        ctx.log("ERROR: Failed to write MCP config: " .. (err or "unknown error"))
        return false, err or "Failed to write MCP config"
    end

    ctx.add_args("--mcp-config", config_path)
    mcp_utils.set_metadata(ctx, config_path)
    ctx.log("MCP configured at " .. config_path)
    return true
end

-- Setup MCP via project settings (for Gemini CLI style)
-- Writes to .gemini/settings.json in work_dir
-- Returns true on success, false and error message on failure
function mcp_utils.setup_via_gemini_settings(ctx)
    if mcp_utils.skip_if_disabled(ctx) then
        return true
    end

    local gemini_dir = ctx.sandbox.work_dir .. "/.gemini"
    local config_path = gemini_dir .. "/settings.json"

    local ok, err = ctx.mkdir(gemini_dir)
    if not ok then
        ctx.log("ERROR: Failed to create .gemini directory: " .. (err or "unknown error"))
        return false, err or "Failed to create .gemini directory"
    end

    -- Read existing config or create new
    local settings = ctx.read_json(config_path) or {}
    settings.mcpServers = settings.mcpServers or {}

    -- Add MCP server
    settings.mcpServers[mcp_utils.SERVER_NAME] = mcp_utils.build_server_config(ctx, "gemini")

    local json_content = ctx.json_encode(settings)
    if json_content == nil then
        ctx.log("ERROR: Failed to encode Gemini settings to JSON")
        return false, "Failed to encode settings"
    end

    ok, err = ctx.write_file(config_path, json_content)
    if not ok then
        ctx.log("ERROR: Failed to write Gemini settings: " .. (err or "unknown error"))
        return false, err or "Failed to write settings"
    end

    mcp_utils.set_metadata(ctx, config_path)
    ctx.log("MCP configured at " .. config_path)
    return true
end

-- Setup MCP via command line arguments (for Codex CLI style)
-- Uses -c parameter to override config
-- Returns true on success (this method cannot fail as it only adds args)
function mcp_utils.setup_via_cli_args(ctx)
    if mcp_utils.skip_if_disabled(ctx) then
        return true
    end

    local url = mcp_utils.get_url(ctx)
    local pod_key = ctx.sandbox.pod_key

    ctx.add_args("-c", string.format('mcp_servers.%s.url="%s"', mcp_utils.SERVER_NAME, url))
    ctx.add_args("-c", string.format('mcp_servers.%s.http_headers={"X-Pod-Key":"%s"}', mcp_utils.SERVER_NAME, pod_key))

    mcp_utils.set_metadata(ctx, nil)
    ctx.log("MCP configured via -c arguments")
    return true
end

-- Setup MCP via opencode.json (for OpenCode style)
-- Writes to opencode.json in work_dir
-- Returns true on success, false and error message on failure
function mcp_utils.setup_via_opencode_config(ctx)
    if mcp_utils.skip_if_disabled(ctx) then
        return true
    end

    local config_path = ctx.sandbox.work_dir .. "/opencode.json"

    -- Read existing config or create new
    local config = ctx.read_json(config_path) or {}
    config.mcp = config.mcp or {}

    -- Add MCP server
    config.mcp[mcp_utils.SERVER_NAME] = mcp_utils.build_server_config(ctx, "opencode")

    local json_content = ctx.json_encode(config)
    if json_content == nil then
        ctx.log("ERROR: Failed to encode OpenCode config to JSON")
        return false, "Failed to encode config"
    end

    local ok, err = ctx.write_file(config_path, json_content)
    if not ok then
        ctx.log("ERROR: Failed to write OpenCode config: " .. (err or "unknown error"))
        return false, err or "Failed to write config"
    end

    mcp_utils.set_metadata(ctx, config_path)
    ctx.log("MCP configured at " .. config_path)
    return true
end
