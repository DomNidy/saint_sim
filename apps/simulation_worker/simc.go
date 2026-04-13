package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	api_types "github.com/DomNidy/saint_sim/pkg/api_types"
	utils "github.com/DomNidy/saint_sim/pkg/utils"
)

const simcNoArgumentsExitCode = 50

type simcRunner struct {
	binaryPath string
}

type simulationTarget struct {
	region string
	realm  string
	name   string
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

func (runner simcRunner) Run(ctx context.Context, target simulationTarget) ([]byte, error) {
	// #nosec G204 -- validated target fields are interpolated into a single simc armory argument.
	command := exec.CommandContext(
		ctx,
		runner.binaryPath,
		fmt.Sprintf("armory=%s,%s,%s", target.region, target.realm, target.name),
	)

	var output bytes.Buffer

	command.Stdout = &output
	command.Stderr = os.Stderr

	err := command.Run()
	if err != nil {
		return nil, fmt.Errorf("execute simc binary: %w", err)
	}

	return output.Bytes(), nil
}

func simulationTargetFromOptions(options api_types.SimulationOptions) (simulationTarget, error) {
	err := utils.ValidateSimOptions(&options)
	if err != nil {
		return simulationTarget{}, fmt.Errorf("validate simulation options: %w", err)
	}

	return simulationTarget{
		region: string(options.WowCharacter.Region),
		realm:  string(options.WowCharacter.Realm),
		name:   options.WowCharacter.CharacterName,
	}, nil
}
