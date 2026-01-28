-- name: CreateJob :one
INSERT INTO jobs (
    parent_job_id,
    title,
    description,
    payload,
    max_retries
) VALUES (
    $1, $2, $3, $4, $5
) RETURNING *;

-- name: ListJobs :many
SELECT * FROM jobs
ORDER BY created_at DESC;

-- name: GetPendingJobs :many
SELECT * FROM jobs
WHERE status = 'pending'
ORDER BY created_at ASC;

-- name: UpdateJobStatus :one
UPDATE jobs
SET 
    status = $2,
    retries = $3,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: CreateJobLog :one
INSERT INTO job_logs (
    job_id,
    stdout,
    stderr,
    exit_code
) VALUES (
    $1, $2, $3, $4
) RETURNING *;

-- name: GetJobLogs :many
SELECT * FROM job_logs
WHERE job_id = $1
ORDER BY created_at ASC;