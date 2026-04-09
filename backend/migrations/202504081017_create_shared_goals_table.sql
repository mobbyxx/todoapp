-- +goose Up
-- Create shared_goals table (collaborative goals)
CREATE TABLE shared_goals (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(255) NOT NULL,
    description TEXT,
    created_by UUID NOT NULL,
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'completed', 'cancelled')),
    target_date DATE,
    completed_at TIMESTAMP WITH TIME ZONE,
    total_contributions INTEGER DEFAULT 0,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create shared_goal_members junction table
CREATE TABLE shared_goal_members (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    goal_id UUID NOT NULL,
    user_id UUID NOT NULL,
    role VARCHAR(20) DEFAULT 'member' CHECK (role IN ('owner', 'admin', 'member')),
    contribution_points INTEGER DEFAULT 0,
    joined_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    left_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT uq_goal_member UNIQUE (goal_id, user_id)
);

-- Create trigger for shared_goals table
CREATE TRIGGER update_shared_goals_updated_at
    BEFORE UPDATE ON shared_goals
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Create trigger for shared_goal_members table
CREATE TRIGGER update_shared_goal_members_updated_at
    BEFORE UPDATE ON shared_goal_members
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Create indexes for shared_goals table
CREATE INDEX idx_shared_goals_created_by ON shared_goals(created_by);
CREATE INDEX idx_shared_goals_status ON shared_goals(status);

-- Create indexes for shared_goal_members table
CREATE INDEX idx_shared_goal_members_goal_id ON shared_goal_members(goal_id);
CREATE INDEX idx_shared_goal_members_user_id ON shared_goal_members(user_id);
CREATE INDEX idx_shared_goal_members_active ON shared_goal_members(goal_id, user_id) WHERE left_at IS NULL;

-- Create foreign keys for shared_goals
ALTER TABLE shared_goals
    ADD CONSTRAINT fk_shared_goals_created_by FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE CASCADE;

-- Create foreign keys for shared_goal_members
ALTER TABLE shared_goal_members
    ADD CONSTRAINT fk_sgm_goal FOREIGN KEY (goal_id) REFERENCES shared_goals(id) ON DELETE CASCADE,
    ADD CONSTRAINT fk_sgm_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

-- +goose Down
ALTER TABLE shared_goal_members DROP CONSTRAINT IF EXISTS fk_sgm_goal;
ALTER TABLE shared_goal_members DROP CONSTRAINT IF EXISTS fk_sgm_user;
ALTER TABLE shared_goals DROP CONSTRAINT IF EXISTS fk_shared_goals_created_by;
DROP TRIGGER IF EXISTS update_shared_goals_updated_at ON shared_goals;
DROP TRIGGER IF EXISTS update_shared_goal_members_updated_at ON shared_goal_members;
DROP TABLE IF EXISTS shared_goal_members;
DROP TABLE IF EXISTS shared_goals;
