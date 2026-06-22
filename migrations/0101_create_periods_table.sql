CREATE TABLE
  periods (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    start_time TEXT NOT NULL,
    duration INT NOT NULL,
    days INT NOT NULL,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TEXT,
		last_event_commit_position INTEGER NOT NULL,
		last_event_prepare_position INTEGER NOT NULL
  );