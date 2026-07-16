CREATE TABLE IF NOT EXISTS auth_user (
	id TEXT PRIMARY KEY,
	name TEXT NOT NULL,
	email TEXT NOT NULL UNIQUE,
	email_verified INTEGER NOT NULL DEFAULT 0,
	image TEXT,
	username TEXT UNIQUE,
	display_username TEXT,
	role TEXT,
	banned INTEGER DEFAULT 0,
	ban_reason TEXT,
	ban_expires TEXT,
	user_registered_id TEXT NOT NULL,
	created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS auth_user_email_idx ON auth_user (email);
CREATE INDEX IF NOT EXISTS auth_user_username_idx ON auth_user (username);
CREATE INDEX IF NOT EXISTS auth_user_user_registered_id_idx ON auth_user (user_registered_id);