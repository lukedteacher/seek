-- name: AddStudentToCaseload :exec
INSERT INTO caseload_students (
	educator_id, 
	student_id, 
	last_event_commit_position, 
	last_event_prepare_position,
	created_at, 
	updated_at
)
VALUES (
	@educator_id, 
	@student_id, 
	@last_event_commit_position, 
	@last_event_prepare_position,
	@created_at, 
	@created_at
)
ON CONFLICT (educator_id, student_id) DO NOTHING;

-- name: SetStudentCaseManager :exec
UPDATE students
SET
	case_manager_id = @case_manager_id,
	last_event_commit_position = @last_event_commit_position,
	last_event_prepare_position = @last_event_prepare_position,
	updated_at = @updated_at
WHERE id = @id;

-- name: RemoveStudentFromCaseload :exec
DELETE FROM caseload_students
WHERE educator_id = @educator_id
	AND student_id = @student_id;

-- name: ClearStudentCaseManager :exec
UPDATE students
SET
	case_manager_id = '',
	last_event_commit_position = @last_event_commit_position,
	last_event_prepare_position = @last_event_prepare_position,
	updated_at = @updated_at
WHERE id = @id;