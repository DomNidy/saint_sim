package sims

import (
	"fmt"
	"strings"

	"github.com/DomNidy/saint_sim/internal/api"
)

type BasicSimManifest struct {
	simConfig api.SimulationConfigBasic
}

func NewBasicSimManifest(simConfig api.SimulationConfigBasic) (BasicSimManifest, error) {
	return BasicSimManifest{
		simConfig: simConfig,
	}, nil
}

func (m BasicSimManifest) buildSimcProfile() (simcProfileString, error) {
	var characterName string
	if m.simConfig.Character.Name == nil {
		characterName = "UnknownCharacter"
	} else {
		characterName = *m.simConfig.Character.Name
	}

	baseLines := []string{
		fmt.Sprintf(`%s="%s"`, m.simConfig.Character.CharacterClass, characterName),
		fmt.Sprintf(`level=%v`, m.simConfig.Character.Level),
		fmt.Sprintf(`race=%s`, m.simConfig.Character.Race),
		fmt.Sprintf(`spec=%s`, m.simConfig.Character.Spec),
		"iterations=5", // for testing purposes
	}

	profileText := strings.Join(baseLines, "\n")

	return simcProfileString(profileText), nil
}

// BuildBasicResult projects a completed basic‑sim run into the API's
// `simulation_result_basic` DTO. The raw stdout is carried through as the
// runLog so clients can still render simc's human‑readable report;
// DPS is read from the structured json2 block.
func (manifest BasicSimManifest) prepareReportFromRunResult(
	result runResult,
) (api.SimulationResult, error) {
	out := result.JSON2
	if len(out.Sim.Players) == 0 {
		return api.SimulationResult{}, errSimcNoPlayerResult
	}

	// Basic sims are single‑actor; if simc ever emits more we take the first.
	dps := out.Sim.Players[0].CollectedData.DPS.Mean
	rawLog := string(result.Stdout)

	basicRes := api.SimulationResultBasic{
		Dps:    dps,
		Kind:   api.Basic,
		RawLog: &rawLog,
	}

	var apiRes api.SimulationResult
	err := apiRes.FromSimulationResultBasic(basicRes)
	if err != nil {
		return api.SimulationResult{}, err
	}

	return apiRes, nil
}
