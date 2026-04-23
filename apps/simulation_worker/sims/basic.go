package sims

import (
	"github.com/DomNidy/saint_sim/apps/simulation_worker/json2"
	"github.com/DomNidy/saint_sim/internal/api"
)

// BuildBasicResult projects a completed basic‑sim run into the API's
// `simulation_result_basic` DTO. The raw stdout is carried through as the
// runLog so clients can still render simc's human‑readable report;
// DPS is read from the structured json2 block.
func BuildBasicResult(runLog string, out json2.JSON2Output) (api.SimulationResultBasic, error) {
	if len(out.Sim.Players) == 0 {
		return api.SimulationResultBasic{}, errSimcNoPlayerResult
	}

	// Basic sims are single‑actor; if simc ever emits more we take the first.
	dps := out.Sim.Players[0].CollectedData.DPS.Mean
	rawLog := runLog

	return api.SimulationResultBasic{
		Kind:   api.Basic,
		Dps:    dps,
		RawLog: &rawLog,
	}, nil
}
