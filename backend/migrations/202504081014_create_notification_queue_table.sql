-- +goose Up
-- Create notification_queue table
CREATE TABLE notification_queue (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    notification_id UUID,
    user_id UUID NOT NULL,
    token_id UUID,
    type VARCHAR(20) NOT NULL CHECK (type IN ('push', 'email', 'sms')),
    priority INTEGER DEFAULT 5,
    payload JSONB NOT NULL,
    scheduled_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    sent_at TIMESTAMP WITH TIME ZONE,
    delivered_at TIMESTAMP WITH TIME ZONE,
    failed_at TIMESTAMP WITH TIME ZONE,
    error_message TEXT,
    retry_count INTEGER DEFAULT 0,
    max_retries INTEGER DEFAULT 3,
    status VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending', 'processing', 'sent', 'delivered', 'failed', 'cancelled')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create trigger for notification_queue table
CREATE TRIGGER update_notification_queue_updated_at
    BEFORE UPDATE ON notification_queue
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Create indexes for notification_queue table
CREATE INDEX idx_notification_queue_status ON notification_queue(status);
CREATE INDEX idx_notification_queue_scheduled ON notification_queue(scheduled_at) WHERE status = 'pending';
CREATE INDEX idx_notification_queue_user_id ON notification_queue(user_id);
CREATE INDEX idx_notification_queue_priority ON notification_queue(priority, created_at);

-- Create foreign keys
ALTER TABLE notification_queue
    ADD CONSTRAINT fk_notification_queue_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    ADD CONSTRAINT fk_notification_queue_token FOREIGN KEY (token_id) REFERENCES push_notification_tokens(id) ON DELETE SET NULL;

-- +goose Down
ALTER TABLE notification_queue DROP CONSTRAINT IF EXISTS fk_notification_queue_user;
ALTER TABLE notification_queue DROP CONSTRAINT IF EXISTS fk_notification_queue_token;
DROP TRIGGER IF EXISTS update_notification_queue_updated_at ON notification_queue;
DROP TABLE IF EXISTS notification_queue;
