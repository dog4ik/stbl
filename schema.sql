CREATE TABLE IF NOT EXISTS gateway_id_mapping (
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    gateway_id TEXT NOT NULL UNIQUE,
    token TEXT NOT NULL,
    merchant_private_key TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS token_cache (
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    credentials_hash TEXT NOT NULL UNIQUE,
    access_token TEXT NOT NULL,
    access_refreshed_at DATETIME DEFAULT CURRENT_TIMESTAMP NOT NULL,
    refresh_token TEXT NOT NULL,
    refresh_refreshed_at DATETIME DEFAULT CURRENT_TIMESTAMP NOT NULL
);
