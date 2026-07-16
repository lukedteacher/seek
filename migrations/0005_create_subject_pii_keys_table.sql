CREATE TABLE IF NOT EXISTS subject_pii_keys (
	subject_id TEXT PRIMARY KEY,
	encrypted_data_key TEXT NOT NULL,
	encryption_nonce TEXT NOT NULL,
	key_version TEXT NOT NULL,
	created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);