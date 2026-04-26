package sims

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"reflect"

	"github.com/DomNidy/saint_sim/internal/api"
)

// SimcRunner is a wrapper over the simc binary.
type SimcRunner struct {
	simcBinaryPath string
	workspace      Workspace
}

// NewSimcRunner creates and returns a SimcRunner.
func NewSimcRunner(simcBinaryPath string, workspace Workspace) SimcRunner {
	return SimcRunner{simcBinaryPath: simcBinaryPath, workspace: workspace}
}

// Run executes a simulation end-to-end using the provided plan as the
// "plan".
func (r SimcRunner) Run(
	ctx context.Context,
	plan Plan,
) (api.SimulationResult, error) {
	profileText, err := plan.BuildSimcProfile()
	if err != nil {
		return api.SimulationResult{}, err
	}

	profilePath, cleanupFunc, err := r.workspace.writeSimcProfileTemp(string(profileText))
	defer cleanupFunc()

	if err != nil {
		return api.SimulationResult{}, err
	}

	outputPath := r.workspace.generateOutputPath(profilePath)
	command := exec.CommandContext(ctx, r.simcBinaryPath, profilePath, "json2="+outputPath)

	// harden by making env of the launched process empty
	command.Env = []string{}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr

	err = command.Run()

	if err != nil {
		return api.SimulationResult{}, fmt.Errorf(
			"execute simc binary: %w. stderr: %s\n"+
				"Did the plan for this sim return a valid simc profile?\n"+
				"plan concrete type: %s\n"+
				"Profile content:\n%s",
			err,
			stderr.String(),
			reflect.TypeOf(plan).Name(),
			profileText,
		)
	}

	// try to read json2 sim results
	parsedJson2, err := r.workspace.readSimulationFile(outputPath)
	if err != nil {
		return api.SimulationResult{}, err
	}

	apiRes, err := plan.prepareReportFromRunResult(runResult{
		Stdout: stdout.Bytes(),
		Stderr: stderr.Bytes(),
		JSON2:  parsedJson2,
	})
	if err != nil {
		return api.SimulationResult{}, err
	}

	return apiRes, nil
}

// SIMC EXIT CODES:
// https://github.com/simulationcraft/simc/blob/0842ccb8804867859ece740fc7d72549f46fd305/engine/util/util.hpp#L51
// Exception & Exit Code Handling
// 0: normal exit
// 1: other exceptions
// 30: invalid APL argument
// 40: sim/player/action/buff initialization error
// 50: simulation iteration runtime error
// 51: simulation stuck
// 60: network/file error
// 61: report output error
// 70: invalid sim-scope argument
// 71: invalid fight style
// 80: invalid player-scope argument
// 81: invalid talent string
// 82: invalid item string
