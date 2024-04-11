INSERT INTO users (id, email, pass_hash)
VALUES (1, 'test@email.com', 'test')
ON CONFLICT DO NOTHING;
INSERT INTO admins (uid, isAdmin)
VALUES (1, true)
ON CONFLICT DO NOTHING;
