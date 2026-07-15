-- name: GetTeacher :one
SELECT id, first_name, chosen_name, last_name, created_at, updated_at
FROM teachers
WHERE deleted_at IS NULL
	AND id = @id;

-- name: ListTeachers :many
SELECT id, first_name, chosen_name, last_name, created_at, updated_at
FROM teachers
WHERE deleted_at IS NULL
ORDER BY last_name DESC, id DESC;

-- name: CreateTeacher :exec
INSERT INTO teachers (id, first_name, chosen_name, last_name, created_at, updated_at, last_event_commit_position, last_event_prepare_position)
VALUES (@id, @first_name, @chosen_name, @last_name, @created_at, @created_at, @last_event_commit_position, @last_event_prepare_position)
ON CONFLICT (id) DO NOTHING;

-- name: UpdateTeacher :exec
UPDATE teachers
SET first_name = @first_name,
	chosen_name = @chosen_name,
	last_name = @last_name,
	updated_at = @updated_at,
	last_event_commit_position = @last_event_commit_position,
	last_event_prepare_position = @last_event_prepare_position
WHERE id = @id;

-- name: DeleteTeacher :exec
DELETE FROM teachers
WHERE id = @id;
