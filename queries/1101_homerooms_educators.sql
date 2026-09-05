-- name: GetHomeroomEducator :one
SELECT homeroom_id, educator_id
FROM homerooms_educators
WHERE homeroom_id = @homeroom_id AND educator_id = @educator_id;

-- name: ListHomeroomsEducators :many
SELECT homeroom_id, educator_id
FROM homerooms_educators;

-- name: ListEducatorIDsForHomeroom :many
SELECT homeroom_id, educator_id
FROM homerooms_educators
WHERE homeroom_id = @homeroom_id
ORDER BY educator_id ASC;

-- name: ListHomeroomIDsForEducator :many
SELECT homeroom_id, educator_id
FROM homerooms_educators
WHERE educator_id = @educator_id
ORDER BY homeroom_id ASC;

-- name: AddEducatorToHomeroom :exec
INSERT INTO homerooms_educators (
	homeroom_id, 
	educator_id, 
	last_event_commit_position, 
	last_event_prepare_position,
	created_at, 
	updated_at
)
VALUES (
	@homeroom_id, 
	@educator_id, 
	@last_event_commit_position, 
	@last_event_prepare_position,
	@created_at, 
	@created_at
)
ON CONFLICT (homeroom_id, educator_id) DO NOTHING;

-- name: RemoveEducatorFromHomeroom :exec
DELETE FROM homerooms_educators
WHERE homeroom_id = @homeroom_id 
	AND educator_id = @educator_id;