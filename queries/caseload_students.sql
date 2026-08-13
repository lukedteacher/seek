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

-- name: RemoveStudentFromCaseload :exec
DELETE FROM caseload_students
WHERE educator_id = @educator_id
	AND student_id = @student_id;