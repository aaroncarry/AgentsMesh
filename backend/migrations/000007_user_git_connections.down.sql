-- Drop user_git_connections table
DROP TRIGGER IF EXISTS update_user_git_connections_updated_at ON user_git_connections;
DROP TABLE IF EXISTS user_git_connections;
