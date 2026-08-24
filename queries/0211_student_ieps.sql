-- name: GetIEP :one
SELECT
	id,
	student_id,
	start_date,
	end_date,
	amended_date,
	created_at,
	updated_at,
	archived_at
FROM student_ieps
WHERE id = @id;

-- name: ListIEPs :many
SELECT
	id,
	student_id,
	start_date,
	end_date,
	amended_date,
	created_at,
	updated_at,
	archived_at
FROM student_ieps;

-- name: AddIEPToStudent :exec
INSERT INTO student_ieps (
	id, 
	student_id,
	start_date,
	end_date,
	amended_date,
	last_event_commit_position,
	last_event_prepare_position,
	created_at, 
	updated_at
)
VALUES (
	@id, 
	@student_id,
	@start_date,
	@end_date,
	@amended_date,
	@last_event_commit_position, 
	@last_event_prepare_position,
	@created_at, 
	@created_at
)
ON CONFLICT (id) DO NOTHING;

-- name: UpdateIEP :exec
UPDATE student_ieps
SET 
	start_date = @start_date,
	end_date = @end_date,
	amended_date = @amended_date,
	last_event_commit_position = @last_event_commit_position,
	last_event_prepare_position = @last_event_prepare_position,
	updated_at = @updated_at
WHERE id = @id;

-- name: ArchiveIEP :exec
UPDATE student_ieps
SET
	last_event_commit_position = @last_event_commit_position,
	last_event_prepare_position = @last_event_prepare_position,
	updated_at = @archived_at,
	archived_at = @archived_at
WHERE id = @id;

-- name: DeleteIEP :exec
DELETE FROM student_ieps
WHERE id = @id;