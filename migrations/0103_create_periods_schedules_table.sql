CREATE TABLE IF NOT EXISTS periods_schedules (
	period_id TEXT NOT NULL,
	schedule_id TEXT NOT NULL,
	created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
	last_event_commit_position INTEGER NOT NULL,
	last_event_prepare_position INTEGER NOT NULL,
	PRIMARY KEY (period_id, schedule_id),
	FOREIGN KEY (period_id) REFERENCES periods(id),
	FOREIGN KEY (schedule_id) REFERENCES schedules(id)
);