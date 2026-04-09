-- +goose Up
-- Create points_history table (gamification ledger)
CREATE TABLE points_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    amount INTEGER NOT NULL,
    balance_after INTEGER NOT NULL,
    reason VARCHAR(100) NOT NULL,
    reference_type VARCHAR(50),
    reference_id UUID,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes for points_history table
CREATE INDEX idx_points_history_user_id ON points_history(user_id);
CREATE INDEX idx_points_history_created_at ON points_history(created_at);
CREATE INDEX idx_points_history_user_created ON points_history(user_id, created_at);
CREATE INDEX idx_points_history_reference ON points_history(reference_type, reference_id);

-- Create foreign key
ALTER TABLE points_history
    ADD CONSTRAINT fk_points_history_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

-- Add foreign key to users table for current_level_id (must be after levels table exists)
ALTER TABLE users
    ADD CONSTRAINT fk_users_current_level FOREIGN KEY (current_level_id) REFERENCES levels(id) ON DELETE SET NULL;

-- +goose Down
ALTER TABLE users DROP CONSTRAINT IF EXISTS fk_users_current_level;
ALTER TABLE points_history DROP CONSTRAINT IF EXISTS fk_points_history_user;
DROP TABLE IF EXISTS points_history;
