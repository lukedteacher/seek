-- name: GetStudentService :one
SELECT 
	id, 
	student_id, 
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
	updated_at,
	deleted_at
FROM student_services
WHERE deleted_at IS NULL
	AND id = @id;

-- name: ListStudentServices :many
SELECT 
	id, 
	student_id, 
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
FROM student_services
WHERE deleted_at IS NULL
ORDER BY student_id DESC, service_type DESC;

-- name: ListServicesForStudent :many
SELECT 
	id, 
	student_id, 
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
FROM student_services
WHERE student_id = @student_id
	AND deleted_at IS NULL
ORDER BY service_type DESC;

-- name: CreateStudentService :exec
INSERT INTO student_services (
	id, 
	student_id, 
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

-- name: UpdateStudentService :exec
UPDATE student_services
SET
	student_id = @student_id,
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

-- name: DeleteStudentService :exec
DELETE FROM student_services
WHERE id = @id;