-- name: AddStudentBookmark :exec
INSERT INTO user_student_bookmarks (
	user_id, 
	student_id, 
	last_event_commit_position, 
	last_event_prepare_position,
	added_at
)
VALUES (
	@user_id, 
	@student_id, 
	@last_event_commit_position, 
	@last_event_prepare_position,
	@added_at
)
ON CONFLICT (user_id, student_id) DO NOTHING;

-- name: RemoveStudentBookmark :exec
DELETE FROM user_student_bookmarks
WHERE user_id = @user_id 
	AND student_id = @student_id;