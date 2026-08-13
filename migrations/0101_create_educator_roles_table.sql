CREATE TABLE IF NOT EXISTS educator_roles (
	educator_id TEXT NOT NULL,
	role TEXT NOT NULL,
	last_event_commit_position INTEGER NOT NULL,
	last_event_prepare_position INTEGER NOT NULL,
	created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
	archived_at TEXT,
	FOREIGN KEY (educator_id) REFERENCES educators(id) ON DELETE CASCADE,
	PRIMARY KEY (educator_id, role)
);

CREATE INDEX IF NOT EXISTS educator_roles_role_idx ON educator_roles(role);