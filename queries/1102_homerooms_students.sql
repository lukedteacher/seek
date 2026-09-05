-- name: GetHomeroomStudent :one
SELECT homeroom_id, student_id
FROM homerooms_students
WHERE homeroom_id = @homeroom_id AND student_id = @student_id;

-- name: ListHomeroomsStudents :many
SELECT homeroom_id, student_id
FROM homerooms_students;

-- name: ListStudentIDsForHomeroom :many
SELECT homeroom_id, student_id
FROM homerooms_students
WHERE homeroom_id = @homeroom_id
ORDER BY student_id ASC;

-- name: ListHomeroomIDsForStudent :many
SELECT homeroom_id, student_id
FROM homerooms_students
WHERE student_id = @student_id
ORDER BY homeroom_id ASC;

-- name: AddStudentToHomeroom :exec
INSERT INTO homerooms_students (
	homeroom_id, 
	student_id, 
	last_event_commit_position, 
	last_event_prepare_position,
	created_at, 
	updated_at
)
VALUES (
	@homeroom_id, 
	@student_id, 
	@last_event_commit_position, 
	@last_event_prepare_position,
	@created_at, 
	@created_at
)
ON CONFLICT (homeroom_id, student_id) DO NOTHING;

-- name: RemoveStudentFromHomeroom :exec
DELETE FROM homerooms_students
WHERE homeroom_id = @homeroom_id 
	AND student_id = @student_id;