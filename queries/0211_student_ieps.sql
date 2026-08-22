-- name: GetIEPByStudentID :one
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
	s.updated_at,
	si.id AS iep_id,
	si.case_manager_id, 
	si.start_date,
	si.end_date,
	si.amended_date,
	si.created_at AS iep_created_at,
	si.updated_at AS iep_updated_at
FROM student_ieps si
INNER JOIN students s ON si.student_id = s.id
WHERE s.archived_at IS NULL
	AND si.archived_at IS NULL
	AND si.student_id = @id;

-- name: ListStudentsWithIEPs :many
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
	s.updated_at,
	si.id AS iep_id,
	si.case_manager_id,
	si.start_date,
	si.end_date,
	si.amended_date,
	si.created_at AS iep_created_at,
	si.updated_at AS iep_updated_at
FROM student_ieps si
INNER JOIN students s ON si.student_id = s.id
WHERE s.archived_at IS NULL
  AND si.archived_at IS NULL
ORDER BY s.family_name, s.given_name;

-- name: AddIEPToStudent :exec
INSERT INTO student_ieps (
	id, 
	student_id,
	case_manager_id, 
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
	@case_manager_id, 
	@start_date,
	@end_date,
	@amended_date,
	@last_event_commit_position, 
	@last_event_prepare_position,
	@created_at, 
	@created_at
)
ON CONFLICT (id) DO NOTHING;

-- name: UpdateIEPCaseManager :exec
UPDATE student_ieps
SET 
	case_manager_id = @case_manager_id,
	last_event_commit_position = @last_event_commit_position,
	last_event_prepare_position = @last_event_prepare_position,
	updated_at = @updated_at
WHERE id = @id;

-- name: UpdateIEPDates :exec
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