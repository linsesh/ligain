-- Add avatar columns to player table
ALTER TABLE player ADD COLUMN avatar_object_key TEXT;
ALTER TABLE player ADD COLUMN avatar_signed_url TEXT;
ALTER TABLE player ADD COLUMN avatar_signed_url_expires_at TIMESTAMP WITH TIME ZONE;
