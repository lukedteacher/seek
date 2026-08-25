-- name: GetRolesForEducator :many
SELECT
	role
FROM educator_roles
WHERE educator_id = @educator_id;

-- name: AddRoleToEducator :exec
INSERT INTO educator_roles ( 
	educator_id, 
	role,
	last_event_commit_position, 
	last_event_prepare_position,
	created_at, 
	updated_at
)
VALUES (
	@educator_id, 
	@role,
	@last_event_commit_position, 
	@last_event_prepare_position,
	@created_at, 
	@created_at
)
ON CONFLICT (educator_id, role) DO NOTHING;

-- name: RemoveRoleFromEducator :exec
DELETE FROM educator_roles
WHERE educator_id = @educator_id 
	AND role = @role;