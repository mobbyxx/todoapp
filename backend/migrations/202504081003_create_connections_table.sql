-- +goose Up
-- Create connections table
CREATE TABLE connections (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_a_id UUID NOT NULL,
    user_b_id UUID NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'accepted', 'blocked')),
    requested_by UUID NOT NULL,
    accepted_at TIMESTAMP WITH TIME ZONE,
    blocked_at TIMESTAMP WITH TIME ZONE,
    block_reason TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT chk_different_users CHECK (user_a_id != user_b_id),
    CONSTRAINT uq_connection_users UNIQUE (user_a_id, user_b_id)
);

-- Create trigger for connections table
CREATE TRIGGER update_connections_updated_at
    BEFORE UPDATE ON connections
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Create indexes for connections table
CREATE INDEX idx_connections_user_a ON connections(user_a_id);
CREATE INDEX idx_connections_user_b ON connections(user_b_id);
CREATE INDEX idx_connections_status ON connections(status);
CREATE INDEX idx_connections_pending ON connections(user_a_id, user_b_id, status) WHERE status = 'pending';

-- Create foreign keys
ALTER TABLE connections
    ADD CONSTRAINT fk_connections_user_a FOREIGN KEY (user_a_id) REFERENCES users(id) ON DELETE CASCADE,
    ADD CONSTRAINT fk_connections_user_b FOREIGN KEY (user_b_id) REFERENCES users(id) ON DELETE CASCADE;

-- +goose Down
ALTER TABLE connections DROP CONSTRAINT IF EXISTS fk_connections_user_a;
ALTER TABLE connections DROP CONSTRAINT IF EXISTS fk_connections_user_b;
DROP TRIGGER IF EXISTS update_connections_updated_at ON connections;
DROP TABLE IF EXISTS connections;
