-- name: GetPeriodEducator :one
SELECT period_id, educator_id
FROM educators_periods
WHERE period_id = @period_id AND educator_id = @educator_id;

-- name: ListPeriodsEducators :many
SELECT period_id, educator_id
FROM educators_periods;

-- name: ListEducatorIDsForPeriod :many
SELECT period_id, educator_id
FROM educators_periods
WHERE period_id = @period_id
ORDER BY educator_id DESC;

-- name: ListPeriodIDsForEducator :many
SELECT period_id, educator_id
FROM educators_periods
WHERE educator_id = @educator_id
ORDER BY period_id DESC;

-- name: AddEducatorToPeriod :exec
INSERT INTO educators_periods (
	period_id, 
	educator_id, 
	last_event_commit_position, 
	last_event_prepare_position,
	created_at, 
	updated_at
)
VALUES (
	@period_id, 
	@educator_id, 
	@last_event_commit_position, 
	@last_event_prepare_position,
	@created_at, 
	@created_at
)
ON CONFLICT (period_id, educator_id) DO NOTHING;

-- name: RemoveEducatorFromPeriod :exec
DELETE FROM educators_periods
WHERE period_id = @period_id 
	AND educator_id = @educator_id;