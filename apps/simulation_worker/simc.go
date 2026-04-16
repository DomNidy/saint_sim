package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	api_types "github.com/DomNidy/saint_sim/internal/api_types"
	utils "github.com/DomNidy/saint_sim/internal/utils"
)

const simcNoArgumentsExitCode = 50
const simcProfileFileMode = 0o600

type simcRunner struct {
	binaryPath string
}

type simulationInput struct {
	simcAddonExport string
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

func (runner simcRunner) Run(ctx context.Context, input simulationInput) ([]byte, error) {
	tempDir, err := os.MkdirTemp("", "saint-simc-*")
	if err != nil {
		return nil, fmt.Errorf("create simc temp dir: %w", err)
	}

	defer func() {
		removeErr := os.RemoveAll(tempDir)
		if removeErr != nil {
			fmt.Fprintf(os.Stderr, "remove simc temp dir: %v\n", removeErr)
		}
	}()

	profilePath := filepath.Join(tempDir, "input.simc")

	err = os.WriteFile(profilePath, []byte(input.simcAddonExport), simcProfileFileMode)
	if err != nil {
		return nil, fmt.Errorf("write simc profile: %w", err)
	}

	// #nosec G204 -- the binary path comes from deployment configuration and the
	// profile path is created locally.
	command := exec.CommandContext(ctx, runner.binaryPath, profilePath)

	var output bytes.Buffer

	command.Stdout = &output
	command.Stderr = os.Stderr

	err = command.Run()
	if err != nil {
		return nil, fmt.Errorf("execute simc binary: %w", err)
	}

	return output.Bytes(), nil
}

func simulationInputFromOptions(options api_types.SimulationOptions) (simulationInput, error) {
	err := utils.ValidateSimOptions(&options)
	if err != nil {
		return simulationInput{}, fmt.Errorf("validate simulation options: %w", err)
	}

	return simulationInput{
		simcAddonExport: options.SimcAddonExport,
	}, nil
}
