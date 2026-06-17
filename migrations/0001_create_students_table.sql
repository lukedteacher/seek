CREATE TABLE IF NOT EXISTS students (
	id TEXT PRIMARY KEY,
	first_name TEXT NOT NULL,
	chosen_name TEXT,
	last_name TEXT NOT NULL,
	grade INTEGER NOT NULL,
	homeroom TEXT NOT NULL,
	case_manager TEXT,
	last_event_commit_position INTEGER NOT NULL,
	last_event_prepare_position INTEGER NOT NULL,
	created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
	deleted_at TEXT
);

