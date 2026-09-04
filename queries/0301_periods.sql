-- name: GetPeriod :one
SELECT
	id, 
	title, 
	service_type,
	start_time, 
	duration, 
	days_bitmask, 
	created_at, 
	updated_at
FROM periods
WHERE archived_at IS NULL
	AND id = @id;

-- name: GetPeriodWithIDs :one
SELECT
    p.id,
    p.title,
    p.service_type,
    p.start_time,
    p.duration,
    p.days_bitmask,
    p.created_at,
    p.updated_at,
    CAST(
        COALESCE(
            (SELECT json_group_array(ps.student_id) FROM periods_students ps WHERE ps.period_id = p.id),
            '[]'
        ) AS TEXT
    ) AS student_ids,
    CAST(
        COALESCE(
            (SELECT json_group_array(ep.educator_id) FROM educators_periods ep WHERE ep.period_id = p.id),
            '[]'
        ) AS TEXT
    ) AS educator_ids
FROM periods p
WHERE p.id = @id;

-- name: ListPeriods :many
SELECT
	id, 
	title, 
	service_type,
	start_time, 
	duration, 
	days_bitmask, 
	created_at, 
	updated_at
FROM periods
WHERE archived_at IS NULL
ORDER BY service_type DESC, title DESC;

-- name: ListPeriodsWithIDs :many
SELECT
    p.id,
    p.title,
    p.service_type,
    p.start_time,
    p.duration,
    p.days_bitmask,
    p.created_at,
    p.updated_at,
    CAST(
        COALESCE(
            (SELECT json_group_array(ps.student_id) FROM periods_students ps WHERE ps.period_id = p.id),
            '[]'
        ) AS TEXT
    ) AS student_ids,
    CAST(
        COALESCE(
            (SELECT json_group_array(ep.educator_id) FROM educators_periods ep WHERE ep.period_id = p.id),
            '[]'
        ) AS TEXT
    ) AS educator_ids
FROM periods p;

-- name: ListPeriodsForEducator :many
SELECT
	periods.id, 
	periods.title, 
	periods.service_type,
	periods.start_time, 
	periods.duration, 
	periods.days_bitmask, 
	periods.created_at, 
	periods.updated_at
FROM periods
JOIN educators_periods ON periods.id = educators_periods.period_id
WHERE educators_periods.educator_id = @educator_id
ORDER BY start_time DESC;

-- name: ListPeriodsForStudent :many
SELECT
	periods.id, 
	periods.title, 
	periods.service_type,
	periods.start_time, 
	periods.duration, 
	periods.days_bitmask, 
	periods.created_at, 
	periods.updated_at
FROM periods
JOIN periods_students ON periods.id = periods_students.period_id
WHERE periods_students.student_id = @student_id
ORDER BY start_time DESC;

-- name: CreatePeriod :exec
INSERT INTO periods (
	id, 
	title, 
	service_type,
	start_time, 
	duration, 
	days_bitmask, 
	last_event_commit_position, 
	last_event_prepare_position,
	created_at, 
	updated_at
)
VALUES (
	@id, 
	@title, 
	@service_type,
	@start_time, 
	@duration, 
	@days_bitmask, 
	@last_event_commit_position, 
	@last_event_prepare_position,
	@created_at, 
	@created_at
)
ON CONFLICT (id) DO NOTHING;

-- name: UpdatePeriod :exec
UPDATE periods
SET
	title = @title,
	service_type = @service_type,
	start_time = @start_time,
	duration = @duration,
	days_bitmask = @days_bitmask,
	last_event_commit_position = @last_event_commit_position,
	last_event_prepare_position = @last_event_prepare_position,
	updated_at = @updated_at
WHERE id = @id;

-- name: ArchivePeriod :exec
UPDATE periods
SET
	last_event_commit_position = @last_event_commit_position,
	last_event_prepare_position = @last_event_prepare_position,
	updated_at = @archived_at,
	archived_at = @archived_at
WHERE id = @id;

-- name: DeletePeriod :exec
DELETE FROM periods
WHERE id = @id;