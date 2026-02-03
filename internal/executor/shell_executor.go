package executor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"syscall"
	"time"
)

// ShellExecutor executes shell commands
type ShellExecutor struct{}

// Execute runs a shell command and captures stdout, stderr, and exit code
func (se *ShellExecutor) Execute(ctx context.Context, payload json.RawMessage) (*ExecutionResult, error) {
	var execPayload ExecutionPayload
	if err := json.Unmarshal(payload, &execPayload); err != nil {
		return &ExecutionResult{
			ExitCode: 1,
			Error:    fmt.Errorf("invalid payload: %v", err),
		}, err
	}

	// Parse timeout if provided, default to context timeout
	timeout := 30 * time.Second
	if execPayload.Timeout != "" {
		parsedTimeout, err := time.ParseDuration(execPayload.Timeout)
		if err != nil {
			return &ExecutionResult{
				ExitCode: 1,
				Error:    fmt.Errorf("invalid timeout format: %v", err),
			}, err
		}
		timeout = parsedTimeout
	}

	// Create a context with timeout
	execCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Create the command
	cmd := exec.CommandContext(execCtx, execPayload.Command, execPayload.Args...)

	// Capture stdout and stderr
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Run the command
	err := cmd.Run()

	result := &ExecutionResult{
		Stdout: stdout.String(),
		Stderr: stderr.String(),
	}

	// Extract exit code
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
				result.ExitCode = int32(status.ExitStatus())
			} else {
				result.ExitCode = 1
			}
		} else if execCtx.Err() == context.DeadlineExceeded {
			result.ExitCode = 124 // Standard timeout exit code
			result.Error = fmt.Errorf("command timed out after %v", timeout)
			return result, result.Error
		} else {
			result.ExitCode = 1
			result.Error = err
		}
	} else {
		result.ExitCode = 0
	}

	return result, nil
}
