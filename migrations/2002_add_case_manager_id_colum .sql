ALTER TABLE students ADD COLUMN case_manager_id TEXT NOT NULL DEFAULT '';
CREATE INDEX IF NOT EXISTS students_case_manager_id_idx ON students(case_manager_id);