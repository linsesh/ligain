CREATE TABLE IF NOT EXISTS game (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    season_year VARCHAR(255) NOT NULL,
    competition_name VARCHAR(255) NOT NULL,
    status VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS match (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    home_team_id TEXT NOT NULL,
    away_team_id TEXT NOT NULL,
    home_team_score INT,
    away_team_score INT,
    match_date TIMESTAMP WITH TIME ZONE NOT NULL,
    match_status VARCHAR(255) NOT NULL,
    season_code VARCHAR(255) NOT NULL,
    competition_code VARCHAR(255) NOT NULL,
    matchday INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (home_team_id, away_team_id, season_code, competition_code, matchday)
);

CREATE TABLE IF NOT EXISTS player (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL UNIQUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS bet (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    game_id UUID NOT NULL REFERENCES game(id),
    match_id UUID NOT NULL REFERENCES match(id),
    player_id UUID NOT NULL REFERENCES player(id),
    predicted_home_goals INT NOT NULL,
    predicted_away_goals INT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(match_id, player_id)
);

CREATE TABLE IF NOT EXISTS score (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    bet_id UUID NOT NULL REFERENCES bet(id) UNIQUE,
    points INTEGER NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);