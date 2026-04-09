-- +goose Up
-- Fix schema mismatches between domain model and database

-- shared_goals: add columns expected by the repository
ALTER TABLE shared_goals
    ADD COLUMN IF NOT EXISTS connection_id UUID,
    ADD COLUMN IF NOT EXISTS target_type VARCHAR(50) DEFAULT 'todos_completed',
    ADD COLUMN IF NOT EXISTS target_value INTEGER DEFAULT 0,
    ADD COLUMN IF NOT EXISTS current_value INTEGER DEFAULT 0,
    ADD COLUMN IF NOT EXISTS reward_description TEXT DEFAULT '';

-- connections: allow user_a_id/user_b_id to be the nil UUID for pending invitations
-- (invitation flow uses NormalizeUserPair(userID, uuid.Nil) which may set either to nil)
ALTER TABLE connections DROP CONSTRAINT IF EXISTS chk_different_users;
ALTER TABLE connections DROP CONSTRAINT IF EXISTS fk_connections_user_b;
ALTER TABLE connections DROP CONSTRAINT IF EXISTS fk_connections_user_a;

-- +goose Down
ALTER TABLE connections
    ADD CONSTRAINT fk_connections_user_b FOREIGN KEY (user_b_id) REFERENCES users(id) ON DELETE CASCADE,
    ADD CONSTRAINT chk_different_users CHECK (user_a_id <> user_b_id);

ALTER TABLE shared_goals
    DROP COLUMN IF EXISTS reward_description,
    DROP COLUMN IF EXISTS current_value,
    DROP COLUMN IF EXISTS target_value,
    DROP COLUMN IF EXISTS target_type,
    DROP COLUMN IF EXISTS connection_id;
