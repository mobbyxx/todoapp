-- +goose Up
-- Add user_id and status columns for custom rewards

-- Add user_id to rewards table for user-defined rewards
ALTER TABLE rewards 
    ADD COLUMN IF NOT EXISTS user_id UUID REFERENCES users(id) ON DELETE CASCADE;

-- Create index for user rewards lookup
CREATE INDEX IF NOT EXISTS idx_rewards_user_id ON rewards(user_id) WHERE user_id IS NOT NULL;

-- Add status column to reward_redemptions for tracking redemption state
ALTER TABLE reward_redemptions 
    ADD COLUMN IF NOT EXISTS status VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected', 'completed'));

-- Create index for redemption status lookups
CREATE INDEX IF NOT EXISTS idx_reward_redemptions_status ON reward_redemptions(status);

-- +goose Down
DROP INDEX IF EXISTS idx_reward_redemptions_status;
DROP INDEX IF EXISTS idx_rewards_user_id;
ALTER TABLE reward_redemptions DROP COLUMN IF EXISTS status;
ALTER TABLE rewards DROP COLUMN IF EXISTS user_id;
