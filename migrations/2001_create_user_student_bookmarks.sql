CREATE TABLE IF NOT EXISTS user_student_bookmarks (
	user_id TEXT NOT NULL,
	student_id TEXT NOT NULL,
	last_event_commit_position INTEGER NOT NULL,
	last_event_prepare_position INTEGER NOT NULL,
	added_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
	PRIMARY KEY (user_id, student_id),
	FOREIGN KEY (user_id) REFERENCES auth_user(id) ON DELETE CASCADE,
	FOREIGN KEY (student_id) REFERENCES students(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS user_student_bookmarks_user_id_idx ON user_student_bookmarks(user_id);
CREATE INDEX IF NOT EXISTS user_student_bookmarks_student_id_idx ON user_student_bookmarks(student_id);