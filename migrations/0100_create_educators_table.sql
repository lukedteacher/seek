CREATE TABLE IF NOT EXISTS educators (
	id TEXT PRIMARY KEY,
	given_name TEXT NOT NULL,
	chosen_name TEXT NOT NULL,
	family_name TEXT NOT NULL,
	email TEXT NOT NULL,
	role TEXT NOT NULL,
	last_event_commit_position INTEGER NOT NULL,
	last_event_prepare_position INTEGER NOT NULL,
	created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
	archived_at TEXT
);

