-- name: ListSchedulePeriods :many
SELECT id, schedule_id, period_id
FROM schedule_periods
WHERE schedule_id = @schedule_id AND
	deleted_at IS NULL
ORDER BY id DESC;

-- name: AddSchedulePeriod :exec
INSERT INTO schedule_periods (id, schedule_id, period_id, created_at, updated_at, last_event_commit_position, last_event_prepare_position)
VALUES (@id, @schedule_id, @period_id, @created_at, @created_at, @last_event_commit_position, @last_event_prepare_position)
ON CONFLICT (id) DO NOTHING;

-- name: RemoveSchedulePeriod :exec
UPDATE schedule_periods
SET deleted_at = @deleted_at,
	updated_at = @deleted_at,
	last_event_commit_position = @last_event_commit_position,
	last_event_prepare_position = @last_event_prepare_position
WHERE id = @id;