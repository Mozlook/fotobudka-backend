CREATE INDEX idx_jobs_pending_due
ON jobs (next_run_at, id)
WHERE status = 'pending';

CREATE INDEX idx_jobs_running_locked_at
ON jobs (locked_at)
WHERE status = 'running';
