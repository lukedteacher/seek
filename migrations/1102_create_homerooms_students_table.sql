CREATE TABLE IF NOT EXISTS homerooms_students (
	homeroom_id TEXT NOT NULL,
	student_id TEXT NOT NULL,
	last_event_commit_position INTEGER NOT NULL,
	last_event_prepare_position INTEGER NOT NULL,
	created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
	PRIMARY KEY (homeroom_id, student_id),
	FOREIGN KEY (homeroom_id) REFERENCES homerooms(id) ON DELETE CASCADE,
	FOREIGN KEY (student_id) REFERENCES students(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS homerooms_students_student_id_idx ON homerooms_students(student_id);