CREATE TABLE IF NOT EXISTS students (
	id TEXT PRIMARY KEY,
	first_name TEXT,
	chosen_name TEXT,
	last_name TEXT,
	grade INTEGER,
	homeroom TEXT,
	case_manager TEXT,
	last_event_commit_position INTEGER,
	last_event_prepare_position INTEGER,
	created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
	deleted_at TEXT
);

