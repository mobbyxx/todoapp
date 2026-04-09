-- +goose Up
-- Seed default badges
INSERT INTO badges (name, description, icon_url, type, points_value, criteria) VALUES
(
    'First Todo',
    'Completed your first todo item',
    '/badges/first-todo.svg',
    'milestone',
    10,
    '{"type": "todo_completed", "count": 1}'
),
(
    'Week Warrior',
    'Completed at least one todo every day for 7 days',
    '/badges/week-warrior.svg',
    'achievement',
    50,
    '{"type": "streak", "days": 7}'
),
(
    'Century Club',
    'Completed 100 todos',
    '/badges/century-club.svg',
    'milestone',
    100,
    '{"type": "todo_completed", "count": 100}'
),
(
    'Connection Master',
    'Successfully connected with another user and completed 10 shared tasks',
    '/badges/connection-master.svg',
    'achievement',
    75,
    '{"type": "shared_tasks", "count": 10}'
),
(
    'Early Bird',
    'Completed 5 todos before 8 AM',
    '/badges/early-bird.svg',
    'special',
    25,
    '{"type": "early_completion", "count": 5, "before_hour": 8}'
);

-- +goose Down
DELETE FROM badges WHERE name IN ('First Todo', 'Week Warrior', 'Century Club', 'Connection Master', 'Early Bird');
