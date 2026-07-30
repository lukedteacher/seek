CREATE TABLE IF NOT EXISTS 
	students (
		id TEXT PRIMARY KEY,
		marss_id TEXT NOT NULL,
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

CREATE INDEX IF NOT EXISTS students_marss_id_idx ON students (marss_id);
CREATE INDEX IF NOT EXISTS students_given_name_idx ON students (given_name);
CREATE INDEX IF NOT EXISTS students_chosen_name_idx ON students (chosen_name);
CREATE INDEX IF NOT EXISTS students_family_name_idx ON students (family_name);
CREATE INDEX IF NOT EXISTS students_grade_idx ON students (grade);