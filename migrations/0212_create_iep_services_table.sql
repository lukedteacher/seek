CREATE TABLE IF NOT EXISTS iep_services (
	id TEXT PRIMARY KEY,
	iep_id TEXT NOT NULL,
	service_name TEXT NOT NULL,
	service_type TEXT NOT NULL,
	indirect_minutes INTEGER NOT NULL,
	direct_minutes INTEGER NOT NULL,
	frequency_count INTEGER NOT NULL,
	frequency_type TEXT NOT NULL,
	location_id TEXT NOT NULL,
	start_date TEXT NOT NULL,
	end_date TEXT NOT NULL,
	provider_id TEXT NOT NULL,
	last_event_commit_position INTEGER NOT NULL,
	last_event_prepare_position INTEGER NOT NULL,
	created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
	archived_at TEXT,
	FOREIGN KEY (iep_id) REFERENCES student_ieps(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS iep_services_iep_id_idx ON iep_services(iep_id);
CREATE INDEX IF NOT EXISTS iep_services_service_type_idx ON iep_services(service_type);
CREATE INDEX IF NOT EXISTS iep_services_location_id_idx ON iep_services(location_id);
CREATE INDEX IF NOT EXISTS iep_services_provider_id_idx ON iep_services(provider_id);