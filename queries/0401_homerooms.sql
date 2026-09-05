-- name: GetHomeroom :one
SELECT
	id, 
	title, 
	location_id,
	created_at, 
	updated_at
FROM homerooms
WHERE archived_at IS NULL
	AND id = @id;

-- name: GetHomeroomWithIDs :one
SELECT
	h.id,
	h.title,
	h.location_id,
	h.created_at,
	h.updated_at,
	CAST(
		COALESCE(
			(SELECT json_group_array(he.educator_id) FROM homerooms_educators he WHERE he.homeroom_id = h.id),
			'[]'
		) AS TEXT
	) AS educator_ids,
	CAST(
		COALESCE(
			(SELECT json_group_array(hs.student_id) FROM homerooms_students hs WHERE hs.homeroom_id = h.id),
			'[]'
		) AS TEXT
	) AS student_ids
FROM homerooms h
WHERE h.id = @id;

-- name: ListHomerooms :many
SELECT
	id, 
	title, 
	location_id,
	created_at, 
	updated_at
FROM homerooms
WHERE archived_at IS NULL
ORDER BY title ASC;

-- name: ListHomeroomsWithIDs :many
SELECT
	h.id,
	h.title,
	h.location_id,
	h.created_at,
	h.updated_at,
	CAST(
		COALESCE(
			(SELECT json_group_array(he.educator_id) FROM homerooms_educators he WHERE he.homeroom_id = h.id),
			'[]'
		) AS TEXT
	) AS educator_ids,
	CAST(
		COALESCE(
			(SELECT json_group_array(hs.student_id) FROM homerooms_students hs WHERE hs.homeroom_id = h.id),
			'[]'
		) AS TEXT
	) AS student_ids
FROM homerooms h;

-- name: CreateHomeroom :exec
INSERT INTO homerooms (
	id, 
	title, 
	location_id,
	last_event_commit_position, 
	last_event_prepare_position,
	created_at, 
	updated_at
)
VALUES (
	@id, 
	@title, 
	@location_id,
	@last_event_commit_position, 
	@last_event_prepare_position,
	@created_at, 
	@created_at
)
ON CONFLICT (id) DO NOTHING;

-- name: UpdateHomeroom :exec
UPDATE homerooms
SET
	title = @title,
	location_id = @location_id,
	last_event_commit_position = @last_event_commit_position,
	last_event_prepare_position = @last_event_prepare_position,
	updated_at = @updated_at
WHERE id = @id;

-- name: ArchiveHomeroom :exec
UPDATE homerooms
SET
	last_event_commit_position = @last_event_commit_position,
	last_event_prepare_position = @last_event_prepare_position,
	updated_at = @archived_at,
	archived_at = @archived_at
WHERE id = @id;

-- name: DeleteHomeroom :exec
DELETE FROM homerooms
WHERE id = @id;