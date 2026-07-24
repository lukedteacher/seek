-- name: GetStudent :one
SELECT 
	id, 
	given_name, 
	chosen_name, 
	family_name, 
	grade, 
	homeroom, 
	case_manager, 
	created_at, 
	updated_at,
	archived_at
FROM students
WHERE deleted_at IS NULL
	AND id = @id;

-- name: ListStudents :many
SELECT 
	id, 
	given_name, 
	chosen_name, 
	family_name, 
	grade, 
	homeroom, 
	case_manager, 
	created_at, 
	updated_at,
	archived_at
FROM students
WHERE deleted_at IS NULL
ORDER BY family_name DESC, given_name DESC;

-- name: CreateStudent :exec
INSERT INTO students (
	id, 
	given_name, 
	chosen_name, 
	family_name, 
	grade, 
	homeroom, 
	case_manager, 
	created_at, 
	updated_at,
	last_event_commit_position,
	last_event_prepare_position
)
VALUES (
	@id, 
	@given_name, 
	@chosen_name, 
	@family_name, 
	@grade, 
	@homeroom, 
	@case_manager, 
	@created_at, 
	@created_at, 
	@last_event_commit_position, 
	@last_event_prepare_position
)
ON CONFLICT (id) DO NOTHING;

-- name: UpdateStudent :exec
UPDATE students
SET given_name = @given_name,
	chosen_name = @chosen_name,
	family_name = @family_name,
	grade = @grade,
	homeroom = @homeroom,
	case_manager = @case_manager,
	updated_at = @updated_at,
	last_event_commit_position = @last_event_commit_position,
	last_event_prepare_position = @last_event_prepare_position
WHERE id = @id;

-- name: DeleteStudent :exec
DELETE FROM students
WHERE id = @id;
