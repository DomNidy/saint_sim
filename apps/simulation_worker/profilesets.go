// Package main hosts the simulation worker executable.
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

var (
	errTopGearMissingRawLine       = errors.New("top gear equipment item missing raw_line")
	errTopGearUnsupportedSlot      = errors.New("unsupported top gear slot")
	errTopGearMissingRequiredSlot  = errors.New("missing required top gear slot")
	errTopGearInsufficientRings    = errors.New("need at least two distinct ring candidates")
	errTopGearInsufficientTrinkets = errors.New("need at least two distinct trinket candidates")
	errTopGearMissingEquipment     = errors.New("missing top gear equipment")
	errTopGearProfilesetLimit      = errors.New("top gear profileset count exceeds max")
)

// profileset is a fully materialized SimulationCraft profileset stanza.
//
// Each field stores the raw assignment line that should be appended to the
// generated `profileset."Name"+=...` output.
type profileset struct {
	// a name to identify this profileset.
	// should be unique w.r.t other profileset,
	// names in the simulation. (If a simc file
	// has multiple profilesets w/ the same name,
	// not what we want)
	name     string
	head     string
	neck     string
	shoulder string
	back     string
	chest    string
	wrist    string
	hands    string
	waist    string
	legs     string
	feet     string
	finger1  string
	finger2  string
	trinket1 string
	trinket2 string
	mainHand string
	offHand  string
	talents  string
}

// Profileset converts the loadout combination into the corresponding
// lines for the profileset. These lines will be written to the simc profile.
func (l *profileset) Profileset() []string {
	fields := []string{
		l.head,
		l.neck,
		l.shoulder,
		l.back,
		l.chest,
		l.wrist,
		l.hands,
		l.waist,
		l.legs,
		l.feet,
		l.finger1,
		l.finger2,
		l.trinket1,
		l.trinket2,
		l.mainHand,
		l.offHand,
		l.talents,
	}

	const exLinesPerProfileset = 17 // 17 lines per profileset (16 for gear, 1 for talent)

	lines := make([]string, 0, exLinesPerProfileset)

	profilesetLine := func(line string) string {
		return fmt.Sprintf(`profileset."%s"+=%s`, l.name, line)
	}

	for _, field := range fields {
		lines = append(lines, profilesetLine(field))
	}

	return lines
}

// topGearCandidatePools is the organized set of gear choices we can build
// profilesets from.
type topGearCandidatePools struct {
	head     []string
	neck     []string
	shoulder []string
	back     []string
	chest    []string
	wrist    []string
	hands    []string
	waist    []string
	legs     []string
	feet     []string
	rings    []string
	trinkets []string
	mainHand []string
	offHand  []string
}

// singletonPool represents one normal gear slot where we simply pick one item
// from the available choices.
type singletonPool struct {
	slot  string
	lines []string
	set   func(*profileset, string)
}

// slotPair is one chosen pair of interchangeable items, like two rings or two
// trinkets.
type slotPair struct {
	first  string
	second string
}

// buildTopGearCandidatePools normalizes the API equipment payload into the
// pools used by counting and generation.
//
// The normalization rules are intentionally different for singleton slots vs
// ring/trinket pools:
//   - singleton slots are deduplicated by raw line, because picking two copies
//     of the exact same simc assignment would only create duplicate profilesets
//   - rings/trinkets are deduplicated by item instance identity so two copies of
//     the same item can still be paired when they come from different sources
func buildTopGearCandidatePools(
	equipment []api.EquipmentItem,
) (topGearCandidatePools, error) {
	var pools topGearCandidatePools

	singletonDestinations := map[api.EquipmentSlot]*[]string{
		api.Head: &pools.head, api.Neck: &pools.neck, api.Shoulder: &pools.shoulder,
		api.Back: &pools.back, api.Chest: &pools.chest, api.Wrist: &pools.wrist,
		api.Hands: &pools.hands, api.Waist: &pools.waist, api.Legs: &pools.legs,
		api.Feet: &pools.feet, api.MainHand: &pools.mainHand, api.OffHand: &pools.offHand,
	}

	// Singleton slots: dedup by raw simc line - identical assignments would
	// only produce duplicate profilesets.
	singletonSeen := map[api.EquipmentSlot]*set.Set[string]{}
	appendSingleton := func(item api.EquipmentItem, dest *[]string) {
		seen, ok := singletonSeen[item.Slot]
		if !ok {
			seen = set.New[string]()
			singletonSeen[item.Slot] = seen
		}
		if seen.Add(item.RawLine) {
			*dest = append(*dest, item.RawLine)
		}
	}

	// Paired slots: dedup by *instance identity* so two copies of the same
	// item from different sources can still be worn together.
	instanceIdentity := func(item api.EquipmentItem) string {
		if item.Fingerprint != "" {
			return item.Fingerprint
		}
		return item.RawLine + "|" + string(item.Source)
	}
	ringsSeen, trinketsSeen := set.New[string](), set.New[string]()
	appendInstance := func(seen *set.Set[string], item api.EquipmentItem, dest *[]string) {
		if seen.Add(instanceIdentity(item)) {
			*dest = append(*dest, item.RawLine)
		}
	}

	for _, item := range equipment {
		if item.RawLine == "" {
			return topGearCandidatePools{}, fmt.Errorf(
				"%w: %q",
				errTopGearMissingRawLine,
				item.Slot,
			)
		}

		switch {
		case isRingSlot(item.Slot):
			appendInstance(ringsSeen, item, &pools.rings)
		case isTrinketSlot(item.Slot):
			appendInstance(trinketsSeen, item, &pools.trinkets)
		case isIgnoredTopGearSlot(item.Slot):
			// cosmetic — not part of top-gear generation
		default:
			dest, ok := singletonDestinations[item.Slot]
			if !ok {
				return topGearCandidatePools{}, fmt.Errorf(
					"%w: %q",
					errTopGearUnsupportedSlot,
					item.Slot,
				)
			}
			appendSingleton(item, dest)
		}
	}

	return pools, nil
}

func (p topGearCandidatePools) singletonPools() []singletonPool {
	return []singletonPool{
		{slot: "head", lines: p.head, set: setHead},
		{slot: "neck", lines: p.neck, set: setNeck},
		{slot: "shoulder", lines: p.shoulder, set: setShoulder},
		{slot: "back", lines: p.back, set: setBack},
		{slot: "chest", lines: p.chest, set: setChest},
		{slot: "wrist", lines: p.wrist, set: setWrist},
		{slot: "hands", lines: p.hands, set: setHands},
		{slot: "waist", lines: p.waist, set: setWaist},
		{slot: "legs", lines: p.legs, set: setLegs},
		{slot: "feet", lines: p.feet, set: setFeet},
		{slot: "main_hand", lines: p.mainHand, set: setMainHand},
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
		if len(pool.lines) == 0 {
			return 0, fmt.Errorf("%w: %q", errTopGearMissingRequiredSlot, pool.slot)
		}

		total *= len(pool.lines)
	}

	if len(pools.rings) < minPairedItems {
		return 0, fmt.Errorf("%w: got %d", errTopGearInsufficientRings, len(pools.rings))
	}
	total *= unorderedPairCount(len(pools.rings))

	if len(pools.trinkets) < minPairedItems {
		return 0, fmt.Errorf("%w: got %d", errTopGearInsufficientTrinkets, len(pools.trinkets))
	}
	total *= unorderedPairCount(len(pools.trinkets))

	// off_hand is intentionally optional for two-handed weapon setups.
	if len(pools.offHand) == 0 {
		total *= 1
	} else {
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
	var base profileset
	base.talents = "talents=" + talents

	comboIndex := 1
	offHandOptions := pools.offHand
	if len(offHandOptions) == 0 {
		offHandOptions = []string{"off_hand=,"}
	}

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
		for _, line := range pool.lines {
			next := current
			pool.set(&next, line)
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

func makeUnorderedPairs(lines []string) []slotPair {
	pairs := make([]slotPair, 0, unorderedPairCount(len(lines)))
	for leftIndex := range lines {
		for rightIndex := leftIndex + 1; rightIndex < len(lines); rightIndex++ {
			pairs = append(pairs, slotPair{
				first:  lines[leftIndex],
				second: lines[rightIndex],
			})
		}
	}

	return pairs
}

func retargetEquipmentLine(line string, slot api.EquipmentSlot) string {
	_, rest, foundAssignment := strings.Cut(line, "=")
	if !foundAssignment {
		return line
	}

	return string(slot) + "=" + rest
}

func appendCompletedProfilesets(
	profilesets *[]profileset,
	current profileset,
	ringPairs []slotPair,
	trinketPairs []slotPair,
	offHandOptions []string,
	startingIndex int,
) int {
	nextComboIndex := startingIndex

	// Rings and trinkets are pooled separately because their slots are
	// interchangeable. We only emit unordered pairs, so A/B appears once instead
	// of again as the mirror swap.
	for _, rings := range ringPairs {
		withRings := current
		withRings.finger1 = retargetEquipmentLine(rings.first, api.Finger1)
		withRings.finger2 = retargetEquipmentLine(rings.second, api.Finger2)

		for _, trinkets := range trinketPairs {
			withTrinkets := withRings
			withTrinkets.trinket1 = retargetEquipmentLine(trinkets.first, api.Trinket1)
			withTrinkets.trinket2 = retargetEquipmentLine(trinkets.second, api.Trinket2)

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

func setHead(profile *profileset, line string) {
	profile.head = line
}

func setNeck(profile *profileset, line string) {
	profile.neck = line
}

func setShoulder(profile *profileset, line string) {
	profile.shoulder = line
}

func setBack(profile *profileset, line string) {
	profile.back = line
}

func setChest(profile *profileset, line string) {
	profile.chest = line
}

func setWrist(profile *profileset, line string) {
	profile.wrist = line
}

func setHands(profile *profileset, line string) {
	profile.hands = line
}

func setWaist(profile *profileset, line string) {
	profile.waist = line
}

func setLegs(profile *profileset, line string) {
	profile.legs = line
}

func setFeet(profile *profileset, line string) {
	profile.feet = line
}

func setMainHand(profile *profileset, line string) {
	profile.mainHand = line
}
