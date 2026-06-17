-- name: ListStudents :many
SELECT id, first_name, chosen_name, last_name, grade, homeroom, case_manager, created_at, updated_at
FROM students
WHERE deleted_at IS NULL
ORDER BY created_at DESC, id DESC;

-- name: InsertCreatedStudent :exec
INSERT INTO students (id, first_name, chosen_name, last_name, grade, homeroom, case_manager, deleted_at, last_event_commit_position, last_event_prepare_position, created_at, updated_at)
VALUES (@id, @first_name, @chosen_name, @last_name, @grade, @homeroom, @case_manager, null, @last_event_commit_position, @last_event_prepare_position, @created_at, @created_at)
ON CONFLICT (id) DO NOTHING;

-- name: RenameStudent :exec
UPDATE students
SET first_name = @first_name,
	chosen_name = @chosen_name,
	last_name = @last_name,
	last_event_commit_position = @last_event_commit_position,
	last_event_prepare_position = @last_event_prepare_position,
	updated_at = @updated_at
WHERE id = @id;

-- name: DeleteStudent :exec
UPDATE students
SET deleted_at = @deleted_at,
	last_event_commit_position = @last_event_commit_position,
	last_event_prepare_position = @last_event_prepare_position,
	updated_at = @deleted_at
WHERE id = @id;
