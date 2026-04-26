package sims

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/DomNidy/saint_sim/apps/simulation_worker/json2"
	"github.com/DomNidy/saint_sim/apps/simulation_worker/set"
	"github.com/DomNidy/saint_sim/internal/api"
)

var (
	errSimcNoPlayerResult       = errors.New("simc produced no player results")
	errSimcNoProfilesetsSection = errors.New(
		"simc produced no profilesets section",
	)
	errPlanHasNoProfilesets                  = errors.New("plan has no profilesets")
	errPlanHasNilProfilesets                 = errors.New("plan has nil profilesets")
	errPlanAndOutputProfilesetsCountMismatch = errors.New(
		"the plan has a different number of profilesets than the corresponding output does",
	)
	errSimcProfilesetUnmatched = errors.New(
		"simc result has no matching plan profileset",
	)
	errTopGearProfilesetSlotMiss = errors.New(
		"profileset is missing a required slot",
	)
)
var (
	errTopGearMissingRawLine       = errors.New("top gear equipment item missing raw_line")
	errTopGearUnsupportedSlot      = errors.New("unsupported top gear slot")
	errTopGearMissingRequiredSlot  = errors.New("missing required top gear slot")
	errTopGearInsufficientRings    = errors.New("need at least two distinct ring candidates")
	errTopGearInsufficientTrinkets = errors.New("need at least two distinct trinket candidates")
	errTopGearMissingEquipment     = errors.New("missing top gear equipment")
	errTopGearProfilesetLimit      = errors.New("top gear profileset count exceeds max")
	errTopGearTooManyCopies        = errors.New(
		"more than two copies of the same ring/trinket were provided",
	)
)

const maxGeneratedProfilesets = 1000
const minPairedItems = 2

// maxCopiesPerPairedItem caps how many stat‑identical copies of a ring or
// trinket may enter the candidate pool. Two is the ceiling because there are
// exactly two finger / two trinket slots; a third copy could never produce a
// distinct loadout.
const maxCopiesPerPairedItem = 2

// noItemIndex marks a profileset slot as intentionally empty (currently only
// off_hand, for two‑handed weapon setups).
const noItemIndex = -1

// emptyOffHandLine is the simc assignment emitted for an empty off‑hand slot.
const emptyOffHandLine = "off_hand=,"

// Profileset is one fully materialized gear loadout.
//
// Each gear field holds an index into the original []api.EquipmentItem the
// request was built from. Storing indices (rather than raw simc lines).
type Profileset struct {
	// name identifies this profileset within the simc input and is the join
	// key against simc's json2 profilesets.results[].name.
	Name string

	Head     int
	Neck     int
	Shoulder int
	Back     int
	Chest    int
	Wrist    int
	Hands    int
	Waist    int
	Legs     int
	Feet     int
	Finger1  int
	Finger2  int
	Trinket1 int
	Trinket2 int
	MainHand int
	OffHand  int // may be noItemIndex

	Talents string
}

// TopGearSimPlan is an abstraction over a topgear sim job.
// it binds generated profilesets to the equipment table their
// slot indices reference. Indices are only meaningful against this exact
// slice, so the two are never exposed separately.
type TopGearSimPlan struct {
	characterName  string // name of the character being simmed
	level          int    // level of the character being simmed
	race           string // race of the character being simmed
	characterClass api.CharacterClass
	spec           string

	equipment   []api.EquipmentItem // defensive copy of the request payload
	Profilesets []Profileset
}

// topGearCandidatePools is the organized set of gear choices we can build
// profilesets from. Every entry is an index into the request's equipment slice.
type topGearCandidatePools struct {
	head     []int
	neck     []int
	shoulder []int
	back     []int
	chest    []int
	wrist    []int
	hands    []int
	waist    []int
	legs     []int
	feet     []int
	rings    []int
	trinkets []int
	mainHand []int
	offHand  []int
}

// NewTopGearSimPlan expands the equipment pools into deterministic
// profilesets. The recursive singleton walk mirrors the counting logic so the
// generated order remains intuitive: earlier input candidates appear earlier in
// the resulting Combo1/Combo2/... sequence.
func NewTopGearSimPlan(
	config api.SimulationConfigTopGear,
) (TopGearSimPlan, error) {
	pools, err := dedupeEquipmentAndPool(config.Equipment)
	if err != nil {
		return TopGearSimPlan{}, err
	}

	plan := TopGearSimPlan{
		equipment: config.Equipment,
	}

	count, err := plan.CountProfilesets()
	if err != nil {
		return TopGearSimPlan{}, err
	}

	if count > maxGeneratedProfilesets {
		return TopGearSimPlan{}, fmt.Errorf(
			"%w: had %v profilesets",
			errTopGearProfilesetLimit,
			count,
		)
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

	return TopGearSimPlan{
		Profilesets:    profilesets,
		equipment:      eqCopy,
		characterName:  characterName,
		level:          config.Character.Level,
		race:           config.Character.Race,
		characterClass: config.Character.CharacterClass,
		spec:           config.Character.Spec,
	}, nil
}

func (plan TopGearSimPlan) BuildSimcProfile() (simcProfileString, error) {
	prof, err := plan.SimcLines()
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
func (plan *TopGearSimPlan) CountProfilesets() (int, error) {
	equipment := plan.equipment
	pools, err := dedupeEquipmentAndPool(equipment)
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

// SimcLines renders the plan to a complete simc profile.
//
// We define the base profile, and then append the lines of all profilesets:
// `profileset."ComboN"+=…`.
func (plan *TopGearSimPlan) SimcLines() ([]string, error) {
	if plan.Profilesets == nil {
		return nil, errPlanHasNilProfilesets
	}

	if len(plan.Profilesets) == 0 {
		return nil, errPlanHasNoProfilesets
	}

	var out []string

	// write the base profile: name, race and level
	// we need to take one of the profilesets and use it to "seed"
	// the base profile with a baseline set of equipment. We'll use the
	// first profileset for this purpose

	baseLines, err := characterBaseRawlines(
		plan.characterClass,
		&plan.characterName,
		plan.level,
		plan.race,
		plan.spec,
	)
	if err != nil {
		return nil, err
	}

	seedProfileset := plan.Profilesets[0]

	baseEquipment, err := equipmentLinesForProfileset(seedProfileset, plan.equipment)
	if err != nil {
		return nil, err
	}

	out = append(out, baseLines...)
	out = append(out, baseEquipment...)

	// then, add all of the profileset lines!
	for i := range plan.Profilesets {
		lines, err := plan.Profilesets[i].lines(plan.equipment)
		if err != nil {
			return nil, err
		}
		out = append(out, lines...)
	}

	return out, nil
}

func (plan *TopGearSimPlan) Equipment() []api.EquipmentItem { return plan.equipment }
func (plan *TopGearSimPlan) Len() int                       { return len(plan.Profilesets) }

// prepareReportFromRunResult joins the simulation output to the api shape.
func (plan TopGearSimPlan) prepareReportFromRunResult(
	result runResult,
) (api.SimulationResult, error) {
	out := result.JSON2
	if out.Sim.Profilesets == nil {
		return api.SimulationResult{}, errSimcNoProfilesetsSection
	}

	if plan.Profilesets == nil {
		return api.SimulationResult{}, errPlanHasNilProfilesets
	}

	if len(plan.Profilesets) == 0 {
		return api.SimulationResult{}, errPlanHasNoProfilesets
	}

	if len(out.Sim.Profilesets.Results) != len(plan.Profilesets) {
		return api.SimulationResult{}, errPlanAndOutputProfilesetsCountMismatch
	}

	// Index simc's results for O(1) lookup. Map size is bounded by
	// maxGeneratedProfilesets, so allocation is trivial.
	byName := make(map[string]json2.JSON2ProfilesetResult, len(out.Sim.Profilesets.Results))
	for _, r := range out.Sim.Profilesets.Results {
		byName[r.Name] = r
	}

	entries := make([]api.TopGearProfilesetResult, 0, plan.Len())
	for _, planPset := range plan.Profilesets {
		// match the profileset in the plan to the profileset
		// stored in the output json2 result from simc
		name := planPset.Name
		metric, ok := byName[name]
		if !ok {
			return api.SimulationResult{}, fmt.Errorf(
				"%w: %q",
				errSimcProfilesetUnmatched,
				name,
			)
		}

		items := mapProfilesetToTopGearProfilesetItems(planPset)

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
		Equipment:   plan.Equipment(),
		Profilesets: entries,
	}

	var apiRes api.SimulationResult
	err := apiRes.FromSimulationResultTopGear(topGearRes)
	if err != nil {
		return api.SimulationResult{}, err
	}

	return apiRes, nil
}

// newProfileset returns a profileset with every gear slot set to noItemIndex
// so an accidentally‑unset slot is loudly wrong (index ‑1) rather than
// silently aliasing equipment[0].
func newProfileset() Profileset {
	return Profileset{
		Head: noItemIndex, Neck: noItemIndex, Shoulder: noItemIndex,
		Back: noItemIndex, Chest: noItemIndex, Wrist: noItemIndex,
		Hands: noItemIndex, Waist: noItemIndex, Legs: noItemIndex,
		Feet: noItemIndex, Finger1: noItemIndex, Finger2: noItemIndex,
		Trinket1: noItemIndex, Trinket2: noItemIndex,
		MainHand: noItemIndex, OffHand: noItemIndex,
	}
}

// singletonPool represents a normal gear slot where we simply pick one item
// from the available choices. The items in the pool are repesented as indices
// into an array, the array is implicitly the equipment array of the TopGearSimPlan
// that produced this singletonPool.
type singletonPool struct {
	slot    string
	indices []int
	set     func(*Profileset, int)
}

// slotPair is one chosen pair of interchangeable items (rings or trinkets),
// expressed as equipment indices.
type slotPair struct {
	first  int
	second int
}

// simIdentity returns a canonical key for "same item as far as SimC is
// concerned". Two EquipmentItems with the same simIdentity will produce
// identical stats when assigned to the same slot, so the worker treats them as
// copies of one item.
//
// Deliberately excludes:
//   - Slot prefix (finger1= vs finger2= is an assignment detail, not an item
//     property)
//   - Source (equipped vs bag is inventory metadata, irrelevant to simc)
func simIdentity(item api.EquipmentItem) string {
	_, attrs, _ := strings.Cut(item.RawLine, "=")

	return strings.ToLower(strings.TrimSpace(attrs))
}

// dedupeEquipmentAndPool dupdes a list of Equipment and returns a "pool"
// for each slot.
//
// Dedup rules:
//   - Singleton slots: keep one entry per simIdentity. Two items that produce
//     identical simc assignments would only generate duplicate profilesets.
//   - Rings/trinkets: keep up to 2 entries per simIdentity, so owning two
//     stat‑identical rings still allows an (A, A) pair. A third copy is
//     rejected — it could never yield a new loadout.
//
// Pools:
//
// Each equipment slot has a pool, which is the set of all unique items in
// the top gear options payload for that slot. We implement this as: each
// slot has a slice of integers that index into the equipment slice provided
// to this method.
func dedupeEquipmentAndPool(
	equipment []api.EquipmentItem,
) (topGearCandidatePools, error) {
	var pools topGearCandidatePools

	singletonDestinations := map[api.EquipmentSlot]*[]int{
		api.Head:     &pools.head,
		api.Neck:     &pools.neck,
		api.Shoulder: &pools.shoulder,
		api.Back:     &pools.back,
		api.Chest:    &pools.chest,
		api.Wrist:    &pools.wrist,
		api.Hands:    &pools.hands,
		api.Waist:    &pools.waist,
		api.Legs:     &pools.legs,
		api.Feet:     &pools.feet,
		api.MainHand: &pools.mainHand,
		api.OffHand:  &pools.offHand,
	}

	singletonSeen := map[api.EquipmentSlot]*set.Set[string]{}
	appendSingleton := func(idx int, item api.EquipmentItem, dest *[]int) {
		seen, ok := singletonSeen[item.Slot]
		if !ok {
			seen = set.New[string]()
			singletonSeen[item.Slot] = seen
		}
		if seen.Add(simIdentity(item)) {
			*dest = append(*dest, idx)
		}
	}

	ringCopies, trinketCopies := map[string]int{}, map[string]int{}
	appendPaired := func(copies map[string]int, idx int, item api.EquipmentItem, dest *[]int) error {
		id := simIdentity(item)
		if copies[id] >= maxCopiesPerPairedItem {
			return fmt.Errorf("%w: %s", errTopGearTooManyCopies, item.DisplayName)
		}
		copies[id]++
		*dest = append(*dest, idx)

		return nil
	}

	for idx, item := range equipment {
		if item.RawLine == "" {
			return topGearCandidatePools{}, fmt.Errorf(
				"%w: %q",
				errTopGearMissingRawLine,
				item.Slot,
			)
		}

		switch {
		case isRingSlot(item.Slot):
			if err := appendPaired(ringCopies, idx, item, &pools.rings); err != nil {
				return topGearCandidatePools{}, err
			}
		case isTrinketSlot(item.Slot):
			if err := appendPaired(trinketCopies, idx, item, &pools.trinkets); err != nil {
				return topGearCandidatePools{}, err
			}
		case isIgnoredTopGearSlot(item.Slot):
			// cosmetic slots don't participate in top‑gear generation
		default:
			dest, ok := singletonDestinations[item.Slot]
			if !ok {
				return topGearCandidatePools{}, fmt.Errorf(
					"%w: %q",
					errTopGearUnsupportedSlot,
					item.Slot,
				)
			}
			appendSingleton(idx, item, dest)
		}
	}

	return pools, nil
}

func (p topGearCandidatePools) singletonPools() []singletonPool {
	return []singletonPool{
		{slot: "head", indices: p.head, set: setHead},
		{slot: "neck", indices: p.neck, set: setNeck},
		{slot: "shoulder", indices: p.shoulder, set: setShoulder},
		{slot: "back", indices: p.back, set: setBack},
		{slot: "chest", indices: p.chest, set: setChest},
		{slot: "wrist", indices: p.wrist, set: setWrist},
		{slot: "hands", indices: p.hands, set: setHands},
		{slot: "waist", indices: p.waist, set: setWaist},
		{slot: "legs", indices: p.legs, set: setLegs},
		{slot: "feet", indices: p.feet, set: setFeet},
		{slot: "main_hand", indices: p.mainHand, set: setMainHand},
	}
}

func appendCompletedProfilesets(
	profilesets *[]Profileset,
	current Profileset,
	ringPairs []slotPair,
	trinketPairs []slotPair,
	offHandOptions []int,
	startingIndex int,
) int {
	nextComboIndex := startingIndex

	// Rings and trinkets are pooled separately because their slots are
	// interchangeable. We only emit unordered pairs, so A/B appears once
	// instead of again as the mirror swap.
	for _, rings := range ringPairs {
		withRings := current
		withRings.Finger1 = rings.first
		withRings.Finger2 = rings.second

		for _, trinkets := range trinketPairs {
			withTrinkets := withRings
			withTrinkets.Trinket1 = trinkets.first
			withTrinkets.Trinket2 = trinkets.second

			for _, offHand := range offHandOptions {
				complete := withTrinkets
				complete.OffHand = offHand
				complete.Name = fmt.Sprintf("Combo%d", nextComboIndex)
				nextComboIndex++
				*profilesets = append(*profilesets, complete)
			}
		}
	}

	return nextComboIndex
}

func isRingSlot(slot api.EquipmentSlot) bool {
	return slot == api.Finger1 || slot == api.Finger2
}

func isTrinketSlot(slot api.EquipmentSlot) bool {
	return slot == api.Trinket1 || slot == api.Trinket2
}

func isIgnoredTopGearSlot(slot api.EquipmentSlot) bool {
	return slot == api.Shirt || slot == api.Tabard
}

func setHead(p *Profileset, i int)     { p.Head = i }
func setNeck(p *Profileset, i int)     { p.Neck = i }
func setShoulder(p *Profileset, i int) { p.Shoulder = i }
func setBack(p *Profileset, i int)     { p.Back = i }
func setChest(p *Profileset, i int)    { p.Chest = i }
func setWrist(p *Profileset, i int)    { p.Wrist = i }
func setHands(p *Profileset, i int)    { p.Hands = i }
func setWaist(p *Profileset, i int)    { p.Waist = i }
func setLegs(p *Profileset, i int)     { p.Legs = i }
func setFeet(p *Profileset, i int)     { p.Feet = i }
func setMainHand(p *Profileset, i int) { p.MainHand = i }

// lines converts the loadout into the `profileset."Name"+=…` lines that will
// be appended to the simc profile. equipment must be the same slice that
// buildTopGearCandidatePools was called with.
func (l *Profileset) lines(equipment []api.EquipmentItem) ([]string, error) {
	equipmentLines, err := equipmentLinesForProfileset(*l, equipment)
	if err != nil {
		return nil, err
	}

	const exLinesPerProfileset = 17 // 16 gear + 1 talents
	lines := make([]string, 0, exLinesPerProfileset)

	emit := func(raw string) {
		lines = append(lines, fmt.Sprintf(`profileset."%s"+=%s`, l.Name, raw))
	}

	for _, equipmentLine := range equipmentLines {
		emit(equipmentLine)
	}

	talentsLine, err := talentsRawline(l.Talents)
	if err != nil {
		return nil, err
	}

	emit(talentsLine)

	return lines, nil
}

func equipmentLinesForProfileset(pset Profileset, equipment []api.EquipmentItem) ([]string, error) {
	type slotLine struct {
		idx  int
		slot api.EquipmentSlot
	}

	slots := []slotLine{
		{pset.Head, api.Head}, {pset.Neck, api.Neck}, {pset.Shoulder, api.Shoulder},
		{pset.Back, api.Back}, {pset.Chest, api.Chest}, {pset.Wrist, api.Wrist},
		{pset.Hands, api.Hands}, {pset.Waist, api.Waist}, {pset.Legs, api.Legs},
		{pset.Feet, api.Feet}, {pset.Finger1, api.Finger1}, {pset.Finger2, api.Finger2},
		{pset.Trinket1, api.Trinket1}, {pset.Trinket2, api.Trinket2},
		{pset.MainHand, api.MainHand}, {pset.OffHand, api.OffHand},
	}

	lines := make([]string, 0, len(slots))
	for _, slot := range slots {
		switch {
		case slot.idx == noItemIndex && slot.slot == api.OffHand: // case where the offhand is empty (allow to proceed)
			lines = append(lines, emptyOffHandLine)
		case slot.idx == noItemIndex: // we require all other slots to be filled
			return nil, fmt.Errorf("%w: %q", errTopGearProfilesetSlotMiss, slot.slot)
		default:
			line, err := equipmentRawlineForSlot(equipment[slot.idx], slot.slot)
			if err != nil {
				return nil, err
			}
			lines = append(lines, line)
		}
	}

	return lines, nil
}
