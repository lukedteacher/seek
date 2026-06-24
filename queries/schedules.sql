-- name: GetSchedule :one
SELECT id, title, teacher_id, 
FROM schedules
WHERE deleted_at IS NULL
	AND id = @id;