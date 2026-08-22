-- name: GetUserProfileByUserID :one
SELECT 
	user_id,
	coalesce(avatar, '') AS avatar,
	coalesce(bio, '') AS bio
FROM profile_stats
WHERE user_id = @user_id;

-- name: CreateProfileForUser :exec
INSERT INTO profile_stats (
	user_id, 
	last_event_commit_position, 
	last_event_prepare_position
)
VALUES (
	@user_id, 
	@last_event_commit_position, 
	@last_event_prepare_position
)
ON CONFLICT (user_id) DO NOTHING;

-- name: ProfileUpdateAvatar :exec
UPDATE profile_stats 
 SET 
	avatar = @avatar, 
	last_event_commit_position = @last_event_commit_position,
	last_event_prepare_position = @last_event_prepare_position,
	updated_at = @updated_at
WHERE user_id = @user_id;

-- name: ProfileUpdateBio :exec
UPDATE profile_stats
	SET
		bio = @bio, 
		last_event_commit_position = @last_event_commit_position,
		last_event_prepare_position = @last_event_prepare_position,
		updated_at = @updated_at
WHERE user_id = @user_id;

-- name: DeleteProfileByRegisteredID :exec
DELETE FROM profile_stats
WHERE user_id = @user_registered_id;
