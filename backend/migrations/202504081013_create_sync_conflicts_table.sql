-- +goose Up
-- Create sync_conflicts table
CREATE TABLE sync_conflicts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    entity_type VARCHAR(50) NOT NULL,
    entity_id UUID NOT NULL,
    local_version INTEGER NOT NULL,
    remote_version INTEGER NOT NULL,
    local_data JSONB NOT NULL,
    remote_data JSONB NOT NULL,
    resolution_strategy VARCHAR(50) NOT NULL CHECK (resolution_strategy IN ('client_wins', 'server_wins', 'merge', 'manual')),
    resolved_data JSONB,
    resolved_at TIMESTAMP WITH TIME ZONE,
    resolved_by UUID,
    status VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending', 'resolved', 'ignored')),
    client_timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    server_timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create trigger for sync_conflicts table
CREATE TRIGGER update_sync_conflicts_updated_at
    BEFORE UPDATE ON sync_conflicts
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Create indexes for sync_conflicts table
CREATE INDEX idx_sync_conflicts_user_id ON sync_conflicts(user_id);
CREATE INDEX idx_sync_conflicts_entity ON sync_conflicts(entity_type, entity_id);
CREATE INDEX idx_sync_conflicts_status ON sync_conflicts(status);
CREATE INDEX idx_sync_conflicts_pending ON sync_conflicts(user_id, status) WHERE status = 'pending';

-- Create foreign keys
ALTER TABLE sync_conflicts
    ADD CONSTRAINT fk_sync_conflicts_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    ADD CONSTRAINT fk_sync_conflicts_resolved_by FOREIGN KEY (resolved_by) REFERENCES users(id) ON DELETE SET NULL;

-- +goose Down
ALTER TABLE sync_conflicts DROP CONSTRAINT IF EXISTS fk_sync_conflicts_user;
ALTER TABLE sync_conflicts DROP CONSTRAINT IF EXISTS fk_sync_conflicts_resolved_by;
DROP TRIGGER IF EXISTS update_sync_conflicts_updated_at ON sync_conflicts;
DROP TABLE IF EXISTS sync_conflicts;
