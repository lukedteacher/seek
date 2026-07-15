-- name: GetSchedule :one
SELECT id, title, teacher_id
FROM schedules
WHERE deleted_at IS NULL
	AND id = @id;

-- name: ListSchedules :many
SELECT id, title, teacher_id, created_at, updated_at
FROM schedules
WHERE deleted_at IS NULL
ORDER BY title DESC, id DESC;

-- name: CreateSchedule :exec
INSERT INTO schedules (id, title, teacher_id, created_at, updated_at, last_event_commit_position, last_event_prepare_position)
VALUES (@id, @title, @teacher_id, @created_at, @created_at, @last_event_commit_position, @last_event_prepare_position)
ON CONFLICT (id) DO NOTHING;

-- name: UpdateSchedule :exec
UPDATE schedules
SET title = @title,
	teacher_id = @teacher_id,
	updated_at = @updated_at,
	last_event_commit_position = @last_event_commit_position,
	last_event_prepare_position = @last_event_prepare_position
WHERE id = @id;

-- name: DeleteSchedule :exec
DELETE FROM schedules
WHERE id = @id;