package executor

// ExecutionPayload represents the payload structure for executing a job
type ExecutionPayload struct {
	Type    string   `json:"type"`    // e.g., "SHELL"
	Command string   `json:"command"` // e.g., "/usr/local/bin/script.sh"
	Args    []string `json:"args"`    // Command arguments
	Timeout string   `json:"timeout"` // e.g., "5m", "30s"
}

// ExecutionResult contains the output and status of a job execution
type ExecutionResult struct {
	Stdout   string
	Stderr   string
	ExitCode int32
	Error    error
}
