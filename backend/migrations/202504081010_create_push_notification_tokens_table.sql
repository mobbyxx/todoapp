-- +goose Up
-- Create push_notification_tokens table
CREATE TABLE push_notification_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    token TEXT NOT NULL,
    platform VARCHAR(10) NOT NULL CHECK (platform IN ('ios', 'android', 'web')),
    device_info JSONB DEFAULT '{}',
    is_active BOOLEAN DEFAULT true,
    last_used_at TIMESTAMP WITH TIME ZONE,
    expires_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT uq_user_platform_token UNIQUE (user_id, platform, token)
);

-- Create trigger for push_notification_tokens table
CREATE TRIGGER update_push_notification_tokens_updated_at
    BEFORE UPDATE ON push_notification_tokens
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Create indexes for push_notification_tokens table
CREATE INDEX idx_push_tokens_user_id ON push_notification_tokens(user_id);
CREATE INDEX idx_push_tokens_platform ON push_notification_tokens(platform);
CREATE INDEX idx_push_tokens_active ON push_notification_tokens(is_active) WHERE is_active = true;

-- Create foreign key
ALTER TABLE push_notification_tokens
    ADD CONSTRAINT fk_push_tokens_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

-- +goose Down
ALTER TABLE push_notification_tokens DROP CONSTRAINT IF EXISTS fk_push_tokens_user;
DROP TRIGGER IF EXISTS update_push_notification_tokens_updated_at ON push_notification_tokens;
DROP TABLE IF EXISTS push_notification_tokens;
