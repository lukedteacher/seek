-- name: GetEducator :one
SELECT 
	id, 
	given_name, 
	chosen_name, 
	family_name, 
	email, 
	role,
	created_at, 
	updated_at
FROM educators
WHERE archived_at IS NULL
	AND id = @id;

-- name: ListEducators :many
SELECT 
	id, 
	given_name, 
	chosen_name, 
	family_name, 
	email, 
	role,
	created_at, 
	updated_at
FROM educators
WHERE archived_at IS NULL
ORDER BY family_name DESC, given_name DESC;

-- name: CreateEducator :exec
INSERT INTO educators (
	id, 
	given_name, 
	chosen_name, 
	family_name, 
	email,
	role,
	last_event_commit_position, 
	last_event_prepare_position,
	created_at, 
	updated_at
)
VALUES (
	@id, 
	@given_name, 
	@chosen_name, 
	@family_name, 
	@email,
	@role,
	@last_event_commit_position, 
	@last_event_prepare_position,
	@created_at, 
	@created_at
)
ON CONFLICT (id) DO NOTHING;

-- name: UpdateEducator :exec
UPDATE educators
SET
	given_name = @given_name,
	chosen_name = @chosen_name,
	family_name = @family_name,
	email = @email,
	role = @role,
	last_event_commit_position = @last_event_commit_position,
	last_event_prepare_position = @last_event_prepare_position,
	updated_at = @updated_at
WHERE id = @id;

-- name: ArchiveEducator :exec
UPDATE educators
SET
	last_event_commit_position = @last_event_commit_position,
	last_event_prepare_position = @last_event_prepare_position,
	updated_at = @archived_at,
	archived_at = @archived_at
WHERE id = @id;

-- name: DeleteEducator :exec
DELETE FROM educators
WHERE id = @id;
