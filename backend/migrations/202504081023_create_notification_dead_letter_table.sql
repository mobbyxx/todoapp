-- +goose Up
-- Create notification dead letter queue table
CREATE TABLE notification_dead_letter (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    type VARCHAR(20) NOT NULL,
    payload JSONB NOT NULL,
    error_message TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create index for querying dead letter entries
CREATE INDEX idx_notification_dead_letter_user_id ON notification_dead_letter(user_id);
CREATE INDEX idx_notification_dead_letter_created_at ON notification_dead_letter(created_at);

-- Add foreign key constraint
ALTER TABLE notification_dead_letter
    ADD CONSTRAINT fk_dead_letter_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

-- +goose Down
ALTER TABLE notification_dead_letter DROP CONSTRAINT IF EXISTS fk_dead_letter_user;
DROP INDEX IF EXISTS idx_notification_dead_letter_created_at;
DROP INDEX IF EXISTS idx_notification_dead_letter_user_id;
DROP TABLE IF EXISTS notification_dead_letter;
