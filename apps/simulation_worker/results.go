package main

import (
	"errors"
	"fmt"
	"sort"

	"github.com/DomNidy/saint_sim/internal/api"
	"github.com/DomNidy/saint_sim/internal/simc"
)

var (
	errSimcNoPlayerResult       = errors.New("simc produced no player results")
	errSimcNoProfilesetsSection = errors.New(
		"simc produced no profilesets section",
	)
	errManifestHasNoProfilesets                  = errors.New("manifest has no profilesets")
	errManifestHasNilProfilesets                 = errors.New("manifest has nil profilesets")
	errManifestAndOutputProfilesetsCountMismatch = errors.New(
		"the manifest has a different number of profilesets than the corresponding output does",
	)
	errSimcProfilesetUnmatched = errors.New(
		"simc result has no matching manifest profileset",
	)
	errTopGearProfilesetSlotMiss = errors.New(
		"profileset is missing a required slot",
	)
)

// buildBasicResult projects a completed basic‑sim run into the API's
// `simulation_result_basic` DTO. The raw stdout is carried through as the
// optional raw_log so clients can still render simc's human‑readable report;
// DPS is read from the structured json2 block.
func buildBasicResult(run RunResult, out simc.JSON2Output) (api.SimulationResultBasic, error) {
	if len(out.Sim.Players) == 0 {
		return api.SimulationResultBasic{}, errSimcNoPlayerResult
	}

	// Basic sims are single‑actor; if simc ever emits more we take the first.
	dps := out.Sim.Players[0].CollectedData.DPS.Mean
	rawLog := string(run.Stdout)

	return api.SimulationResultBasic{
		Kind:   api.SimulationResultBasicKindBasic,
		Dps:    dps,
		RawLog: &rawLog,
	}, nil
}

// buildTopGearResult joins the worker's loadout manifest with simc's
// per‑profileset metrics, sorts by mean descending, and returns the API
// DTO that gets persisted verbatim to simulation.sim_result.
//
// The join key is the profileset name (e.g. "Combo7"), which is how the
// worker labels its generated stanzas and how simc echoes them back in
// sim.profilesets.results[].name.
func buildTopGearResult(
	manifest *topGearManifest,
	out simc.JSON2Output,
) (api.SimulationResultTopGear, error) {
	if out.Sim.Profilesets == nil {
		return api.SimulationResultTopGear{}, errSimcNoProfilesetsSection
	}

	if manifest.profilesets == nil {
		return api.SimulationResultTopGear{}, errManifestHasNilProfilesets
	}

	if len(manifest.profilesets) == 0 {
		return api.SimulationResultTopGear{}, errManifestHasNoProfilesets
	}

	if len(out.Sim.Profilesets.Results) != len(manifest.profilesets) {
		return api.SimulationResultTopGear{}, errManifestAndOutputProfilesetsCountMismatch
	}

	// Index simc's results for O(1) lookup. Map size is bounded by
	// maxGeneratedProfilesets, so allocation is trivial.
	byName := make(map[string]simc.JSON2ProfilesetResult, len(out.Sim.Profilesets.Results))
	for _, r := range out.Sim.Profilesets.Results {
		byName[r.Name] = r
	}

	entries := make([]api.TopGearProfilesetResult, 0, manifest.Len())
	for _, manifestPset := range manifest.profilesets {
		// match the profileset in the manifest to the profileset
		// stored in the output json2 result from simc
		name := manifestPset.name
		metric, ok := byName[name]
		if !ok {
			return api.SimulationResultTopGear{}, fmt.Errorf(
				"%w: %q",
				errSimcProfilesetUnmatched,
				name,
			)
		}

		items := mapProfilesetToTopGearProfilesetItems(manifestPset)

		meanError := metric.MeanError
		entries = append(entries, api.TopGearProfilesetResult{
			Name:      name,
			Mean:      metric.Mean,
			MeanError: &meanError,
			Items:     items,
		})
	}

	// Best first. SliceStable keeps the Combo1/Combo2/… order on ties so
	// equivalent loadouts render deterministically.
	sort.SliceStable(entries, func(i, j int) bool {
		return entries[i].Mean > entries[j].Mean
	})

	return api.SimulationResultTopGear{
		Kind:        api.SimulationResultTopGearKindTopGear,
		Metric:      out.Sim.Profilesets.Metric,
		Equipment:   manifest.Equipment(),
		Profilesets: entries,
	}, nil
}

// mapProfilesetToTopGearProfilesetItems does a simple mapping from
// profileset to the API shape.
func mapProfilesetToTopGearProfilesetItems(pset profileset) api.TopGearProfilesetItems {
	return api.TopGearProfilesetItems{
		Back:     pset.back,
		Chest:    pset.chest,
		Feet:     pset.feet,
		Finger1:  pset.finger1,
		Finger2:  pset.finger2,
		Hands:    pset.hands,
		Head:     pset.head,
		Legs:     pset.legs,
		MainHand: pset.mainHand,
		Neck:     pset.neck,

		OffHand:  &pset.offHand,
		Shoulder: pset.shoulder,
		Trinket1: pset.trinket1,
		Trinket2: pset.trinket2,
		Waist:    pset.waist,
		Wrist:    pset.wrist,
	}
}
