-- Remove avatar columns from player table
ALTER TABLE player DROP COLUMN IF EXISTS avatar_object_key;
ALTER TABLE player DROP COLUMN IF EXISTS avatar_signed_url;
ALTER TABLE player DROP COLUMN IF EXISTS avatar_signed_url_expires_at;
