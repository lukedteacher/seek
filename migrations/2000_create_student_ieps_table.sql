-- drop the old table and its index
DROP INDEX IF EXISTS student_ieps_student_id_idx;
DROP TABLE IF EXISTS student_ieps;

-- create the new table with updated columns
CREATE TABLE IF NOT EXISTS student_ieps (
	id TEXT PRIMARY KEY,
	student_id TEXT NOT NULL,
	plan_manager_id TEXT NOT NULL,
	disability_1 INTEGER NOT NULL,
	disability_2 INTEGER NOT NULL,
	federal_setting INTEGER NOT NULL,
	meeting_date TEXT NOT NULL,
	iep_due_date TEXT NOT NULL,
	last_eval_date TEXT NOT NULL,
	eval_due_date TEXT NOT NULL,
	amended_date TEXT NOT NULL DEFAULT '',
	iep_type INTEGER NOT NULL,
	special_transportation INTEGER NOT NULL,
	last_event_commit_position INTEGER NOT NULL,
	last_event_prepare_position INTEGER NOT NULL,
	created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
	archived_at TEXT,
	FOREIGN KEY (student_id) REFERENCES students(id) ON DELETE CASCADE
);

-- recreate the index on student_id
CREATE INDEX IF NOT EXISTS student_ieps_student_id_idx ON student_ieps(student_id);
CREATE INDEX IF NOT EXISTS student_ieps_plan_manager_id_idx ON student_ieps(plan_manager_id);