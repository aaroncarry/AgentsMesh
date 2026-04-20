-- Restore Claude resume support in AgentFile after command_template removal.
-- Resume reuses the previous session_id via system overrides:
--   resume_enabled=true, resume_session=<session_id>
UPDATE agents
SET agentfile_source = REPLACE(
    agentfile_source,
    E'# === Build Logic ===\narg "--model" config.model when config.model != ""\n\n',
    E'# === Build Logic ===\narg "--model" config.model when config.model != ""\narg "--resume" config.resume_session when config.resume_enabled\n\n'
)
WHERE slug = 'claude-code'
  AND agentfile_source NOT LIKE '%arg "--resume" config.resume_session when config.resume_enabled%';
