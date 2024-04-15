CREATE TABLE IF NOT EXISTS users
(
    id        SERIAL PRIMARY KEY,
    email     TEXT NOT NULL UNIQUE,
    pass_hash TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_email ON users (email);

CREATE TABLE IF NOT EXISTS apps
(
    id     SERIAL PRIMARY KEY,
    name   TEXT NOT NULL UNIQUE,
    secret TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS admins
(
    uid      INTEGER REFERENCES users (id),
    app_id   INTEGER REFERENCES apps (id)
);

CREATE TABLE IF NOT EXISTS creators
(
    uid      INTEGER REFERENCES users (id),
    app_id   INTEGER REFERENCES apps (id)
);