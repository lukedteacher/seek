CREATE TABLE IF NOT EXISTS schedule_periods (
	id TEXT PRIMARY KEY,
	schedule_id TEXT NOT NULL,
	period_id TEXT NOT NULL,
	FOREIGN KEY(schedule_id) REFERENCES schedules(id),
	FOREIGN KEY(period_id) REFERENCES periods(id)
);