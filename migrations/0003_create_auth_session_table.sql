CREATE TABLE IF NOT EXISTS auth_session (
	id TEXT PRIMARY KEY,
	expires_at TEXT NOT NULL,
	token TEXT NOT NULL UNIQUE,
	ip_address TEXT,
	user_agent TEXT,
	user_id TEXT NOT NULL REFERENCES auth_user(id) ON DELETE CASCADE,
	impersonated_by TEXT,
	created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS session_user_id_idx ON auth_session (user_id);