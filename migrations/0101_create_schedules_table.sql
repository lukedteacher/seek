CREATE TABLE IF NOT EXISTS schedules (
	id TEXT PRIMARY KEY,
	title TEXT NOT NULL,
	teacher_id TEXT NOT NULL,
	created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
	deleted_at TEXT,
	last_event_commit_position INTEGER NOT NULL,
	last_event_prepare_position INTEGER NOT NULL,
	FOREIGN KEY(teacher_id) REFERENCES teachers(id) ON DELETE CASCADE
);