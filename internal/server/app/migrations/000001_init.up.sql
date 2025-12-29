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
    key_encrypted BYTEA NOT NULL
);

CREATE TABLE IF NOT EXISTS items
(
    id         UUID PRIMARY KEY,
    user_id    UUID         NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    type       VARCHAR(32)  NOT NULL CHECK (type IN ('credential', 'text', 'binary', 'card')),
    title      VARCHAR(255) NOT NULL    DEFAULT '',
    metadata   TEXT         NOT NULL    DEFAULT '',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS encrypted_data
(
    id                 UUID PRIMARY KEY,
    item_id            UUID  NOT NULL REFERENCES items (id) ON DELETE CASCADE,
    data_encrypted     BYTEA NOT NULL,
    data_key_encrypted BYTEA NOT NULL,
    UNIQUE (item_id)
);

CREATE INDEX IF NOT EXISTS idx_users_username ON users (username);
CREATE INDEX IF NOT EXISTS idx_items_user_id ON items (user_id);
CREATE INDEX IF NOT EXISTS idx_items_updated_at ON items (updated_at);
CREATE INDEX IF NOT EXISTS idx_encrypted_data_item_id ON encrypted_data (item_id);

COMMIT;