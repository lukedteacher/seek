-- name: GetIEPService :one
SELECT 
	id, 
	student_id, 
	service_name,
	service_type, 
	indirect_minutes,
	direct_minutes,
	frequency_count,
	frequency_type,
	location,
	start_date,
	end_date,
	provider,
	created_at, 
	updated_at
FROM iep_services
WHERE archived_at IS NULL
	AND id = @id;

-- name: ListIEPServices :many
SELECT id, 
	student_id, 
	service_name,
	service_type, 
	indirect_minutes,
	direct_minutes,
	frequency_count,
	frequency_type,
	location,
	start_date,
	end_date,
	provider, 
	created_at, 
	updated_at
FROM iep_services
WHERE archived_at IS NULL
ORDER BY student_id DESC, service_type DESC;

-- name: ListIEPServicesForStudent :many
SELECT id, 
	student_id, 
	service_name,
	service_type, 
	indirect_minutes,
	direct_minutes,
	frequency_count,
	frequency_type,
	location,
	start_date,
	end_date,
	provider, 
	created_at, 
	updated_at
FROM iep_services
WHERE student_id = @student_id
	AND archived_at IS NULL
ORDER BY service_type DESC;

-- name: AddIEPServiceToStudent :exec
INSERT INTO iep_services (
	id, 
	student_id, 
	service_name,
	service_type, 
	indirect_minutes,
	direct_minutes,
	frequency_count,
	frequency_type,
	location,
	start_date,
	end_date,
	provider,
	created_at, 
	updated_at
)
VALUES (
	@id, 
	@student_id, 
	@service_name,
	@service_type, 
	@indirect_minutes,
	@direct_minutes,
	@frequency_count,
	@frequency_type,
	@location,
	@start_date,
	@end_date,
	@provider,
	@created_at, 
	@created_at
)
ON CONFLICT (id) DO NOTHING;

-- name: UpdateIEPService :exec
UPDATE iep_services
SET
	student_id = @student_id,
	service_name = @service_name,
	service_type = @service_type,
	indirect_minutes = @indirect_minutes,
	direct_minutes = @direct_minutes,
	frequency_count = @frequency_count,
	frequency_type = @frequency_type,
	location = @location,
	start_date = @start_date,
	end_date = @end_date,
	provider = @provider,
	updated_at = @updated_at
WHERE id = @id;

-- name: ArchiveIEPService :exec
UPDATE iep_services
SET
	updated_at = @archived_at,
	archived_at = @archived_at,
	last_event_commit_position = @last_event_commit_position,
	last_event_prepare_position = @last_event_prepare_position
WHERE id = @id;

-- name: DeleteIEPService :exec
DELETE FROM iep_services
WHERE id = @id;