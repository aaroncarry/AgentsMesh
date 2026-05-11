INSERT INTO agents (slug, name, launch_command, executable, is_builtin, is_active, supported_modes, agentfile_source)
VALUES ('factory-cli', 'Factory CLI', 'droid', 'droid', true, true, 'pty,acp',
  E'# === Identity ===\nAGENT droid\nEXECUTABLE droid\n\n# === Mode ===\nMODE pty\nMODE acp "exec" "--output-format" "acp"\n\n# === Configuration ===\nCONFIG autonomy_level SELECT("", "off", "low", "medium", "high") = ""\nCONFIG interaction_mode SELECT("", "auto", "spec") = ""\nCONFIG skip_permissions_unsafe BOOL = false\n\n# === Environment ===\nENV FACTORY_API_KEY SECRET OPTIONAL\nENV FACTORY_BASE_URL TEXT OPTIONAL\n\n# === Prompt ===\nPROMPT_POSITION append\n\n# === Capabilities ===\nMCP ON\nSKILLS am-delegate, am-channel\n\n# === Build Logic ===\narg "--resume" when config.resume_enabled and mode != "acp"\narg "--skip-permissions-unsafe" when config.skip_permissions_unsafe and mode == "acp"\n\nfactory_settings_path = sandbox.root + "/factory-runtime-settings.json"\n\nif config.autonomy_level != "" and config.interaction_mode != "" {\n  file factory_settings_path json({ sessionDefaultSettings: { autonomyLevel: config.autonomy_level, interactionMode: config.interaction_mode } })\n  arg "--settings" factory_settings_path\n}\nif config.autonomy_level != "" and config.interaction_mode == "" {\n  file factory_settings_path json({ sessionDefaultSettings: { autonomyLevel: config.autonomy_level } })\n  arg "--settings" factory_settings_path\n}\nif config.autonomy_level == "" and config.interaction_mode != "" {\n  file factory_settings_path json({ sessionDefaultSettings: { interactionMode: config.interaction_mode } })\n  arg "--settings" factory_settings_path\n}\n\nif mcp.enabled {\n  mkdir sandbox.work_dir + "/.factory"\n  file sandbox.work_dir + "/.factory/mcp.json" json({ mcpServers: mcp.servers })\n}\n')
ON CONFLICT (slug) DO UPDATE SET
  name = EXCLUDED.name,
  launch_command = EXCLUDED.launch_command,
  executable = EXCLUDED.executable,
  is_builtin = EXCLUDED.is_builtin,
  is_active = EXCLUDED.is_active,
  supported_modes = EXCLUDED.supported_modes,
  agentfile_source = EXCLUDED.agentfile_source,
  updated_at = now();

UPDATE organization_agents
SET agent_slug = 'factory-cli'
WHERE agent_slug = 'factory-droid';

UPDATE organization_agent_configs
SET agent_slug = 'factory-cli'
WHERE agent_slug = 'factory-droid';

UPDATE pods
SET agent_slug = 'factory-cli'
WHERE agent_slug = 'factory-droid';

UPDATE loops
SET agent_slug = 'factory-cli'
WHERE agent_slug = 'factory-droid';

UPDATE user_agent_configs
SET agent_slug = 'factory-cli'
WHERE agent_slug = 'factory-droid';

UPDATE user_agent_credential_profiles
SET agent_slug = 'factory-cli'
WHERE agent_slug = 'factory-droid';

UPDATE token_usages
SET agent_slug = 'factory-cli'
WHERE agent_slug = 'factory-droid';

UPDATE autopilot_controllers
SET control_agent_slug = 'factory-cli'
WHERE control_agent_slug = 'factory-droid';

UPDATE runners
SET available_agents = (
  SELECT COALESCE(
    jsonb_agg(
      CASE
        WHEN elem = to_jsonb('factory-droid'::text) THEN to_jsonb('factory-cli'::text)
        ELSE elem
      END
    ),
    '[]'::jsonb
  )
  FROM jsonb_array_elements(available_agents) AS elem
)
WHERE available_agents ? 'factory-droid';

UPDATE runners
SET agent_versions = (
  SELECT COALESCE(
    jsonb_agg(
      CASE
        WHEN elem->>'slug' = 'factory-droid' THEN jsonb_set(elem, '{slug}', to_jsonb('factory-cli'::text), false)
        ELSE elem
      END
    ),
    '[]'::jsonb
  )
  FROM jsonb_array_elements(agent_versions) AS elem
)
WHERE agent_versions @> '[{"slug":"factory-droid"}]'::jsonb;

UPDATE skill_market_items
SET agent_filter = (
  SELECT jsonb_agg(to_jsonb(slug) ORDER BY first_seen)
  FROM (
    SELECT
      CASE
        WHEN elem #>> '{}' = 'factory-droid' THEN 'factory-cli'
        ELSE elem #>> '{}'
      END AS slug,
      MIN(ord) AS first_seen
    FROM jsonb_array_elements(agent_filter) WITH ORDINALITY AS t(elem, ord)
    GROUP BY 1
  ) normalized
)
WHERE agent_filter @> '["factory-droid"]'::jsonb;

UPDATE mcp_market_items
SET agent_filter = (
  SELECT jsonb_agg(to_jsonb(slug) ORDER BY first_seen)
  FROM (
    SELECT
      CASE
        WHEN elem #>> '{}' = 'factory-droid' THEN 'factory-cli'
        ELSE elem #>> '{}'
      END AS slug,
      MIN(ord) AS first_seen
    FROM jsonb_array_elements(agent_filter) WITH ORDINALITY AS t(elem, ord)
    GROUP BY 1
  ) normalized
)
WHERE agent_filter @> '["factory-droid"]'::jsonb;

DELETE FROM agents
WHERE slug = 'factory-droid'
  AND is_builtin = true;
