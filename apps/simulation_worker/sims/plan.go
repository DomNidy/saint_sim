package sims

import (
	"github.com/DomNidy/saint_sim/apps/simulation_worker/json2"
	"github.com/DomNidy/saint_sim/internal/api"
)

// type alias for a simc profile string that is "final"
// final as in: sanitized, and ready to be written to disk
// and passed to simc.
type simcProfileString string

// Plan is a "plan" for how to perform a simulation.
type Plan interface {
	// BuildSimcProfile converts the plan into the final simc profile text
	// the profile text should be writable directly to a simc profile without any
	// processing.
	BuildSimcProfile() (simcProfileString, error)

	// prepareReportFromRunResult takes the artifacts produced by executing this
	// simulation, and then returns the appropriate api-shaped result object for
	// the plan type.
	//
	// For example, a Basic Simulation would return a Basic Simulation Result
	prepareReportFromRunResult(result runResult) (api.SimulationResult, error)
}

// runResult holds the artifacts produced by a single simc invocation.
type runResult struct {
	// Stdout is the human‑readable log simc writes to standard output.
	Stdout []byte

	// Stderr is the stderr output from simc
	Stderr []byte

	// JSON2 is the raw contents of the file produced by passing `json2=<path>`
	// to simc
	JSON2 json2.JSON2Output
}
