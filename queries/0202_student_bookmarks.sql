-- name: GetStudentBookmark :one
SELECT
	s.id,
	s.marss_id,
	s.birthdate,
	s.given_name,
	s.chosen_name,
	s.family_name,
	s.pronouns,
	s.email,
	s.username,
	s.grade,
	s.homeroom_id, 
	s.plan_type, 
	s.created_at,
	s.updated_at
FROM students s
INNER JOIN user_student_bookmarks b ON s.id = b.student_id
WHERE s.archived_at IS NULL
	AND b.user_id = @user_id
	AND s.id = @student_id;

-- name: ListStudentBookmarksByUserID :many
SELECT
	s.id,
	s.marss_id,
	s.birthdate,
	s.given_name,
	s.chosen_name,
	s.family_name,
	s.pronouns,
	s.email,
	s.username,
	s.grade,
	s.homeroom_id, 
	s.plan_type, 
	s.created_at,
	s.updated_at
FROM students s
INNER JOIN user_student_bookmarks b ON s.id = b.student_id
WHERE s.archived_at IS NULL
	AND b.user_id = @user_id
ORDER BY s.family_name COLLATE NOCASE ASC, s.chosen_name COLLATE NOCASE ASC, s.given_name COLLATE NOCASE ASC;