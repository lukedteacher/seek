-- name: GetPeriodStudent :one
SELECT period_id, student_id
FROM periods_students
WHERE period_id = @period_id AND student_id = @student_id;

-- name: ListPeriodsStudents :many
SELECT period_id, student_id
FROM periods_students;

-- name: ListStudentIDsForPeriod :many
SELECT period_id, student_id
FROM periods_students
WHERE period_id = @period_id
ORDER BY student_id DESC;

-- name: ListPeriodIDsForStudent :many
SELECT period_id, student_id
FROM periods_students
WHERE student_id = @student_id
ORDER BY period_id DESC;

-- name: AddStudentToPeriod :exec
INSERT INTO periods_students (
	period_id, 
	student_id, 
	last_event_commit_position, 
	last_event_prepare_position,
	created_at, 
	updated_at
)
VALUES (
	@period_id, 
	@student_id, 
	@last_event_commit_position, 
	@last_event_prepare_position,
	@created_at, 
	@created_at
)
ON CONFLICT (period_id, student_id) DO NOTHING;

-- name: RemoveStudentFromPeriod :exec
DELETE FROM periods_students
WHERE period_id = @period_id 
	AND student_id = @student_id;