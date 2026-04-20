UPDATE agents
SET agentfile_source = REPLACE(
    agentfile_source,
    E'arg "--model" config.model when config.model != ""\narg "--resume" config.resume_session when config.resume_enabled\n\n',
    E'arg "--model" config.model when config.model != ""\n\n'
)
WHERE slug = 'claude-code';
