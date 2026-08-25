-- name: ResetReadModelAuthSessions :exec
DELETE FROM auth_session;

-- name: ResetReadModelAuthAccounts :exec
DELETE FROM auth_account;

-- name: ResetReadModelAuthUsers :exec
DELETE FROM auth_user;

-- name: ResetReadModelProfiles :exec
DELETE FROM profile_stats;

-- name: ResetReadModelServices :exec
DELETE FROM iep_services;

-- name: ResetReadModelPeriods :exec
DELETE FROM periods;

-- name: ResetReadModelStudents :exec
DELETE FROM students;

-- name: ResetReadModelPeriodsStudents :exec
DELETE FROM periods_students;

-- name: ResetEventHandlerCheckpoints :exec
DELETE FROM projector_checkpoint;
