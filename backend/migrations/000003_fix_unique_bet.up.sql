-- 1️⃣ Drop the existing unique constraint
ALTER TABLE bet DROP CONSTRAINT IF EXISTS bet_match_id_player_id_key;

-- 2️⃣ Add the new unique constraint including game_id
ALTER TABLE bet ADD CONSTRAINT bet_game_id_match_id_player_id_key UNIQUE (game_id, match_id, player_id);