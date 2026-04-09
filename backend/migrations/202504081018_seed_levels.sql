-- +goose Up
-- Seed levels data (Levels 1-8 with XP requirements: 0, 100, 300, 700, 1500, 3000, 6000, 12000)
INSERT INTO levels (level_number, name, min_points, max_points, rewards) VALUES
(1, 'Beginner', 0, 99, '["profile_badge", "basic_avatar"]'),
(2, 'Starter', 100, 299, '["custom_themes", "streak_freeze_token"]'),
(3, 'Rookie', 300, 699, '["priority_todos", "streak_freeze_token"]'),
(4, 'Apprentice', 700, 1499, '["advanced_filters", "streak_freeze_token"]'),
(5, 'Practitioner', 1500, 2999, '["collaboration_access", "streak_freeze_token"]'),
(6, 'Expert', 3000, 5999, '["analytics_dashboard", "streak_freeze_token", "custom_notifications"]'),
(7, 'Master', 6000, 11999, '["api_access", "streak_freeze_token", "early_access"]'),
(8, 'Legend', 12000, 2147483647, '["legend_badge", "streak_freeze_token", "lifetime_premium"]');

-- +goose Down
DELETE FROM levels WHERE level_number BETWEEN 1 AND 8;
