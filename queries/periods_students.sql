-- name: ListStudentsForPeriod :many
SELECT period_id, student_id
FROM periods_students
WHERE period_id = @period_id AND
	deleted_at IS NULL
ORDER BY period_id DESC;

-- name: ListPeriodsForStudent :many
SELECT period_id, student_id
FROM periods_students
WHERE student_id = @student_id AND
	deleted_at IS NULL
ORDER BY student_id DESC;

-- name: AddStudentToPeriod :exec
INSERT INTO periods_students (period_id, student_id, created_at, updated_at, last_event_commit_position, last_event_prepare_position)
VALUES (@period_id, @student_id, @created_at, @created_at, @last_event_commit_position, @last_event_prepare_position)
ON CONFLICT (period_id, student_id) DO NOTHING;

-- name: RemoveStudentFromPeriod :exec
UPDATE periods_students
SET deleted_at = @deleted_at,
    updated_at = @deleted_at,
    last_event_commit_position = @last_event_commit_position,
    last_event_prepare_position = @last_event_prepare_position
WHERE period_id = @period_id AND student_id = @student_id;