CREATE TABLE IF NOT EXISTS schedule_periods (
	schedule_id TEXT NOT NULL,
	period_id TEXT NOT NULL,
	created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
	deleted_at TEXT,
	last_event_commit_position INTEGER NOT NULL,
	last_event_prepare_position INTEGER NOT NULL,
	PRIMARY KEY (schedule_id, period_id),
	FOREIGN KEY (schedule_id) REFERENCES schedules(id),
	FOREIGN KEY (period_id) REFERENCES periods(id)
);