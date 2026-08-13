CREATE TABLE IF NOT EXISTS auth_user (
	id TEXT PRIMARY KEY,
	email TEXT NOT NULL UNIQUE,
	email_verified INTEGER NOT NULL DEFAULT 0,
	image TEXT,
	role TEXT,
	user_registered_id TEXT NOT NULL,
	created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS auth_user_email_idx ON auth_user(email);
CREATE INDEX IF NOT EXISTS auth_user_user_registered_id_idx ON auth_user(user_registered_id);