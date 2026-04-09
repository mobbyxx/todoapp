-- +goose Up
-- Create rewards table
CREATE TABLE rewards (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    description TEXT NOT NULL,
    type VARCHAR(20) NOT NULL CHECK (type IN ('badge', 'points', 'feature')),
    value INTEGER NOT NULL,
    icon_url TEXT,
    required_level INTEGER,
    cost_points INTEGER DEFAULT 0,
    is_active BOOLEAN DEFAULT true,
    valid_from TIMESTAMP WITH TIME ZONE,
    valid_until TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create trigger for rewards table
CREATE TRIGGER update_rewards_updated_at
    BEFORE UPDATE ON rewards
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Create indexes for rewards table
CREATE INDEX idx_rewards_type ON rewards(type);
CREATE INDEX idx_rewards_active ON rewards(is_active) WHERE is_active = true;
CREATE INDEX idx_rewards_required_level ON rewards(required_level);

-- +goose Down
DROP TRIGGER IF EXISTS update_rewards_updated_at ON rewards;
DROP TABLE IF EXISTS rewards;
