-- name: GetEducatorByID :one
SELECT 
	id, 
	given_name, 
	chosen_name, 
	family_name, 
	pronouns,
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
	pronouns,
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
	e.pronouns,
	e.email,
	e.username,
	e.created_at,
	e.updated_at,
	er.role
FROM educators e
LEFT JOIN educator_roles er ON e.id = er.educator_id AND er.archived_at IS NULL
WHERE e.archived_at IS NULL AND e.username = @username
ORDER BY er.role;

-- name: GetEducatorByUsernameWithCaseload :many
SELECT
	e.id AS educator_id,
	e.given_name,
	e.chosen_name,
	e.family_name,
	e.pronouns,
	e.email,
	e.username,
	e.created_at,
	e.updated_at,
	s.id AS student_id,
	s.marss_id,
	s.birthdate AS student_birthdate,
	s.given_name AS student_given_name,
	s.chosen_name AS student_chosen_name,
	s.family_name AS student_family_name,
	s.pronouns AS student_pronouns,
	s.email AS student_email,
	s.username AS student_username,
	s.grade,
	s.homeroom_id,
	s.plan_type,
	s.created_at AS student_created_at,
	s.updated_at AS student_updated_at
FROM educators e
LEFT JOIN caseload_students cl ON e.id = cl.educator_id
LEFT JOIN students s ON cl.student_id = s.id AND s.archived_at IS NULL
WHERE e.username = @username
  AND e.archived_at IS NULL
ORDER BY s.family_name DESC, s.given_name DESC;

-- name: ListEducators :many
SELECT 
	id, 
	given_name, 
	chosen_name, 
	family_name, 
	pronouns,
	email, 
	username,
	created_at, 
	updated_at
FROM educators
WHERE archived_at IS NULL
ORDER BY family_name ASC, given_name ASC;

-- name: ListEducatorsByRole :many
SELECT
	e.id,
	e.given_name,
	e.chosen_name,
	e.family_name,
	e.pronouns,
	e.email,
	e.username,
	e.created_at,
	e.updated_at,
	er.role
FROM educators e
INNER JOIN educator_roles er ON e.id = er.educator_id AND er.archived_at IS NULL
WHERE e.archived_at IS NULL
  AND er.role = @role
ORDER BY e.family_name DESC, e.given_name DESC, er.role ASC;

-- name: ListEducatorsWithRoles :many
SELECT
	e.id,
	e.given_name,
	e.chosen_name,
	e.family_name,
	e.pronouns,
	e.email,
	e.username,
	e.created_at,
	e.updated_at,
	er.role
FROM educators e
LEFT JOIN educator_roles er ON e.id = er.educator_id AND er.archived_at IS NULL
WHERE e.archived_at IS NULL
ORDER BY e.family_name COLLATE NOCASE ASC, e.given_name ASC, er.role ASC;

-- name: CreateEducator :exec
INSERT INTO educators (
	id, 
	given_name, 
	chosen_name, 
	family_name, 
	pronouns,
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
	@pronouns,
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
	pronouns = @pronouns,
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
