CREATE TABLE IF NOT EXISTS educators (
	id TEXT PRIMARY KEY,
	given_name TEXT NOT NULL,
	chosen_name TEXT NOT NULL,
	family_name TEXT NOT NULL,
	email TEXT NOT NULL,
	username TEXT NOT NULL,
	last_event_commit_position INTEGER NOT NULL,
	last_event_prepare_position INTEGER NOT NULL,
	created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
	archived_at TEXT
);

CREATE INDEX IF NOT EXISTS educators_given_name_idx ON educators(given_name);
CREATE INDEX IF NOT EXISTS educators_chosen_name_idx ON educators(chosen_name);
CREATE INDEX IF NOT EXISTS educators_family_name_idx ON educators(family_name);
CREATE INDEX IF NOT EXISTS educators_email_idx ON educators(email);
CREATE INDEX IF NOT EXISTS educators_username_idx ON educators(username);