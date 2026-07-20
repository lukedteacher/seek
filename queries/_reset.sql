-- name: ResetReadModelAuthSessions :exec
DELETE FROM auth_session;

-- name: ResetReadModelAuthAccounts :exec
DELETE FROM auth_account;

-- name: ResetReadModelAuthUsers :exec
DELETE FROM auth_user;

-- name: ResetReadModelAuthVerifications :exec
DELETE FROM auth_verification;

-- name: ResetReadModelProfiles :exec
DELETE FROM profile_stats;

-- name: ResetReadModelIEPServices :exec
DELETE FROM iep_services;

-- name: ResetReadModelPeriods :exec
DELETE FROM periods;

-- name: ResetReadModelSchedules :exec
DELETE FROM schedules;

-- name: ResetReadModelStudents :exec
DELETE FROM students;

-- name: ResetReadModelPeriodsSchedules :exec
DELETE FROM periods_schedules;

-- name: ResetReadModelPeriodsStudents :exec
DELETE FROM periods_students;

-- name: ResetReadModelTeachers :exec
DELETE FROM teachers;

-- name: ResetEventHandlerCheckpoints :exec
DELETE FROM projector_checkpoint;
