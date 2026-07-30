-- name: GetPeriod :one
SELECT
	id, 
	title, 
	service_type,
	start_time, 
	duration, 
	days_bitmask, 
	created_at, 
	updated_at
FROM periods
WHERE archived_at IS NULL
	AND id = @id;

-- name: ListPeriods :many
SELECT
	id, 
	title, 
	service_type,
	start_time, 
	duration, 
	days_bitmask, 
	created_at, 
	updated_at
FROM periods
WHERE archived_at IS NULL
ORDER BY service_type DESC, title DESC;

-- name: CreatePeriod :exec
INSERT INTO periods (
	id, 
	title, 
	service_type,
	start_time, 
	duration, 
	days_bitmask, 
	last_event_commit_position, 
	last_event_prepare_position,
	created_at, 
	updated_at
)
VALUES (
	@id, 
	@title, 
	@service_type,
	@start_time, 
	@duration, 
	@days_bitmask, 
	@last_event_commit_position, 
	@last_event_prepare_position,
	@created_at, 
	@created_at
)
ON CONFLICT (id) DO NOTHING;

-- name: UpdatePeriod :exec
UPDATE periods
SET
	title = @title,
	service_type = @service_type,
	start_time = @start_time,
	duration = @duration,
	days_bitmask = @days_bitmask,
	last_event_commit_position = @last_event_commit_position,
	last_event_prepare_position = @last_event_prepare_position,
	updated_at = @updated_at
WHERE id = @id;

-- name: ArchivePeriod :exec
UPDATE periods
SET
	last_event_commit_position = @last_event_commit_position,
	last_event_prepare_position = @last_event_prepare_position,
	updated_at = @archived_at,
	archived_at = @archived_at
WHERE id = @id;

-- name: DeletePeriod :exec
DELETE FROM periods
WHERE id = @id;