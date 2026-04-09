-- +goose Up
-- Create user_badges table (junction table for users and badges)
CREATE TABLE user_badges (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    badge_id UUID NOT NULL,
    awarded_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    awarded_by UUID,
    context JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT uq_user_badge UNIQUE (user_id, badge_id)
);

-- Create trigger for user_badges table
CREATE TRIGGER update_user_badges_updated_at
    BEFORE UPDATE ON user_badges
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Create indexes for user_badges table
CREATE INDEX idx_user_badges_user_id ON user_badges(user_id);
CREATE INDEX idx_user_badges_badge_id ON user_badges(badge_id);
CREATE INDEX idx_user_badges_awarded_at ON user_badges(awarded_at);

-- Create foreign keys
ALTER TABLE user_badges
    ADD CONSTRAINT fk_user_badges_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    ADD CONSTRAINT fk_user_badges_badge FOREIGN KEY (badge_id) REFERENCES badges(id) ON DELETE CASCADE,
    ADD CONSTRAINT fk_user_badges_awarded_by FOREIGN KEY (awarded_by) REFERENCES users(id) ON DELETE SET NULL;

-- +goose Down
ALTER TABLE user_badges DROP CONSTRAINT IF EXISTS fk_user_badges_user;
ALTER TABLE user_badges DROP CONSTRAINT IF EXISTS fk_user_badges_badge;
ALTER TABLE user_badges DROP CONSTRAINT IF EXISTS fk_user_badges_awarded_by;
DROP TRIGGER IF EXISTS update_user_badges_updated_at ON user_badges;
DROP TABLE IF EXISTS user_badges;
