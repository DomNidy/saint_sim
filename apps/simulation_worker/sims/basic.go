package sims

import (
	"github.com/DomNidy/saint_sim/apps/simulation_worker/json2"
	"github.com/DomNidy/saint_sim/internal/api"
)

type BasicSimManifest struct {
}

// BuildBasicResult projects a completed basic‑sim run into the API's
// `simulation_result_basic` DTO. The raw stdout is carried through as the
// runLog so clients can still render simc's human‑readable report;
// DPS is read from the structured json2 block.
func (manifest BasicSimManifest) BuildResultFromJSON2(
	out json2.JSON2Output,
) (RunResult, error) {
	if len(out.Sim.Players) == 0 {
		return RunResult{}, errSimcNoPlayerResult
	}

	// Basic sims are single‑actor; if simc ever emits more we take the first.
	dps := out.Sim.Players[0].CollectedData.DPS.Mean
	rawLog := "" /* todo: thread the raw stdout into here somehow - probably refactor more */

	return RunResult{
		Stdout: make([]byte, 0),
		JSON2:  out,
		Data: api.SimulationResultBasic{
			Kind:   api.Basic,
			Dps:    dps,
			RawLog: &rawLog,
		},
	}, nil
}

func (manifest BasicSimManifest) BuildSimcProfile() (simcProfileString, error) {
	return "", nil
}
