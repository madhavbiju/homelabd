CREATE TABLE IF NOT EXISTS tokens (
    id TEXT PRIMARY KEY,
    description TEXT,
    token_hash TEXT NOT NULL UNIQUE,
    permissions TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    expires_at DATETIME,
    revoked BOOLEAN DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS audit_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
    token_id TEXT,
    action TEXT NOT NULL,
    target TEXT,
    parameters TEXT,
    result TEXT,
    error_message TEXT,
    client_ip TEXT,
    FOREIGN KEY (token_id) REFERENCES tokens(id)
);
