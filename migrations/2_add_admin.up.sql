INSERT INTO users (id, email, pass_hash)
VALUES (1, 'test@email.com', 'test')
ON CONFLICT DO NOTHING;
INSERT INTO admins (id, isAdmin)
VALUES (1, true)
ON CONFLICT DO NOTHING;
