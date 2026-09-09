-- name: GetIEP :one
SELECT
	id,
	student_id,
	plan_manager_id,
	disability_1,
	disability_2,
	federal_setting,
	meeting_date,
	iep_due_date,
	last_eval_date,
	eval_due_date,
	amended_date,
	iep_type,
	special_transportation,
	created_at,
	updated_at,
	archived_at
FROM student_ieps
WHERE id = @id;

-- name: ListIEPs :many
SELECT
	id,
	student_id,
	plan_manager_id,
	disability_1,
	disability_2,
	federal_setting,
	meeting_date,
	iep_due_date,
	last_eval_date,
	eval_due_date,
	amended_date,
	iep_type,
	special_transportation,
	created_at,
	updated_at,
	archived_at
FROM student_ieps;

-- name: ListIEPsForStudent :many
SELECT
	id,
	student_id,
	plan_manager_id,
	disability_1,
	disability_2,
	federal_setting,
	meeting_date,
	iep_due_date,
	last_eval_date,
	eval_due_date,
	amended_date,
	iep_type,
	special_transportation,
	created_at,
	updated_at,
	archived_at
FROM student_ieps
WHERE student_id = @student_id;

-- name: AddIEPToStudent :exec
INSERT INTO student_ieps (
	id,
	student_id,
	plan_manager_id,
	disability_1,
	disability_2,
	federal_setting,
	meeting_date,
	iep_due_date,
	last_eval_date,
	eval_due_date,
	amended_date,
	iep_type,
	special_transportation,
	last_event_commit_position,
	last_event_prepare_position,
	created_at,
	updated_at
)
VALUES (
	@id,
	@student_id,
	@plan_manager_id,
	@disability_1,
	@disability_2,
	@federal_setting,
	@meeting_date,
	@iep_due_date,
	@last_eval_date,
	@eval_due_date,
	@amended_date,
	@iep_type,
	@special_transportation,
	@last_event_commit_position,
	@last_event_prepare_position,
	@created_at,
	@created_at
)
ON CONFLICT (id) DO NOTHING;

-- name: UpdateIEP :exec
UPDATE student_ieps
SET
	plan_manager_id = @plan_manager_id,
	disability_1 = @disability_1,
	disability_2 = @disability_2,
	federal_setting = @federal_setting,
	meeting_date = @meeting_date,
	iep_due_date = @iep_due_date,
	last_eval_date = @last_eval_date,
	eval_due_date = @eval_due_date,
	amended_date = @amended_date,
	iep_type = @iep_type,
	special_transportation = @special_transportation,
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