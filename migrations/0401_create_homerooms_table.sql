CREATE TABLE IF NOT EXISTS homerooms (
	id TEXT PRIMARY KEY,
	title TEXT NOT NULL,
	grades_bitmask INTEGER NOT NULL,
	location_id TEXT NOT NULL,
	image TEXT NOT NULL,
	last_event_commit_position INTEGER NOT NULL,
	last_event_prepare_position INTEGER NOT NULL,
	created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
	archived_at TEXT
);

CREATE INDEX IF NOT EXISTS homerooms_grades_bitmask_idx ON homerooms(grades_bitmask);
CREATE INDEX IF NOT EXISTS homerooms_location_id_idx ON homerooms(location_id);
CREATE INDEX IF NOT EXISTS homerooms_image_idx ON homerooms(image);