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
	created_at, 
	updated_at, 
	last_event_commit_position, 
	last_event_prepare_position
)
VALUES (
	@id, 
	@title, 
	@service_type,
	@start_time, 
	@duration, 
	@days_bitmask, 
	@created_at, 
	@created_at, 
	@last_event_commit_position, 
	@last_event_prepare_position
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
	updated_at = @updated_at,
	last_event_commit_position = @last_event_commit_position,
	last_event_prepare_position = @last_event_prepare_position
WHERE id = @id;

-- name: ArchivePeriod :exec
UPDATE periods
SET
	updated_at = @archived_at,
	archived_at = @archived_at,
	last_event_commit_position = @last_event_commit_position,
	last_event_prepare_position = @last_event_prepare_position
WHERE id = @id;

-- name: DeletePeriod :exec
DELETE FROM periods
WHERE id = @id;