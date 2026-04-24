package sims

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/DomNidy/saint_sim/apps/simulation_worker/json2"
)

const simcProfileFileMode = 0o600

var (
	ErrSimcProfileTooBig = errors.New("simc profile was too big")
)

// Workspace abstracts access to the file system from the simc runner.
type Workspace struct {
	// Maximum size of a simc profile to write to disk.
	// If exceeded, fail writing
	MaxSimcProfileSizeBytes int
}

// writeSimcProfileTemp writes a simc profile to disk and return a cleanup function,
// along with the path to the file. If an error is returned, no content remains on disk.
// If an error is not returned, then the caller must call the cleanup function to avoid
// leaking files/dirs.
func (w Workspace) writeSimcProfileTemp(
	profileText string,
) (string, func(), error) {
	if len(profileText) > w.MaxSimcProfileSizeBytes {
		return "", func() {}, fmt.Errorf(
			"%w: was %v bytes, max allowed: %v",
			ErrSimcProfileTooBig,
			len(profileText),
			w.MaxSimcProfileSizeBytes,
		)
	}

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
	err = os.WriteFile(profilePath, []byte(profileText), simcProfileFileMode)
	if err != nil {
		cleanupFunc()

		return "", func() {}, fmt.Errorf("write simc profile: %w", err)
	}

	return profilePath, cleanupFunc, nil
}

// generateOutputPath takes a profilePath and then returns a path
// to emit the output of that profile's sim at.
func (w Workspace) generateOutputPath(profilePath string) string {
	return filepath.Join(filepath.Dir(profilePath), "output.json")
}

func (w Workspace) readSimulationFile(outputPath string) (json2.JSON2Output, error) {
	jsonBytes, err := os.ReadFile(outputPath) // #nosec G304 -- path constructed above
	if err != nil {
		return json2.JSON2Output{}, fmt.Errorf("read simc json2 output %q: %w", outputPath, err)
	}

	parsedJson2, err := json2.ParseJSON2(jsonBytes)
	if err != nil {
		return json2.JSON2Output{}, err
	}

	return parsedJson2, nil
}
