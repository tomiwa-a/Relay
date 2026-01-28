ALTER TABLE job_logs
DROP COLUMN level,
DROP COLUMN message;

DROP TYPE log_level;
