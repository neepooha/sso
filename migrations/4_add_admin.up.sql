INSERT INTO users (id, email, pass_hash, is_admin)
VALUES (1, 'test@email.com', 'test', 1)
ON CONFLICT DO NOTHING;