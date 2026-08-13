-- name: GetEducatorByID :one
SELECT 
	id, 
	given_name, 
	chosen_name, 
	family_name, 
	email, 
	username,
	created_at, 
	updated_at
FROM educators
WHERE archived_at IS NULL
	AND id = @id;

-- name: GetEducatorByUsername :one
SELECT 
	id, 
	given_name, 
	chosen_name, 
	family_name, 
	email, 
	username,
	created_at, 
	updated_at
FROM educators
WHERE archived_at IS NULL
	AND username = @username;

-- name: GetEducatorWithRolesByUsername :many
SELECT
	e.id,
	e.given_name,
	e.chosen_name,
	e.family_name,
	e.email,
	e.username,
	e.created_at,
	e.updated_at,
	er.role
FROM educators e
LEFT JOIN educator_roles er ON e.id = er.educator_id AND er.archived_at IS NULL
WHERE e.archived_at IS NULL AND e.username = @username
ORDER BY er.role;

-- name: ListEducators :many
SELECT 
	id, 
	given_name, 
	chosen_name, 
	family_name, 
	email, 
	username,
	created_at, 
	updated_at
FROM educators
WHERE archived_at IS NULL
ORDER BY family_name DESC, given_name DESC;

-- name: ListEducatorsWithRoles :many
SELECT
	e.id,
	e.given_name,
	e.chosen_name,
	e.family_name,
	e.email,
	e.username,
	e.created_at,
	e.updated_at,
	er.role
FROM educators e
LEFT JOIN educator_roles er ON e.id = er.educator_id AND er.archived_at IS NULL
WHERE e.archived_at IS NULL
ORDER BY e.family_name DESC, e.given_name DESC, er.role ASC;

-- name: CreateEducator :exec
INSERT INTO educators (
	id, 
	given_name, 
	chosen_name, 
	family_name, 
	email,
	username,
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
	@username,
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
	username = @username,
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
