CREATE TABLE IF NOT EXISTS
  student_services (
    id TEXT PRIMARY KEY,
    student_id TEXT NOT NULL,
    service_type TEXT NOT NULL,
    indirect_minutes INTEGER,
    direct_minutes INTEGER,
    frequency_count INTEGER,
    frequency_type TEXT,
		location TEXT,
    start_date TEXT,
    end_date TEXT,
    provider TEXT,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
		deleted_at TEXT,
    FOREIGN KEY (student_id) REFERENCES students(id) ON DELETE CASCADE
  );

