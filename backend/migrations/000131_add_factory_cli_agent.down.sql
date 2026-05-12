INSERT INTO agents (
  slug,
  name,
  launch_command,
  executable,
  is_builtin,
  is_active,
  supported_modes,
  agentfile_source
)
SELECT
  'factory-droid',
  'Factory (Droid)',
  launch_command,
  executable,
  is_builtin,
  is_active,
  supported_modes,
  agentfile_source
FROM agents
WHERE slug = 'factory-cli'
  AND NOT EXISTS (SELECT 1 FROM agents WHERE slug = 'factory-droid');

UPDATE organization_agents
SET agent_slug = 'factory-droid'
WHERE agent_slug = 'factory-cli';

UPDATE organization_agent_configs
SET agent_slug = 'factory-droid'
WHERE agent_slug = 'factory-cli';

UPDATE pods
SET agent_slug = 'factory-droid'
WHERE agent_slug = 'factory-cli';

UPDATE loops
SET agent_slug = 'factory-droid'
WHERE agent_slug = 'factory-cli';

UPDATE user_agent_configs
SET agent_slug = 'factory-droid'
WHERE agent_slug = 'factory-cli';

UPDATE user_agent_credential_profiles
SET agent_slug = 'factory-droid'
WHERE agent_slug = 'factory-cli';

UPDATE token_usages
SET agent_slug = 'factory-droid'
WHERE agent_slug = 'factory-cli';

UPDATE autopilot_controllers
SET control_agent_slug = 'factory-droid'
WHERE control_agent_slug = 'factory-cli';

UPDATE runners
SET available_agents = (
  SELECT COALESCE(
    jsonb_agg(
      CASE
        WHEN elem = to_jsonb('factory-cli'::text) THEN to_jsonb('factory-droid'::text)
        ELSE elem
      END
    ),
    '[]'::jsonb
  )
  FROM jsonb_array_elements(available_agents) AS elem
)
WHERE available_agents ? 'factory-cli';

UPDATE runners
SET agent_versions = (
  SELECT COALESCE(
    jsonb_agg(
      CASE
        WHEN elem->>'slug' = 'factory-cli' THEN jsonb_set(elem, '{slug}', to_jsonb('factory-droid'::text), false)
        ELSE elem
      END
    ),
    '[]'::jsonb
  )
  FROM jsonb_array_elements(agent_versions) AS elem
)
WHERE agent_versions @> '[{"slug":"factory-cli"}]'::jsonb;

UPDATE skill_market_items
SET agent_filter = (
  SELECT jsonb_agg(to_jsonb(slug) ORDER BY first_seen)
  FROM (
    SELECT
      CASE
        WHEN elem #>> '{}' = 'factory-cli' THEN 'factory-droid'
        ELSE elem #>> '{}'
      END AS slug,
      MIN(ord) AS first_seen
    FROM jsonb_array_elements(agent_filter) WITH ORDINALITY AS t(elem, ord)
    GROUP BY 1
  ) normalized
)
WHERE agent_filter @> '["factory-cli"]'::jsonb;

UPDATE mcp_market_items
SET agent_filter = (
  SELECT jsonb_agg(to_jsonb(slug) ORDER BY first_seen)
  FROM (
    SELECT
      CASE
        WHEN elem #>> '{}' = 'factory-cli' THEN 'factory-droid'
        ELSE elem #>> '{}'
      END AS slug,
      MIN(ord) AS first_seen
    FROM jsonb_array_elements(agent_filter) WITH ORDINALITY AS t(elem, ord)
    GROUP BY 1
  ) normalized
)
WHERE agent_filter @> '["factory-cli"]'::jsonb;

DELETE FROM agents
WHERE slug = 'factory-cli'
  AND is_builtin = true;
