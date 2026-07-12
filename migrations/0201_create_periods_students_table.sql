CREATE TABLE IF NOT EXISTS periods_students (
	period_id TEXT NOT NULL,
	student_id TEXT NOT NULL,
	created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
	deleted_at TEXT,
	last_event_commit_position INTEGER NOT NULL,
	last_event_prepare_position INTEGER NOT NULL,
	PRIMARY KEY (period_id, student_id),
	FOREIGN KEY (period_id) REFERENCES periods(id),
	FOREIGN KEY (student_id) REFERENCES student(id)
);