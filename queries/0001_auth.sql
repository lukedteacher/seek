-- name: CreateAuthUser :exec
INSERT OR IGNORE INTO auth_user (
	id, 
	email, 
	username,
	user_registered_id
)
VALUES (
	@id, 
	@email, 
	@username, 
	@user_registered_id
);

-- name: CreateAuthAccount :exec
INSERT OR IGNORE INTO auth_account (
	id, 
	account_id, 
	provider_id, 
	user_id, 
	password
)
VALUES (
	@id, 
	@account_id, 
	'credential', 
	@user_id, 
	@password
);

-- name: CreateAuthSession :exec
INSERT INTO auth_session (id, token, user_id, expires_at)
VALUES (@id, @token, @user_id, @expires_at);

-- name: DeleteAuthSessionByToken :exec
DELETE FROM auth_session
WHERE token = @token;

-- name: UserBySessionToken :one
SELECT
	u.id,
	u.user_registered_id,
	u.email,
	u.username,
	coalesce(u.avatar, '') AS avatar,
	coalesce(p.bio, '') AS bio
FROM auth_session s
JOIN auth_user u ON u.id = s.user_id
LEFT JOIN profile_stats p ON p.user_id = u.user_registered_id
WHERE s.token = @token
  AND s.expires_at > CURRENT_TIMESTAMP;

-- name: UserByRegisteredID :one
SELECT 
	u.id,
	u.user_registered_id,
	u.email,
	u.username,
	coalesce(u.avatar, '') AS avatar,
	coalesce(p.bio, '') AS bio
FROM auth_user u
LEFT JOIN profile_stats p ON p.user_id = u.user_registered_id
WHERE u.user_registered_id = @user_registered_id;

-- name: UserByIDOrRegisteredID :one
SELECT 
	u.id,
	u.user_registered_id,
	u.email,
	u.username,
	coalesce(u.avatar, '') AS avatar,
	coalesce(p.bio, '') AS bio
FROM auth_user u
LEFT JOIN profile_stats p ON p.user_id = u.user_registered_id
WHERE u.id = @id OR u.user_registered_id = @id;

-- name: UpdateAuthUserAvatar :exec
UPDATE auth_user
SET avatar = @avatar,
    updated_at = CURRENT_TIMESTAMP
WHERE user_registered_id = @user_registered_id;

-- name: UpdateAuthAccountPassword :exec
UPDATE auth_account
SET password = @password,
    updated_at = CURRENT_TIMESTAMP
WHERE user_id = @user_id
  AND provider_id = 'credential';

-- name: UpdateAuthAccountPasswordByRegisteredID :exec
UPDATE auth_account
SET password = @password,
    updated_at = CURRENT_TIMESTAMP
WHERE user_id = (
    SELECT id
    FROM auth_user
    WHERE user_registered_id = @user_registered_id
)
  AND provider_id = 'credential';

-- name: GetUserByEmailWithPassword :one
SELECT 
	u.id,
	u.user_registered_id,
	u.email,
	u.username,
	coalesce(u.avatar, '') AS avatar,
	coalesce(p.bio, '') AS bio,
	a.password
FROM auth_user u
JOIN auth_account a ON a.user_id = u.id AND a.provider_id = 'credential'
LEFT JOIN profile_stats p ON p.user_id = u.user_registered_id
WHERE u.email = @email;

-- name: GetUserByUsernameWithPassword :one
SELECT 
	u.id,
	u.user_registered_id,
	u.email,
	u.username,
	coalesce(u.avatar, '') AS avatar,
	coalesce(p.bio, '') AS bio,
	a.password
FROM auth_user u
JOIN auth_account a ON a.user_id = u.id AND a.provider_id = 'credential'
LEFT JOIN profile_stats p ON p.user_id = u.user_registered_id
WHERE u.username = @username;

-- name: DeleteAuthSessionsByRegisteredID :exec
DELETE FROM auth_session
WHERE user_id IN (
    SELECT id
    FROM auth_user
    WHERE user_registered_id = @user_registered_id
);

-- name: DeleteAuthAccountsByRegisteredID :exec
DELETE FROM auth_account
WHERE user_id IN (
    SELECT id
    FROM auth_user
    WHERE user_registered_id = @user_registered_id
);

-- name: DeleteAuthUserByRegisteredID :exec
DELETE FROM auth_user
WHERE user_registered_id = @user_registered_id;
