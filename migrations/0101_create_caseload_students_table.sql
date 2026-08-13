CREATE TABLE IF NOT EXISTS caseload_students (
	educator_id TEXT NOT NULL,
	student_id TEXT NOT NULL,
	last_event_commit_position INTEGER NOT NULL,
	last_event_prepare_position INTEGER NOT NULL,
	created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (educator_id) REFERENCES educators(id) ON DELETE CASCADE,
	FOREIGN KEY (student_id) REFERENCES students(id) ON DELETE CASCADE,
	PRIMARY KEY (educator_id, student_id)
);