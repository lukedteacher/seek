-- name: GetPeriod :one
SELECT id, title, start_time, duration, days
FROM periods
WHERE deleted_at IS NULL
	AND id = @id;

-- name: ListPeriods :many
SELECT id, title, start_time, duration, days
FROM periods
WHERE deleted_at IS NULL
ORDER BY title DESC, id DESC;

-- name: InsertCreatedPeriod :exec
INSERT INTO periods (id, title, start_time, duration, days, created_at, updated_at, last_event_commit_position, last_event_prepare_position)
VALUES (@id, @title, @start_time, @duration, @days, @created_at, @created_at, @last_event_commit_position, @last_event_prepare_position)
ON CONFLICT (id) DO NOTHING;

-- name: RenamePeriod :exec
UPDATE periods
SET title = @title,
	updated_at = @updated_at,
	last_event_commit_position = @last_event_commit_position,
	last_event_prepare_position = @last_event_prepare_position
WHERE id = @id;

-- name: DeletePeriod :exec
UPDATE periods
SET deleted_at = @deleted_at,
	updated_at = @deleted_at,
	last_event_commit_position = @last_event_commit_position,
	last_event_prepare_position = @last_event_prepare_position
WHERE id = @id;
