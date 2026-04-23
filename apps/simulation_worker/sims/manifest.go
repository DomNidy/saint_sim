package sims

import "github.com/DomNidy/saint_sim/apps/simulation_worker/json2"

// type alias for a simc profile string that is "final"
// final as in: sanitized, and ready to be written to disk
// and passed to simc
type simcProfileString string

type Manifest interface {
	// BuildSimcProfile converts the manifest into the final simc profile text
	// the profile text should be writable str
	BuildSimcProfile() (simcProfileString, error)

	// BuildResultFromJSON2 takes the output a JSON2 object produced
	// by executing this simulation, and then returns the appropriate
	// api-shaped result object.
	BuildResultFromJSON2(json2 json2.JSON2Output) (RunResult, error)
}

// RunResult holds the artifacts produced by a single simc invocation.
type RunResult struct {
	// Stdout is the human‑readable log simc writes to standard output.
	Stdout []byte

	// JSON2 is the raw contents of the file produced by passing `json2=<path>`
	// to simc
	JSON2 json2.JSON2Output

	// API-shaped structured result object for this simulation
	// each Manifest will need to marshal into this from the
	// JSON2 output
	Data interface{}
}
