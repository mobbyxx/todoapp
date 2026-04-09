-- +goose Up
-- Add invitation token fields to connections table
ALTER TABLE connections
    ADD COLUMN IF NOT EXISTS invitation_token VARCHAR(36) UNIQUE,
    ADD COLUMN IF NOT EXISTS expires_at TIMESTAMP WITH TIME ZONE,
    ADD COLUMN IF NOT EXISTS rejected_at TIMESTAMP WITH TIME ZONE;

-- Create index for invitation token lookups
CREATE INDEX IF NOT EXISTS idx_connections_invitation_token ON connections(invitation_token) WHERE invitation_token IS NOT NULL;

-- +goose Down
DROP INDEX IF EXISTS idx_connections_invitation_token;
ALTER TABLE connections
    DROP COLUMN IF EXISTS invitation_token,
    DROP COLUMN IF EXISTS expires_at,
    DROP COLUMN IF EXISTS rejected_at;
