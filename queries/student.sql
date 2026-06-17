Todo-- name: ListStudents :many
SELECT student_id, first_name, created_at, updated_at
FROM student_items
WHERE deleted_at IS NULL
ORDER BY created_at DESC, student_id DESC;

-- name: InsertCreatedStudent :exec
INSERT INTO student_items (student_id, user_registered_id, first_name, deleted_at, last_event_commit_position, last_event_prepare_position, created_at, updated_at)
VALUES (@student_id, @user_registered_id, @first_name, null, @last_event_commit_position, @last_event_prepare_position, @created_at, @created_at)
ON CONFLICT (student_id) DO NOTHING;

-- name: RenameStudent :exec
UPDATE student_items
SET first_name = @first_name,
    last_event_commit_position = @last_event_commit_position,
    last_event_prepare_position = @last_event_prepare_position,
    updated_at = @updated_at
WHERE student_id = @student_id;

-- name: DeleteStudent :exec
UPDATE student_items
SET deleted_at = @deleted_at,
    last_event_commit_position = @last_event_commit_position,
    last_event_prepare_position = @last_event_prepare_position,
    updated_at = @deleted_at
WHERE student_id = @student_id;
