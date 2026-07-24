CREATE TABLE IF NOT EXISTS 
	students (
		id TEXT PRIMARY KEY,
		given_name TEXT NOT NULL,
		chosen_name TEXT NOT NULL,
		family_name TEXT NOT NULL,
		grade INTEGER NOT NULL,
		homeroom TEXT NOT NULL,
		case_manager TEXT NOT NULL,
		created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
		archived_at TEXT,
		last_event_commit_position INTEGER NOT NULL,
		last_event_prepare_position INTEGER NOT NULL
	);

