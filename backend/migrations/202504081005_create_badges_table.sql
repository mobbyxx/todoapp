-- +goose Up
-- Create badges table
CREATE TABLE badges (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT NOT NULL,
    icon_url TEXT NOT NULL,
    type VARCHAR(20) NOT NULL CHECK (type IN ('achievement', 'milestone', 'special')),
    points_value INTEGER NOT NULL DEFAULT 0,
    criteria JSONB DEFAULT '{}',
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create trigger for badges table
CREATE TRIGGER update_badges_updated_at
    BEFORE UPDATE ON badges
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Create indexes for badges table
CREATE INDEX idx_badges_type ON badges(type);
CREATE INDEX idx_badges_active ON badges(is_active) WHERE is_active = true;
CREATE INDEX idx_badges_criteria ON badges USING GIN(criteria);

-- +goose Down
DROP TRIGGER IF EXISTS update_badges_updated_at ON badges;
DROP TABLE IF EXISTS badges;
