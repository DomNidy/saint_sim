package sims

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/DomNidy/saint_sim/apps/simulation_worker/json2"
	"github.com/DomNidy/saint_sim/internal/api"
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

// TopGearManifest is an abstraction over a topgear sim job.
// it binds generated profilesets to the equipment table their
// slot indices reference. Indices are only meaningful against this exact
// slice, so the two are never exposed separately.
type TopGearManifest struct {
	characterName  string // name of the character being simmed
	level          int    // level of the character being simmed
	race           string // race of the character being simmed
	characterClass api.CharacterClass
	spec           string

	equipment   []api.EquipmentItem // defensive copy of the request payload
	Profilesets []Profileset
}

// NewTopGearManifest expands the equipment pools into deterministic
// profilesets. The recursive singleton walk mirrors the counting logic so the
// generated order remains intuitive: earlier input candidates appear earlier in
// the resulting Combo1/Combo2/... sequence.
func NewTopGearManifest(
	config api.SimulationConfigTopGear,
) (TopGearManifest, error) {
	pools, err := buildTopGearCandidatePools(config.Equipment)
	if err != nil {
		return TopGearManifest{}, err
	}

	manifest := TopGearManifest{
		equipment: config.Equipment,
	}

	count, err := manifest.CountProfilesets()
	if err != nil {
		return TopGearManifest{}, err
	}

	singletonPools := pools.singletonPools()
	ringPairs := makeUnorderedPairs(pools.rings)
	trinketPairs := makeUnorderedPairs(pools.trinkets)

	profilesets := make([]Profileset, 0, count)
	base := newProfileset()
	base.Talents = config.Character.ActiveTalents.Talents

	offHandOptions := pools.offHand
	if len(offHandOptions) == 0 {
		offHandOptions = []int{noItemIndex}
	}

	comboIndex := 1

	var buildSingletons func(poolIndex int, current Profileset)
	buildSingletons = func(poolIndex int, current Profileset) {
		if poolIndex == len(singletonPools) {
			comboIndex = appendCompletedProfilesets(
				&profilesets,
				current,
				ringPairs,
				trinketPairs,
				offHandOptions,
				comboIndex,
			)

			return
		}

		pool := singletonPools[poolIndex]
		for _, idx := range pool.indices {
			next := current
			pool.set(&next, idx)
			buildSingletons(poolIndex+1, next)
		}
	}

	buildSingletons(0, base)

	eqCopy := make([]api.EquipmentItem, len(config.Equipment))
	copy(eqCopy, config.Equipment)

	characterName := "UnknownCharacter"
	if config.Character.Name != nil && *config.Character.Name != "" {
		characterName = *config.Character.Name
	}

	return TopGearManifest{
		Profilesets:    profilesets,
		equipment:      eqCopy,
		characterName:  characterName,
		level:          config.Character.Level,
		race:           config.Character.Race,
		characterClass: config.Character.CharacterClass,
		spec:           config.Character.Spec,
	}, nil
}

func (manifest TopGearManifest) buildSimcProfile() (simcProfileString, error) {
	prof, err := manifest.SimcLines()
	if err != nil {
		return "", err
	}

	profText := strings.Join(prof, "\n")

	return simcProfileString(profText), nil
}

// mapProfilesetToTopGearProfilesetItems does a simple mapping from
// profileset to the API shape.
func mapProfilesetToTopGearProfilesetItems(pset Profileset) api.TopGearProfilesetItems {
	return api.TopGearProfilesetItems{
		Back:     pset.Back,
		Chest:    pset.Chest,
		Feet:     pset.Feet,
		Finger1:  pset.Finger1,
		Finger2:  pset.Finger2,
		Hands:    pset.Hands,
		Head:     pset.Head,
		Legs:     pset.Legs,
		MainHand: pset.MainHand,
		Neck:     pset.Neck,

		OffHand:  &pset.OffHand,
		Shoulder: pset.Shoulder,
		Trinket1: pset.Trinket1,
		Trinket2: pset.Trinket2,
		Waist:    pset.Waist,
		Wrist:    pset.Wrist,
	}
}

func unorderedPairCount(itemCount int) int {
	if itemCount < minPairedItems {
		return 0
	}

	return itemCount * (itemCount - 1) / minPairedItems
}

func makeUnorderedPairs(indices []int) []slotPair {
	pairs := make([]slotPair, 0, unorderedPairCount(len(indices)))
	for left := range indices {
		for right := left + 1; right < len(indices); right++ {
			pairs = append(pairs, slotPair{first: indices[left], second: indices[right]})
		}
	}

	return pairs
}

func retargetEquipmentLine(line string, slot api.EquipmentSlot) string {
	_, rest, found := strings.Cut(line, "=")
	if !found {
		return line
	}

	return string(slot) + "=" + rest
}

// CountProfilesets computes the number of valid profilesets implied by
// the equipment payload without allocating the final profileset slice.
func (manifest *TopGearManifest) CountProfilesets() (int, error) {
	equipment := manifest.equipment
	pools, err := buildTopGearCandidatePools(equipment)
	if err != nil {
		return 0, err
	}

	total := 1

	for _, pool := range pools.singletonPools() {
		if len(pool.indices) == 0 {
			return 0, fmt.Errorf("%w: %q", errTopGearMissingRequiredSlot, pool.slot)
		}
		total *= len(pool.indices)
	}

	if len(pools.rings) < minPairedItems {
		return 0, fmt.Errorf("%w: got %d", errTopGearInsufficientRings, len(pools.rings))
	}
	total *= unorderedPairCount(len(pools.rings))

	if len(pools.trinkets) < minPairedItems {
		return 0, fmt.Errorf("%w: got %d", errTopGearInsufficientTrinkets, len(pools.trinkets))
	}
	total *= unorderedPairCount(len(pools.trinkets))

	// off_hand is intentionally optional for two‑handed weapon setups.
	if len(pools.offHand) > 0 {
		total *= len(pools.offHand)
	}

	return total, nil
}

// SimcLines renders the manifest to a complete simc profile.
//
// We define the base profile, and then append the lines of all profilesets:
// `profileset."ComboN"+=…`.
func (m *TopGearManifest) SimcLines() ([]string, error) {
	if m.Profilesets == nil {
		return nil, errManifestHasNilProfilesets
	}

	if len(m.Profilesets) == 0 {
		return nil, errManifestHasNoProfilesets
	}

	var out []string

	// write the base profile: name, race and level
	// we need to take one of the profilesets and use it to "seed"
	// the base profile with a baseline set of equipment. We'll use the
	// first profileset for this purpose

	baseLines, err := characterBaseRawlines(
		m.characterClass,
		&m.characterName,
		m.level,
		m.race,
		m.spec,
	)
	if err != nil {
		return nil, err
	}

	seedProfileset := m.Profilesets[0]

	baseEquipment, err := equipmentLinesForProfileset(seedProfileset, m.equipment)
	if err != nil {
		return nil, err
	}

	out = append(out, baseLines...)
	out = append(out, baseEquipment...)

	// then, add all of the profileset lines!
	for i := range m.Profilesets {
		lines, err := m.Profilesets[i].lines(m.equipment)
		if err != nil {
			return nil, err
		}
		out = append(out, lines...)
	}

	return out, nil
}

func (m *TopGearManifest) Equipment() []api.EquipmentItem { return m.equipment }
func (m *TopGearManifest) Len() int                       { return len(m.Profilesets) }

// prepareReportFromRunResult joins the simulation output to the api shape.
func (manifest TopGearManifest) prepareReportFromRunResult(
	result runResult,
) (api.SimulationResult, error) {
	out := result.JSON2
	if out.Sim.Profilesets == nil {
		return api.SimulationResult{}, errSimcNoProfilesetsSection
	}

	if manifest.Profilesets == nil {
		return api.SimulationResult{}, errManifestHasNilProfilesets
	}

	if len(manifest.Profilesets) == 0 {
		return api.SimulationResult{}, errManifestHasNoProfilesets
	}

	if len(out.Sim.Profilesets.Results) != len(manifest.Profilesets) {
		return api.SimulationResult{}, errManifestAndOutputProfilesetsCountMismatch
	}

	// Index simc's results for O(1) lookup. Map size is bounded by
	// maxGeneratedProfilesets, so allocation is trivial.
	byName := make(map[string]json2.JSON2ProfilesetResult, len(out.Sim.Profilesets.Results))
	for _, r := range out.Sim.Profilesets.Results {
		byName[r.Name] = r
	}

	entries := make([]api.TopGearProfilesetResult, 0, manifest.Len())
	for _, manifestPset := range manifest.Profilesets {
		// match the profileset in the manifest to the profileset
		// stored in the output json2 result from simc
		name := manifestPset.Name
		metric, ok := byName[name]
		if !ok {
			return api.SimulationResult{}, fmt.Errorf(
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

	topGearRes := api.SimulationResultTopGear{
		Kind:        api.SimulationResultTopGearKindTopGear,
		Metric:      out.Sim.Profilesets.Metric,
		Equipment:   manifest.Equipment(),
		Profilesets: entries,
	}

	var apiRes api.SimulationResult
	err := apiRes.FromSimulationResultTopGear(topGearRes)
	if err != nil {
		return api.SimulationResult{}, err
	}

	return apiRes, nil
}
