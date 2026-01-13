CREATE TYPE job_status AS ENUM ('pending', 'in_progress', 'completed', 'failed', 'dead');

CREATE TABLE IF NOT EXISTS jobs (
	id SERIAL PRIMARY KEY,
	parent_job_id INT REFERENCES jobs(id) ON DELETE SET NULL,
	title VARCHAR(255) NOT NULL,
	description TEXT,
	payload JSONB NOT NULL DEFAULT '{}',
	max_retries INT DEFAULT 3,
	retries INT DEFAULT 0,
	status job_status DEFAULT 'pending',
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS job_logs (
	id SERIAL PRIMARY KEY,
	job_id INT NOT NULL REFERENCES jobs(id) ON DELETE CASCADE,
	stdout TEXT,
	stderr TEXT,
	exit_code INT,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_jobs_parent_job_id ON jobs(parent_job_id);
CREATE INDEX idx_jobs_status ON jobs(status);
CREATE INDEX idx_job_logs_job_id ON job_logs(job_id);
