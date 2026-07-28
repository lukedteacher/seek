CREATE TABLE IF NOT EXISTS teachers (
	id TEXT PRIMARY KEY,
	given_name TEXT NOT NULL,
	chosen_name TEXT,
	family_name TEXT NOT NULL,
	created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
	deleted_at TEXT,
	last_event_commit_position INTEGER NOT NULL,
	last_event_prepare_position INTEGER NOT NULL
);

