-- name: GetPeriodSchedule :one
SELECT period_id, schedule_id
FROM periods_schedules
WHERE period_id = @period_id AND schedule_id = @schedule_id;

-- name: ListPeriodsSchedules :many
SELECT period_id, schedule_id
FROM periods_schedules;

-- name: ListScheduleIDsForPeriod :many
SELECT period_id, schedule_id
FROM periods_schedules
WHERE period_id = @period_id
ORDER BY schedule_id DESC;

-- name: ListPeriodIDsForSchedule :many
SELECT period_id, schedule_id
FROM periods_schedules
WHERE schedule_id = @schedule_id
ORDER BY period_id DESC;

-- name: AddPeriodToSchedule :exec
INSERT INTO periods_schedules (period_id, schedule_id, created_at, updated_at, last_event_commit_position, last_event_prepare_position)
VALUES (@period_id, @schedule_id, @created_at, @created_at, @last_event_commit_position, @last_event_prepare_position)
ON CONFLICT (period_id, schedule_id) DO NOTHING;

-- name: RemovePeriodFromSchedule :exec
DELETE FROM periods_schedules
WHERE period_id = @period_id AND schedule_id = @schedule_id;