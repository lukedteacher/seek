CREATE TABLE IF NOT EXISTS 
	student_ieps (
		id TEXT PRIMARY KEY,
		student_id TEXT NOT NULL UNIQUE,
		case_manager_id TEXT NOT NULL DEFAULT '',
		start_date TEXT NOT NULL,
		end_date TEXT NOT NULL,
		amended_date TEXT NOT NULL DEFAULT '',
		last_event_commit_position INTEGER NOT NULL,
		last_event_prepare_position INTEGER NOT NULL,
		created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
		archived_at TEXT,
		FOREIGN KEY (student_id) REFERENCES students(id) ON DELETE CASCADE,
		FOREIGN KEY (case_manager_id) REFERENCES educators(id) ON DELETE SET DEFAULT
	);

CREATE INDEX IF NOT EXISTS student_ieps_student_id_idx ON student_ieps(student_id);
CREATE INDEX IF NOT EXISTS student_ieps_case_manager_id_idx ON student_ieps(case_manager_id);