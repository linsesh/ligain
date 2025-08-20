-- Add CASCADE DELETE constraints for player deletion

-- First, we need to add the missing foreign key constraints with CASCADE DELETE
-- for bet and score tables

-- For bet table: add CASCADE DELETE on player_id
ALTER TABLE bet DROP CONSTRAINT IF EXISTS bet_player_id_fkey;
ALTER TABLE bet ADD CONSTRAINT bet_player_id_fkey 
    FOREIGN KEY (player_id) REFERENCES player(id) ON DELETE CASCADE;

-- For score table: add CASCADE DELETE on player_id  
ALTER TABLE score DROP CONSTRAINT IF EXISTS score_player_id_fkey;
ALTER TABLE score ADD CONSTRAINT score_player_id_fkey 
    FOREIGN KEY (player_id) REFERENCES player(id) ON DELETE CASCADE;

-- auth_tokens already has CASCADE DELETE, but let's ensure it's there
ALTER TABLE auth_tokens DROP CONSTRAINT IF EXISTS auth_tokens_player_id_fkey;
ALTER TABLE auth_tokens ADD CONSTRAINT auth_tokens_player_id_fkey 
    FOREIGN KEY (player_id) REFERENCES player(id) ON DELETE CASCADE;

-- game_player already has CASCADE DELETE on both game and player
