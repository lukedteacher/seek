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
	homeroom_id, 
	plan_type, 
	created_at, 
	updated_at
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
	homeroom_id, 
	plan_type, 
	created_at, 
	updated_at
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
	homeroom_id, 
	plan_type, 
	created_at, 
	updated_at
FROM students
WHERE archived_at IS NULL
ORDER BY family_name DESC, given_name DESC;

-- name: ListStudentsByGrade :many
SELECT 
	id, 
	marss_id,
	given_name, 
	chosen_name, 
	family_name, 
	email,
	username,
	grade, 
	homeroom_id, 
	plan_type, 
	created_at, 
	updated_at
FROM students
WHERE archived_at IS NULL
  AND grade IN (sqlc.slice('grades'))
ORDER BY family_name DESC, given_name DESC;

-- name: ListStudentsByServiceType :many
SELECT 
	s.id, 
	s.marss_id,
	s.given_name, 
	s.chosen_name, 
	s.family_name, 
	s.email,
	s.username,
	s.grade, 
	s.homeroom_id, 
	s.plan_type, 
	s.created_at, 
	s.updated_at
FROM students s
WHERE EXISTS (
	SELECT 1 FROM iep_services ieps
	WHERE ieps.student_id = s.id
		AND ieps.service_type = sqlc.arg('service_type')
		AND ieps.archived_at IS NULL
		AND s.archived_at IS NULL
);

-- name: ListOnlyStudentsWithIEPs :many
SELECT
	s.id AS student_id,
	s.marss_id,
	s.given_name,
	s.chosen_name,
	s.family_name,
	s.email,
	s.username,
	s.grade,
	s.homeroom_id, 
	s.plan_type, 
	s.created_at,
	s.updated_at,
	i.id AS iep_id,
	i.start_date,
	i.end_date,
	i.amended_date,
	i.created_at AS iep_created_at,
	i.updated_at AS iep_updated_at
FROM students s
INNER JOIN student_ieps i ON s.id = i.student_id
WHERE s.archived_at IS NULL
  AND i.archived_at IS NULL
ORDER BY s.family_name ASC, s.chosen_name ASC, s.given_name ASC;

-- name: ListStudentsWithServices :many
SELECT
	s.id AS student_id,
	s.marss_id,
	s.given_name,
	s.chosen_name,
	s.family_name,
	s.email,
	s.username,
	s.grade,
	s.homeroom_id, 
	s.plan_type, 
	s.created_at,
	s.updated_at,
	i.id AS service_id,
	i.service_name,
	i.service_type,
	i.indirect_minutes,
	i.direct_minutes,
	i.frequency_count,
	i.frequency_type,
	i.location_id,
	i.start_date,
	i.end_date,
	i.provider_id,
	i.created_at AS service_created_at,
	i.updated_at AS service_updated_at
FROM students s
LEFT JOIN iep_services i ON s.id = i.student_id
WHERE s.archived_at IS NULL
  AND i.archived_at IS NULL
ORDER BY s.family_name DESC, s.chosen_name DESC, s.given_name DESC, i.service_type ASC;

-- name: ListStudentsForPeriod :many
SELECT
	s.id, 
	s.marss_id,
	s.given_name, 
	s.chosen_name, 
	s.family_name, 
	s.email,
	s.username,
	s.grade, 
	s.homeroom_id, 
	s.plan_type, 
	s.created_at,
	s.updated_at
FROM students s
JOIN periods_students ps ON s.id = ps.student_id
WHERE ps.period_id = @period_id
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
	homeroom_id, 
	plan_type, 
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
	@homeroom_id, 
	@plan_type, 
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
	homeroom_id = @homeroom_id,
	plan_type = @plan_type,
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