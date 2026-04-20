package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const simcNoArgumentsExitCode = 50
const simcProfileFileMode = 0o600

// RunResult holds the artifacts produced by a single simc invocation.
type RunResult struct {
	// Stdout is the human‑readable log simc writes to standard output.
	Stdout []byte

	// JSON2 is the raw contents of the file produced by passing `json2=<path>`
	// to simc — its structured report. Parse with simc.ParseJSON2.
	JSON2 []byte
}

type Runner interface {
	Run(ctx context.Context, profilePath string) (RunResult, error)
}

type simcRunner struct {
	// simc binary path
	binaryPath string
}

func (runner simcRunner) Version(ctx context.Context) (string, error) {
	// #nosec G204 -- the binary path comes from deployment configuration, not queue input.
	command := exec.CommandContext(ctx, runner.binaryPath)

	var output bytes.Buffer
	command.Stdout = &output

	err := command.Run()
	exitCode := -1
	if command.ProcessState != nil {
		exitCode = command.ProcessState.ExitCode()
	}

	if err != nil && exitCode != simcNoArgumentsExitCode {
		return "", fmt.Errorf("run simc binary: %w", err)
	}

	return strings.TrimSpace(output.String()), nil
}

func (runner simcRunner) Run(ctx context.Context, profilePath string) (RunResult, error) {
	// Write the structured report next to the profile so the caller's temp‑dir
	// cleanup (os.RemoveAll) sweeps it up automatically.
	jsonPath := filepath.Join(filepath.Dir(profilePath), "output.json")

	// #nosec G204 -- the binary path comes from deployment configuration and the
	// profile/json paths are created locally by the worker.
	command := exec.CommandContext(
		ctx,
		runner.binaryPath,
		profilePath,
		"json2="+jsonPath,
	)

	var stdout bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = os.Stderr

	if err := command.Run(); err != nil {
		return RunResult{}, fmt.Errorf("execute simc binary: %w", err)
	}

	jsonBytes, err := os.ReadFile(jsonPath) // #nosec G304 -- path constructed above
	if err != nil {
		return RunResult{}, fmt.Errorf("read simc json2 output %q: %w", jsonPath, err)
	}

	return RunResult{
		Stdout: stdout.Bytes(),
		JSON2:  jsonBytes,
	}, nil
}
