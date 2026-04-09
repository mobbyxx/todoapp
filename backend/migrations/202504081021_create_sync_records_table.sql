-- +goose Up
-- Create sync_records table for tracking user sync state
CREATE TABLE sync_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL UNIQUE,
    last_synced_at_physical BIGINT NOT NULL DEFAULT 0,
    last_synced_at_logical BIGINT NOT NULL DEFAULT 0,
    status VARCHAR(20) NOT NULL DEFAULT 'completed' CHECK (status IN ('pending', 'in_progress', 'completed', 'failed')),
    error_message TEXT,
    client_version VARCHAR(50),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create trigger for sync_records table
CREATE TRIGGER update_sync_records_updated_at
    BEFORE UPDATE ON sync_records
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Create indexes for sync_records table
CREATE INDEX idx_sync_records_user_id ON sync_records(user_id);
CREATE INDEX idx_sync_records_status ON sync_records(status);

-- Create foreign key
ALTER TABLE sync_records
    ADD CONSTRAINT fk_sync_records_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

-- +goose Down
ALTER TABLE sync_records DROP CONSTRAINT IF EXISTS fk_sync_records_user;
DROP TRIGGER IF EXISTS update_sync_records_updated_at ON sync_records;
DROP TABLE IF EXISTS sync_records;
