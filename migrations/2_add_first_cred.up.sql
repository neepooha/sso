INSERT INTO users (email, pass_hash)
VALUES ('test@email.com', '1')
ON CONFLICT DO NOTHING;
INSERT INTO apps (name, secret)
VALUES ('test', 'test-secret')
ON CONFLICT DO NOTHING;
INSERT INTO admins (uid, app_id)
VALUES (1, 1)
ON CONFLICT DO NOTHING;