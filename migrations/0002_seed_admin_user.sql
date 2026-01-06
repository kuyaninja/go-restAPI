INSERT INTO users (id, name, email, password_hash, created_at, updated_at)
VALUES (
    '00000000-0000-0000-0000-000000000001',
    'Admin User',
    'admin@example.com',
    '$2a$10$cWyQ5mv3errPApfPQJKDmuOm4oETcldcdByRThiLgXhFW0RclWXh2',
    NOW(),
    NOW()
)
ON DUPLICATE KEY UPDATE
    name = VALUES(name),
    email = VALUES(email),
    password_hash = VALUES(password_hash),
    updated_at = VALUES(updated_at);
