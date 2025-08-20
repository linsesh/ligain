-- Rollback CASCADE DELETE constraints for player deletion

-- Remove CASCADE DELETE from bet table
ALTER TABLE bet DROP CONSTRAINT IF EXISTS bet_player_id_fkey;
ALTER TABLE bet ADD CONSTRAINT bet_player_id_fkey 
    FOREIGN KEY (player_id) REFERENCES player(id);

-- Remove CASCADE DELETE from score table  
ALTER TABLE score DROP CONSTRAINT IF EXISTS score_player_id_fkey;
ALTER TABLE score ADD CONSTRAINT score_player_id_fkey 
    FOREIGN KEY (player_id) REFERENCES player(id);

-- Remove CASCADE DELETE from auth_tokens table
ALTER TABLE auth_tokens DROP CONSTRAINT IF EXISTS auth_tokens_player_id_fkey;
ALTER TABLE auth_tokens ADD CONSTRAINT auth_tokens_player_id_fkey 
    FOREIGN KEY (player_id) REFERENCES player(id);
