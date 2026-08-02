CREATE TABLE IF NOT EXISTS educators_periods (
	period_id TEXT NOT NULL,
	educator_id TEXT NOT NULL,
	last_event_commit_position INTEGER NOT NULL,
	last_event_prepare_position INTEGER NOT NULL,
	created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
	PRIMARY KEY (period_id, educator_id),
	FOREIGN KEY (period_id) REFERENCES periods(id) ON DELETE CASCADE,
	FOREIGN KEY (educator_id) REFERENCES educators(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS educators_periods_educator_id_idx ON educators_periods (educator_id);