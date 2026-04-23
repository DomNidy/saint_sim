package sims

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
	// to simc — its structured report. Parse with ParseJSON2.
	JSON2 []byte
}

// TODO: Bad encapsulation - we should just have a Run()
// method on the runner that takes the top gear manifest
// directly. To support a basic sim, Runner can either use
// methods per-sim-kind (RunBasic, RunTopGear), or use generics.
type Runner interface {
	Run(ctx context.Context, profileText string) (RunResult, error)
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

// Run writes the profileText to a temporary profile on disk, performs the sim, returns the
// results and cleans up the temp profile from disk.
func (runner simcRunner) Run(ctx context.Context, profileText string) (RunResult, error) {
	profilePath, cleanupFunc, err := runner.writeSimcProfileTemp(ctx, profileText)
	if err == nil {
		return RunResult{}, err
	}
	defer cleanupFunc()

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

// write a simc profile to disk and return a cleanup function along with the
// path to the profile
// must call the cleanup function.
func (runner simcRunner) writeSimcProfileTemp(
	ctx context.Context,
	profileText string,
) (string, func(), error) {
	tempDir, err := os.MkdirTemp("", "saint-simc-*")
	if err != nil {
		return "", func() {}, fmt.Errorf("create simc temp dir: %w", err)
	}

	cleanupFunc := func() {
		if removeErr := os.RemoveAll(tempDir); removeErr != nil {
			fmt.Fprintf(os.Stderr, "remove simc temp dir: %v\n", removeErr)
		}
	}

	profilePath := filepath.Join(tempDir, "input.simc")
	if err := os.WriteFile(profilePath, []byte(profileText), simcProfileFileMode); err != nil {
		return "", func() {}, fmt.Errorf("write simc profile: %w", err)
	}

	return profilePath, cleanupFunc, nil
}
