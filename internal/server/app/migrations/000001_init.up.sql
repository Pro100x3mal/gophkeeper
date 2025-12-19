BEGIN TRANSACTION;

CREATE TABLE IF NOT EXISTS users
(
    id            UUID PRIMARY KEY,
    username      VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255)        NOT NULL,
    created_at    TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at    TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS encryption_keys
(
    user_id       UUID PRIMARY KEY REFERENCES users (id) ON DELETE CASCADE,
    key_encrypted BYTEA NOT NULL,
    created_at    TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at    TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS items
(
    id                 UUID PRIMARY KEY,
    user_id            UUID  NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    type               TEXT  NOT NULL,
    title              TEXT  NOT NULL           DEFAULT '',
    metadata           JSONB NOT NULL           DEFAULT '{}',
    data_encrypted     BYTEA,
    data_key_encrypted BYTEA,
    created_at         TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at         TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_users_username ON users (username);
CREATE INDEX IF NOT EXISTS idx_items_user_id ON items (user_id);

COMMIT;