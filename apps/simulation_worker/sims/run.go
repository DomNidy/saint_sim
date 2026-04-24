package sims

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/DomNidy/saint_sim/apps/simulation_worker/json2"
	"github.com/DomNidy/saint_sim/internal/api"
)

const simcNoArgumentsExitCode = 50
const simcProfileFileMode = 0o600

// Run executes a simulation end-to-end using the provided manifest as the
// "plan".
func Run[T Manifest](
	ctx context.Context,
	manifest T,
	simcBinaryPath string,
) (api.SimulationResult, error) {
	profileText, err := manifest.buildSimcProfile()
	if err != nil {
		return api.SimulationResult{}, err
	}

	profilePath, cleanupFunc, err := writeSimcProfileTemp(string(profileText))
	defer cleanupFunc()

	if err != nil {
		return api.SimulationResult{}, err
	}

	outputPath := filepath.Join(filepath.Dir(profilePath), "output.json")
	command := exec.CommandContext(ctx, simcBinaryPath, profilePath, "json2="+outputPath)

	// harden by making env of the launched process empty
	command.Env = []string{}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr

	if err := command.Run(); err != nil {
		return api.SimulationResult{}, fmt.Errorf("execute simc binary: %w", err)
	}

	// try to read json2 sim results
	jsonBytes, err := os.ReadFile(outputPath) // #nosec G304 -- path constructed above
	if err != nil {
		return api.SimulationResult{}, fmt.Errorf("read simc json2 output %q: %w", outputPath, err)
	}

	parsedJson2, err := json2.ParseJSON2(jsonBytes)
	if err != nil {
		return api.SimulationResult{}, err
	}

	result := runResult{
		Stdout: stdout.Bytes(),
		Stderr: stderr.Bytes(),
		JSON2:  parsedJson2,
	}

	apiRes, err := manifest.prepareReportFromRunResult(result)
	if err != nil {
		return api.SimulationResult{}, err
	}

	return apiRes, nil
}

// writeSimcProfileTemp writes a simc profile to disk and return a cleanup function,
// along with the path to the file. Caller needs to call the cleanup function.
func writeSimcProfileTemp(
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
