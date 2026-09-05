CREATE TABLE IF NOT EXISTS homerooms (
	id TEXT PRIMARY KEY,
	title TEXT NOT NULL,
	location_id TEXT NOT NULL,
	last_event_commit_position INTEGER NOT NULL,
	last_event_prepare_position INTEGER NOT NULL,
	created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
	archived_at TEXT
);