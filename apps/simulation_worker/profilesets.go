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
// request was built from. Storing indices (rather than raw simc lines) means
// the same struct serves two purposes:
//   - Lines() turns it into the `profileset."ComboN"+=…` text fed to simc.
//   - SlotIndices() is the compact per‑combo manifest returned to the client,
//     so the result payload references items instead of duplicating them.
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

// Lines converts the loadout into the `profileset."Name"+=…` lines that will
// be appended to the simc profile. equipment must be the same slice that
// buildTopGearCandidatePools was called with.
func (l *profileset) Lines(equipment []api.EquipmentItem) []string {
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

// SlotIndices returns the equipment‑array index chosen for each slot in this
// loadout, omitting empty slots. This is the per‑profileset payload for the
// top‑gear result contract (`{head: int, neck: int, …}`).
func (l *profileset) SlotIndices() map[api.EquipmentSlot]int {
	out := map[api.EquipmentSlot]int{
		api.Head:     l.head,
		api.Neck:     l.neck,
		api.Shoulder: l.shoulder,
		api.Back:     l.back,
		api.Chest:    l.chest,
		api.Wrist:    l.wrist,
		api.Hands:    l.hands,
		api.Waist:    l.waist,
		api.Legs:     l.legs,
		api.Feet:     l.feet,
		api.Finger1:  l.finger1,
		api.Finger2:  l.finger2,
		api.Trinket1: l.trinket1,
		api.Trinket2: l.trinket2,
		api.MainHand: l.mainHand,
	}
	if l.offHand != noItemIndex {
		out[api.OffHand] = l.offHand
	}
	return out
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

// generateTopGearProfilesets expands the equipment pools into deterministic
// profilesets. The recursive singleton walk mirrors the counting logic so the
// generated order remains intuitive: earlier input candidates appear earlier in
// the resulting Combo1/Combo2/... sequence.
func generateTopGearProfilesets(
	equipment []api.EquipmentItem,
	talents string,
) ([]profileset, error) {
	pools, err := buildTopGearCandidatePools(equipment)
	if err != nil {
		return nil, err
	}

	count, err := countTopGearProfilesets(equipment)
	if err != nil {
		return nil, err
	}

	singletonPools := pools.singletonPools()
	ringPairs := makeUnorderedPairs(pools.rings)
	trinketPairs := makeUnorderedPairs(pools.trinkets)

	profilesets := make([]profileset, 0, count)
	base := newProfileset()
	base.talents = talents

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

	return profilesets, nil
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
