-- name: GetUserProfileByUserID :one
SELECT 
	user_id,
	coalesce(email, '') AS email,
	coalesce(username, '') AS username,
	coalesce(image, '') AS image,
	coalesce(bio, '') AS bio,
	coalesce(header_image_url, '') AS header_image_url
FROM profile_stats
WHERE user_id = @user_id;

-- name: GetUserProfileByUserUsername :one
SELECT 
	user_id,
	coalesce(email, '') AS email,
	coalesce(username, '') AS username,
	coalesce(image, '') AS image,
	coalesce(bio, '') AS bio,
	coalesce(header_image_url, '') AS header_image_url
FROM profile_stats
WHERE username = @username;

-- name: UpsertRegisteredProfileUser :exec
INSERT INTO profile_stats (
	user_id, 
	email, 
	username,
	last_event_commit_position, 
	last_event_prepare_position
)
VALUES (
	@user_id, 
	@email, 
	@username,
	@last_event_commit_position, 
	@last_event_prepare_position
)
ON CONFLICT (user_id) DO UPDATE SET
    email = EXCLUDED.email,
		username = EXCLUDED.username,
    last_event_commit_position = EXCLUDED.last_event_commit_position,
    last_event_prepare_position = EXCLUDED.last_event_prepare_position,
    updated_at = CURRENT_TIMESTAMP;

-- name: UpsertProfileBio :exec
INSERT INTO profile_stats (
	user_id, 
	bio, 
	last_event_commit_position, 
	last_event_prepare_position
)
VALUES (
	@user_id, 
	@bio, 
	@last_event_commit_position, 
	@last_event_prepare_position
)
ON CONFLICT (user_id) DO UPDATE SET
    bio = EXCLUDED.bio,
    last_event_commit_position = EXCLUDED.last_event_commit_position,
    last_event_prepare_position = EXCLUDED.last_event_prepare_position,
    updated_at = CURRENT_TIMESTAMP;

-- name: UpsertProfileImage :exec
INSERT INTO profile_stats (
	user_id, 
	image, 
	last_event_commit_position, 
	last_event_prepare_position
)
VALUES (
	@user_id, 
	@image, 
	@last_event_commit_position, 
	@last_event_prepare_position
)
ON CONFLICT (user_id) DO UPDATE SET
    image = EXCLUDED.image,
    last_event_commit_position = EXCLUDED.last_event_commit_position,
    last_event_prepare_position = EXCLUDED.last_event_prepare_position,
    updated_at = CURRENT_TIMESTAMP;

-- name: UpsertProfileHeaderImage :exec
INSERT INTO profile_stats (
	user_id, 
	header_image_url, 
	last_event_commit_position, 
	last_event_prepare_position
)
VALUES (
	@user_id, 
	@header_image_url, 
	@last_event_commit_position, 
	@last_event_prepare_position
)
ON CONFLICT (user_id) DO UPDATE SET
    header_image_url = EXCLUDED.header_image_url,
    last_event_commit_position = EXCLUDED.last_event_commit_position,
    last_event_prepare_position = EXCLUDED.last_event_prepare_position,
    updated_at = CURRENT_TIMESTAMP;

-- name: DeleteProfileByRegisteredID :exec
DELETE FROM profile_stats
WHERE user_id = @user_registered_id;
