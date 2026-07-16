CREATE TABLE IF NOT EXISTS profile_stats (
	user_id TEXT PRIMARY KEY,
	name TEXT,
	username TEXT,
	email TEXT,
	image TEXT,
	bio TEXT,
	header_image_url TEXT,
	last_event_commit_position INTEGER NOT NULL,
	last_event_prepare_position INTEGER NOT NULL,
	created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS profile_stats_user_id_idx ON profile_stats (user_id);
CREATE INDEX IF NOT EXISTS profile_stats_username_idx ON profile_stats (username);