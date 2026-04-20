package simc

import (
	"encoding/json"
	"fmt"
)

// JSON2Output is the subset of simc's `json2=` report that Saint consumes.
// The full report is large and evolves between simc releases; only fields we
// actually read are modeled. encoding/json ignores unknown fields, so this
// stays forward‑compatible with simc upgrades.
type JSON2Output struct {
	Version string   `json:"version"`
	Sim     JSON2Sim `json:"sim"`
}

type JSON2Sim struct {
	// Profilesets is only present when the input profile contained
	// `profileset."…"` stanzas (i.e. top‑gear runs). Basic sims omit it.
	Profilesets *JSON2Profilesets `json:"profilesets,omitempty"`
}

type JSON2Profilesets struct {
	// Metric is the statistic each result's Mean refers to — typically "dps"
	// for damage specs.
	Metric  string                  `json:"metric"`
	Results []JSON2ProfilesetResult `json:"results"`
}

// JSON2ProfilesetResult is one entry in sim.profilesets.results.
//
// Name matches the `profileset."<Name>"` label the worker generated
// (e.g. "Combo7"); that name is the join key back to the worker's
// loadout manifest.
type JSON2ProfilesetResult struct {
	Name       string  `json:"name"`
	Mean       float64 `json:"mean"`
	Min        float64 `json:"min"`
	Max        float64 `json:"max"`
	Stddev     float64 `json:"stddev"`
	MeanStddev float64 `json:"mean_stddev"`
	MeanError  float64 `json:"mean_error"`
	Median     float64 `json:"median"`
	Iterations int     `json:"iterations"`
}

// ParseJSON2 decodes a simc `json2=` report into the reduced model above.
func ParseJSON2(raw []byte) (JSON2Output, error) {
	var out JSON2Output
	if err := json.Unmarshal(raw, &out); err != nil {
		return JSON2Output{}, fmt.Errorf("decode simc json2 output: %w", err)
	}

	return out, nil
}
