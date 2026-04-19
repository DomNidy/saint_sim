package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const simcNoArgumentsExitCode = 50
const simcProfileFileMode = 0o600

type Runner interface {
	Run(ctx context.Context, profilePath string) ([]byte, error)
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

func (runner simcRunner) Run(ctx context.Context, profilePath string) ([]byte, error) {
	// #nosec G204 -- the binary path comes from deployment configuration and the
	// profile path is created locally.
	command := exec.CommandContext(ctx, runner.binaryPath, profilePath)

	var output bytes.Buffer

	command.Stdout = &output
	command.Stderr = os.Stderr

	err := command.Run()
	if err != nil {
		return nil, fmt.Errorf("execute simc binary: %w", err)
	}

	return output.Bytes(), nil
}
