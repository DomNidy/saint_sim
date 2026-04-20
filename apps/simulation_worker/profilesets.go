package main

import (
	"errors"
	"fmt"
	"strings"

	"github.com/DomNidy/saint_sim/apps/simulation_worker/set"
	"github.com/DomNidy/saint_sim/internal/api"
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

// profileset is one fully materialized gear loadout.
//
// Each gear field holds an index into the original []api.EquipmentItem the
// request was built from. Storing indices (rather than raw simc lines)
type profileset struct {
	// name identifies this profileset within the simc input and is the join
	// key against simc's json2 profilesets.results[].name.
	name string

	head     int
	neck     int
	shoulder int
	back     int
	chest    int
	wrist    int
	hands    int
	waist    int
	legs     int
	feet     int
	finger1  int
	finger2  int
	trinket1 int
	trinket2 int
	mainHand int
	offHand  int // may be noItemIndex

	talents string
}

// newProfileset returns a profileset with every gear slot set to noItemIndex
// so an accidentally‑unset slot is loudly wrong (index ‑1) rather than
// silently aliasing equipment[0].
func newProfileset() profileset {
	return profileset{
		head: noItemIndex, neck: noItemIndex, shoulder: noItemIndex,
		back: noItemIndex, chest: noItemIndex, wrist: noItemIndex,
		hands: noItemIndex, waist: noItemIndex, legs: noItemIndex,
		feet: noItemIndex, finger1: noItemIndex, finger2: noItemIndex,
		trinket1: noItemIndex, trinket2: noItemIndex,
		mainHand: noItemIndex, offHand: noItemIndex,
	}
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

// singletonPool represents one normal gear slot where we simply pick one item
// from the available choices.
type singletonPool struct {
	slot    string
	indices []int
	set     func(*profileset, int)
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

// buildTopGearCandidatePools normalizes the API equipment payload into the
// pools used by counting and generation.
//
// Dedup rules:
//   - Singleton slots: keep one entry per simIdentity. Two items that produce
//     identical simc assignments would only generate duplicate profilesets.
//   - Rings/trinkets: keep up to maxCopiesPerPairedItem entries per
//     simIdentity, so owning two stat‑identical rings still allows an (A, A)
//     pair. A third copy is rejected — it could never yield a new loadout.
func buildTopGearCandidatePools(
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

// countTopGearProfilesets computes the number of valid profilesets implied by
// the equipment payload without allocating the final profileset slice.
func countTopGearProfilesets(equipment []api.EquipmentItem) (int, error) {
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

// topGearManifest is an abstraction over a topgear sim job.
// it binds generated profilesets to the equipment table their
// slot indices reference. Indices are only meaningful against this exact
// slice, so the two are never exposed separately.
type topGearManifest struct {
	characterName  string // name of the character being simmed
	level          int    // level of the character being simmed
	race           string // race of the character being simmed
	characterClass api.CharacterClass
	spec           string

	equipment   []api.EquipmentItem // defensive copy of the request payload
	profilesets []profileset
}

// generateTopGearManifest expands the equipment pools into deterministic
// profilesets. The recursive singleton walk mirrors the counting logic so the
// generated order remains intuitive: earlier input candidates appear earlier in
// the resulting Combo1/Combo2/... sequence.
func generateTopGearManifest(
	opts api.SimulationOptionsTopGear,
) (topGearManifest, error) {
	pools, err := buildTopGearCandidatePools(opts.Equipment)
	if err != nil {
		return topGearManifest{}, err
	}

	count, err := countTopGearProfilesets(opts.Equipment)
	if err != nil {
		return topGearManifest{}, err
	}

	singletonPools := pools.singletonPools()
	ringPairs := makeUnorderedPairs(pools.rings)
	trinketPairs := makeUnorderedPairs(pools.trinkets)

	profilesets := make([]profileset, 0, count)
	base := newProfileset()
	base.talents = opts.TalentLoadout.Talents

	offHandOptions := pools.offHand
	if len(offHandOptions) == 0 {
		offHandOptions = []int{noItemIndex}
	}

	comboIndex := 1

	var buildSingletons func(poolIndex int, current profileset)
	buildSingletons = func(poolIndex int, current profileset) {
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

	eqCopy := make([]api.EquipmentItem, len(opts.Equipment))
	copy(eqCopy, opts.Equipment)

	return topGearManifest{
		profilesets:    profilesets,
		equipment:      eqCopy,
		characterName:  opts.CharacterName,
		level:          90,      /* TODO: assuming 90 for now, thread it through the api later */
		race:           "human", /*TODO: assuming human for now, thread through api */
		characterClass: opts.Class,
		spec:           opts.Spec,
	}, nil
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

func appendCompletedProfilesets(
	profilesets *[]profileset,
	current profileset,
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
		withRings.finger1 = rings.first
		withRings.finger2 = rings.second

		for _, trinkets := range trinketPairs {
			withTrinkets := withRings
			withTrinkets.trinket1 = trinkets.first
			withTrinkets.trinket2 = trinkets.second

			for _, offHand := range offHandOptions {
				complete := withTrinkets
				complete.offHand = offHand
				complete.name = fmt.Sprintf("Combo%d", nextComboIndex)
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

func setHead(p *profileset, i int)     { p.head = i }
func setNeck(p *profileset, i int)     { p.neck = i }
func setShoulder(p *profileset, i int) { p.shoulder = i }
func setBack(p *profileset, i int)     { p.back = i }
func setChest(p *profileset, i int)    { p.chest = i }
func setWrist(p *profileset, i int)    { p.wrist = i }
func setHands(p *profileset, i int)    { p.hands = i }
func setWaist(p *profileset, i int)    { p.waist = i }
func setLegs(p *profileset, i int)     { p.legs = i }
func setFeet(p *profileset, i int)     { p.feet = i }
func setMainHand(p *profileset, i int) { p.mainHand = i }

// SimcLines renders the manifest to a complete simc profile.
//
// We define the base profile, and then append the lines of all profilesets:
// `profileset."ComboN"+=…`
func (m *topGearManifest) SimcLines() ([]string, error) {
	if m.profilesets == nil {
		return nil, errManifestHasNilProfilesets
	}

	if len(m.profilesets) == 0 {
		return nil, errManifestHasNoProfilesets
	}

	var out []string

	// write the base profile: name, race and level
	// we need to take one of the profilesets and use it to "seed"
	// the base profile with a baseline set of equipment. We'll use the
	// first profileset for this purpose

	baseLines := []string{
		fmt.Sprintf(`%s="%s"`, m.characterClass, m.characterName),
		fmt.Sprintf(`level=%v`, m.level),
		fmt.Sprintf(`race=%s`, m.race),
		fmt.Sprintf(`spec=%s`, m.spec),
		"iterations=5", // for testing purposes
	}

	seedProfileset := m.profilesets[0]

	baseEquipment := []string{
		m.equipment[seedProfileset.head].RawLine,
		m.equipment[seedProfileset.neck].RawLine,
		m.equipment[seedProfileset.shoulder].RawLine,
		m.equipment[seedProfileset.back].RawLine,
		m.equipment[seedProfileset.chest].RawLine,
		m.equipment[seedProfileset.wrist].RawLine,
		m.equipment[seedProfileset.hands].RawLine,
		m.equipment[seedProfileset.waist].RawLine,
		m.equipment[seedProfileset.legs].RawLine,
		m.equipment[seedProfileset.feet].RawLine,
		retargetEquipmentLine(m.equipment[seedProfileset.finger1].RawLine, api.Finger1),
		retargetEquipmentLine(m.equipment[seedProfileset.finger2].RawLine, api.Finger2),
		retargetEquipmentLine(m.equipment[seedProfileset.trinket1].RawLine, api.Trinket1),
		retargetEquipmentLine(m.equipment[seedProfileset.trinket2].RawLine, api.Trinket2),
		m.equipment[seedProfileset.mainHand].RawLine,
	}
	if seedProfileset.offHand == noItemIndex {
		baseEquipment = append(baseEquipment, emptyOffHandLine)
	} else {
		baseEquipment = append(
			baseEquipment,
			retargetEquipmentLine(m.equipment[seedProfileset.offHand].RawLine, api.OffHand),
		)
	}

	out = append(out, baseLines...)
	out = append(out, baseEquipment...)

	// then, add all of the profileset lines!
	for i := range m.profilesets {
		out = append(out, m.profilesets[i].lines(m.equipment)...)
	}
	return out, nil
}

func (m *topGearManifest) Equipment() []api.EquipmentItem { return m.equipment }
func (m *topGearManifest) Len() int                       { return len(m.profilesets) }

// lines converts the loadout into the `profileset."Name"+=…` lines that will
// be appended to the simc profile. equipment must be the same slice that
// buildTopGearCandidatePools was called with.
func (l *profileset) lines(equipment []api.EquipmentItem) []string {
	type slotLine struct {
		idx      int
		retarget api.EquipmentSlot // "" => emit RawLine as‑is
	}

	slots := []slotLine{
		{l.head, ""}, {l.neck, ""}, {l.shoulder, ""}, {l.back, ""},
		{l.chest, ""}, {l.wrist, ""}, {l.hands, ""}, {l.waist, ""},
		{l.legs, ""}, {l.feet, ""},
		{l.finger1, api.Finger1}, {l.finger2, api.Finger2},
		{l.trinket1, api.Trinket1}, {l.trinket2, api.Trinket2},
		{l.mainHand, ""}, {l.offHand, api.OffHand},
	}

	const exLinesPerProfileset = 17 // 16 gear + 1 talents
	lines := make([]string, 0, exLinesPerProfileset)

	emit := func(raw string) {
		lines = append(lines, fmt.Sprintf(`profileset."%s"+=%s`, l.name, raw))
	}

	for _, s := range slots {
		switch {
		case s.idx == noItemIndex && s.retarget == api.OffHand:
			emit(emptyOffHandLine)
		case s.idx == noItemIndex:
			// Required slot left unset — this indicates a bug in generation.
			panic(fmt.Sprintf("profileset %q has unset required slot", l.name))
		case s.retarget != "":
			emit(retargetEquipmentLine(equipment[s.idx].RawLine, s.retarget))
		default:
			emit(equipment[s.idx].RawLine)
		}
	}

	emit("talents=" + l.talents)

	return lines
}
