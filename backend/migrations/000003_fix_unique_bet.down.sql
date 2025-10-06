-- 1️⃣ Drop the new unique constraint
ALTER TABLE bet DROP CONSTRAINT IF EXISTS bet_game_id_match_id_player_id_key;

-- 2️⃣ Restore the old unique constraint
ALTER TABLE bet ADD CONSTRAINT bet_match_id_player_id_key UNIQUE (match_id, player_id);