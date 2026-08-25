-- name: GetService :one
SELECT 
	id, 
	iep_id, 
	service_name,
	service_type, 
	indirect_minutes,
	direct_minutes,
	frequency_count,
	frequency_type,
	location_id,
	start_date,
	end_date,
	provider_id,
	created_at, 
	updated_at
FROM iep_services
WHERE archived_at IS NULL
	AND id = @id;

-- name: ListServices :many
SELECT id, 
	iep_id, 
	service_name,
	service_type, 
	indirect_minutes,
	direct_minutes,
	frequency_count,
	frequency_type,
	location_id,
	start_date,
	end_date,
	provider_id, 
	created_at, 
	updated_at
FROM iep_services
WHERE archived_at IS NULL
ORDER BY service_type DESC, service_name DESC;

-- name: ListServicesForIEP :many
SELECT id, 
	iep_id, 
	service_name,
	service_type, 
	indirect_minutes,
	direct_minutes,
	frequency_count,
	frequency_type,
	location_id,
	start_date,
	end_date,
	provider_id, 
	created_at, 
	updated_at
FROM iep_services
WHERE iep_id = @iep_id
	AND archived_at IS NULL
ORDER BY service_type DESC, service_name DESC;

-- name: AddServiceToIEP :exec
INSERT INTO iep_services (
	id, 
	iep_id,
	service_name,
	service_type, 
	indirect_minutes,
	direct_minutes,
	frequency_count,
	frequency_type,
	location_id,
	start_date,
	end_date,
	provider_id,
	last_event_commit_position,
	last_event_prepare_position,
	created_at, 
	updated_at
)
VALUES (
	@id, 
	@iep_id,
	@service_name,
	@service_type, 
	@indirect_minutes,
	@direct_minutes,
	@frequency_count,
	@frequency_type,
	@location_id,
	@start_date,
	@end_date,
	@provider_id,
	@last_event_commit_position,
	@last_event_prepare_position,
	@created_at, 
	@created_at
)
ON CONFLICT (id) DO NOTHING;

-- name: UpdateService :exec
UPDATE iep_services
SET
	service_name = @service_name,
	service_type = @service_type,
	indirect_minutes = @indirect_minutes,
	direct_minutes = @direct_minutes,
	frequency_count = @frequency_count,
	frequency_type = @frequency_type,
	location_id = @location_id,
	start_date = @start_date,
	end_date = @end_date,
	provider_id = @provider_id,
	last_event_commit_position = @last_event_commit_position,
	last_event_prepare_position = @last_event_prepare_position,
	updated_at = @updated_at
WHERE id = @id;

-- name: ArchiveService :exec
UPDATE iep_services
SET
	last_event_commit_position = @last_event_commit_position,
	last_event_prepare_position = @last_event_prepare_position,
	updated_at = @archived_at,
	archived_at = @archived_at
WHERE id = @id;

-- name: DeleteService :exec
DELETE FROM iep_services
WHERE id = @id;