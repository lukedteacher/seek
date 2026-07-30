-- name: GetStudentByID :one
SELECT 
	id, 
	marss_id,
	given_name, 
	chosen_name, 
	family_name, 
	email,
	username,
	grade, 
	homeroom, 
	case_manager, 
	created_at, 
	updated_at,
	archived_at
FROM students
WHERE archived_at IS NULL
	AND id = @id;
	
-- name: GetStudentByUsername :one
SELECT 
	id, 
	marss_id,
	given_name, 
	chosen_name, 
	family_name, 
	email,
	username,
	grade, 
	homeroom, 
	case_manager, 
	created_at, 
	updated_at,
	archived_at
FROM students
WHERE archived_at IS NULL
	AND username = @username;

-- name: ListStudents :many
SELECT 
	id, 
	marss_id,
	given_name, 
	chosen_name, 
	family_name, 
	email,
	username,
	grade, 
	homeroom, 
	case_manager, 
	created_at, 
	updated_at,
	archived_at
FROM students
WHERE archived_at IS NULL
ORDER BY family_name DESC, given_name DESC;

-- name: CreateStudent :exec
INSERT INTO students (
	id, 
	marss_id,
	given_name, 
	chosen_name, 
	family_name, 
	email,
	username,
	grade, 
	homeroom, 
	case_manager, 
	last_event_commit_position,
	last_event_prepare_position,
	created_at, 
	updated_at
)
VALUES (
	@id, 
	@marss_id,
	@given_name, 
	@chosen_name, 
	@family_name, 
	@email,
	@username,
	@grade, 
	@homeroom, 
	@case_manager, 
	@last_event_commit_position, 
	@last_event_prepare_position,
	@created_at, 
	@created_at
)
ON CONFLICT (id) DO NOTHING;

-- name: UpdateStudent :exec
UPDATE students
SET 
	marss_id = @marss_id,
	given_name = @given_name,
	chosen_name = @chosen_name,
	family_name = @family_name,
	email = @email,
	username = @username,
	grade = @grade,
	homeroom = @homeroom,
	case_manager = @case_manager,
	last_event_commit_position = @last_event_commit_position,
	last_event_prepare_position = @last_event_prepare_position,
	updated_at = @updated_at
WHERE id = @id;

-- name: ArchiveStudent :exec
UPDATE students
SET
	last_event_commit_position = @last_event_commit_position,
	last_event_prepare_position = @last_event_prepare_position,
	updated_at = @archived_at,
	archived_at = @archived_at
WHERE id = @id;

-- name: DeleteStudent :exec
DELETE FROM students
WHERE id = @id;

-- name: ListStudentsByIEPServiceType :many
SELECT 
    s.id, 
		s.marss_id,
    s.given_name, 
    s.chosen_name, 
    s.family_name, 
		s.email,
		s.username,
    s.grade, 
    s.homeroom, 
    s.case_manager, 
    s.created_at, 
    s.updated_at,
    s.archived_at
FROM students s
WHERE EXISTS (
    SELECT 1 FROM iep_services ieps
    WHERE ieps.student_id = s.id
      AND ieps.service_type = sqlc.arg('service_type')
      AND ieps.archived_at IS NULL
);