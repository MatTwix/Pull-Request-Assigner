CREATE UNIQUE INDEX IF NOT EXISTS idx_users_user_id
    ON users (user_id);

CREATE INDEX IF NOT EXISTS idx_users_team_name_id
    ON users (team_name, id);

CREATE UNIQUE INDEX IF NOT EXISTS idx_teams_team_name
    ON teams (team_name);