-- +goose Up
-- Create levels table
CREATE TABLE levels (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    level_number INTEGER NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    min_points INTEGER NOT NULL,
    max_points INTEGER NOT NULL,
    icon_url TEXT,
    rewards JSONB DEFAULT '[]',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create trigger for levels table
CREATE TRIGGER update_levels_updated_at
    BEFORE UPDATE ON levels
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Create indexes for levels table
CREATE INDEX idx_levels_level_number ON levels(level_number);
CREATE INDEX idx_levels_min_points ON levels(min_points);

-- +goose Down
DROP TRIGGER IF EXISTS update_levels_updated_at ON levels;
DROP TABLE IF EXISTS levels;
