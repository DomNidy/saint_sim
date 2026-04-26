package sims

import (
	"strings"

	"github.com/DomNidy/saint_sim/internal/api"
)

type BasicSimPlan struct {
	simConfig api.SimulationConfigBasic
}

func NewBasicSimPlan(simConfig api.SimulationConfigBasic) (BasicSimPlan, error) {
	return BasicSimPlan{
		simConfig: simConfig,
	}, nil
}

func (plan BasicSimPlan) BuildSimcProfile() (simcProfileString, error) {
	character := plan.simConfig.Character
	baseLines, err := characterBaseRawlines(
		character.CharacterClass,
		character.Name,
		character.Level,
		character.Race,
		character.Spec,
	)

	if err != nil {
		return "", err
	}

	baseLines = append(baseLines, "target_error=0.2")

	equipmentLines, err := equippedItemsRawlines(character.EquippedItems)
	if err != nil {
		return "", err
	}
	baseLines = append(baseLines, equipmentLines...)

	talentsLines, err := talentsRawline(character.ActiveTalents.Talents)
	if err != nil {
		return "", err
	}
	baseLines = append(baseLines, talentsLines)

	if plan.simConfig.CoreConfig.FightStyle != nil {
		fightStyleLine, err := fightStyleRawline(*plan.simConfig.CoreConfig.FightStyle)
		if err != nil {
			return "", err
		}
		baseLines = append(baseLines, fightStyleLine)
	}

	profileText := strings.Join(baseLines, "\n")

	return simcProfileString(profileText), nil
}

// BuildBasicResult projects a completed basic‑sim run into the API's
// `simulation_result_basic` DTO. The raw stdout is carried through as the
// runLog so clients can still render simc's human‑readable report;
// DPS is read from the structured json2 block.
func (plan BasicSimPlan) prepareReportFromRunResult(
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
