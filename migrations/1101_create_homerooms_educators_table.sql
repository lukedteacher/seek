CREATE TABLE IF NOT EXISTS homerooms_educators (
	homeroom_id TEXT NOT NULL,
	educator_id TEXT NOT NULL,
	last_event_commit_position INTEGER NOT NULL,
	last_event_prepare_position INTEGER NOT NULL,
	created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
	PRIMARY KEY (homeroom_id, educator_id),
	FOREIGN KEY (homeroom_id) REFERENCES homerooms(id) ON DELETE CASCADE,
	FOREIGN KEY (educator_id) REFERENCES educators(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS homerooms_educators_educator_id_idx ON homerooms_educators(educator_id);