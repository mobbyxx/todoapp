-- +goose Up
-- Create todos table
CREATE TABLE todos (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(255) NOT NULL,
    description TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'in_progress', 'completed', 'archived')),
    priority VARCHAR(10) NOT NULL DEFAULT 'medium' CHECK (priority IN ('low', 'medium', 'high', 'urgent')),
    created_by UUID NOT NULL,
    assigned_to UUID,
    due_date DATE,
    completed_at TIMESTAMP WITH TIME ZONE,
    version INTEGER NOT NULL DEFAULT 1,
    tags JSONB DEFAULT '[]',
    metadata JSONB DEFAULT '{}',
    is_deleted BOOLEAN DEFAULT false,
    deleted_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create trigger for todos table
CREATE TRIGGER update_todos_updated_at
    BEFORE UPDATE ON todos
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Create indexes for todos table
CREATE INDEX idx_todos_created_by ON todos(created_by);
CREATE INDEX idx_todos_assigned_to ON todos(assigned_to);
CREATE INDEX idx_todos_status ON todos(status);
CREATE INDEX idx_todos_priority ON todos(priority);
CREATE INDEX idx_todos_due_date ON todos(due_date);
CREATE INDEX idx_todos_active ON todos(created_by, status) WHERE status != 'archived' AND is_deleted = false;
CREATE INDEX idx_todos_tags ON todos USING GIN(tags);
CREATE INDEX idx_todos_metadata ON todos USING GIN(metadata);

-- Create foreign keys
ALTER TABLE todos
    ADD CONSTRAINT fk_todos_created_by FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE CASCADE,
    ADD CONSTRAINT fk_todos_assigned_to FOREIGN KEY (assigned_to) REFERENCES users(id) ON DELETE SET NULL;

-- +goose Down
ALTER TABLE todos DROP CONSTRAINT IF EXISTS fk_todos_created_by;
ALTER TABLE todos DROP CONSTRAINT IF EXISTS fk_todos_assigned_to;
DROP TRIGGER IF EXISTS update_todos_updated_at ON todos;
DROP TABLE IF EXISTS todos;
