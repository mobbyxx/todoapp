-- +goose Up
-- Create reward_redemptions table
CREATE TABLE reward_redemptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    reward_id UUID NOT NULL,
    points_spent INTEGER NOT NULL DEFAULT 0,
    redeemed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE,
    is_used BOOLEAN DEFAULT false,
    used_at TIMESTAMP WITH TIME ZONE,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create trigger for reward_redemptions table
CREATE TRIGGER update_reward_redemptions_updated_at
    BEFORE UPDATE ON reward_redemptions
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Create indexes for reward_redemptions table
CREATE INDEX idx_reward_redemptions_user_id ON reward_redemptions(user_id);
CREATE INDEX idx_reward_redemptions_reward_id ON reward_redemptions(reward_id);
CREATE INDEX idx_reward_redemptions_redeemed_at ON reward_redemptions(redeemed_at);
CREATE INDEX idx_reward_redemptions_active ON reward_redemptions(user_id, is_used) WHERE is_used = false;

-- Create foreign keys
ALTER TABLE reward_redemptions
    ADD CONSTRAINT fk_reward_redemptions_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    ADD CONSTRAINT fk_reward_redemptions_reward FOREIGN KEY (reward_id) REFERENCES rewards(id) ON DELETE CASCADE;

-- +goose Down
ALTER TABLE reward_redemptions DROP CONSTRAINT IF EXISTS fk_reward_redemptions_user;
ALTER TABLE reward_redemptions DROP CONSTRAINT IF EXISTS fk_reward_redemptions_reward;
DROP TRIGGER IF EXISTS update_reward_redemptions_updated_at ON reward_redemptions;
DROP TABLE IF EXISTS reward_redemptions;
