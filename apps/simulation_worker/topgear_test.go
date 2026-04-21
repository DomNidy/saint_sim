package main

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/DomNidy/saint_sim/internal/api"
)

type stubRunner struct{}

func (sr stubRunner) Run(_ context.Context, profilePath string) ([]byte, error) {
	_ = profilePath

	return nil, nil
}

func TestGenerateTopGearProfilesetsDeterministic(t *testing.T) {
	t.Parallel()

	equipment := appendEquipment(
		mustParseEquippedEquipmentLines(t, baseSingletonLines()...),
		mustParseEquippedEquipmentLines(
			t,
			extraHeadLine(),
			baseRingLines()[0],
			baseRingLines()[1],
			"finger1=,id=151311,enchant_id=8021,gem_id=240873,bonus_id=13439/6652/13668/12699/12790",
		)...,
	)
	equipment = appendEquipment(
		equipment,
		mustParseEquippedEquipmentLines(t, baseTrinketLines()...)...,
	)

	gotCount, err := countTopGearProfilesets(equipment)
	if err != nil {
		t.Fatalf("countTopGearProfilesets() error = %v", err)
	}
	if gotCount != 6 {
		t.Fatalf("countTopGearProfilesets() = %d, want 6", gotCount)
	}

	got, err := generateTopGearManifest(equipment, "TALENTS")
	if err != nil {
		t.Fatalf("generateTopGearProfilesets() error = %v", err)
	}
	if len(got) != gotCount {
		t.Fatalf("len(generateTopGearProfilesets()) = %d, want %d", len(got), gotCount)
	}

	want := []struct {
		name    string
		head    string
		finger1 string
		finger2 string
		offHand string
		talents string
	}{
		{
			name:    "Combo1",
			head:    baseSingletonLines()[0],
			finger1: baseRingLines()[0],
			finger2: baseRingLines()[1],
			offHand: "off_hand=,",
			talents: "talents=TALENTS",
		},
		{
			name:    "Combo2",
			head:    baseSingletonLines()[0],
			finger1: baseRingLines()[0],
			finger2: "finger2=,id=151311,enchant_id=8021,gem_id=240873,bonus_id=13439/6652/13668/12699/12790",
			offHand: "off_hand=,",
			talents: "talents=TALENTS",
		},
		{
			name:    "Combo3",
			head:    baseSingletonLines()[0],
			finger1: "finger1=,id=256985,gem_id=213491,bonus_id=12778/6652/13668",
			finger2: "finger2=,id=151311,enchant_id=8021,gem_id=240873,bonus_id=13439/6652/13668/12699/12790",
			offHand: "off_hand=,",
			talents: "talents=TALENTS",
		},
		{
			name:    "Combo4",
			head:    extraHeadLine(),
			finger1: baseRingLines()[0],
			finger2: baseRingLines()[1],
			offHand: "off_hand=,",
			talents: "talents=TALENTS",
		},
		{
			name:    "Combo5",
			head:    extraHeadLine(),
			finger1: baseRingLines()[0],
			finger2: "finger2=,id=151311,enchant_id=8021,gem_id=240873,bonus_id=13439/6652/13668/12699/12790",
			offHand: "off_hand=,",
			talents: "talents=TALENTS",
		},
		{
			name:    "Combo6",
			head:    extraHeadLine(),
			finger1: "finger1=,id=256985,gem_id=213491,bonus_id=12778/6652/13668",
			finger2: "finger2=,id=151311,enchant_id=8021,gem_id=240873,bonus_id=13439/6652/13668/12699/12790",
			offHand: "off_hand=,",
			talents: "talents=TALENTS",
		},
	}

	type profilesetSummary struct {
		name    string
		head    string
		finger1 string
		finger2 string
		offHand string
		talents string
	}

	gotSummaries := make([]profilesetSummary, 0, len(got))
	for _, profile := range got {
		gotSummaries = append(gotSummaries, profilesetSummary{
			name:    profile.name,
			head:    profile.head,
			finger1: profile.finger1,
			finger2: profile.finger2,
			offHand: profile.offHand,
			talents: profile.talents,
		})
	}

	wantSummaries := make([]profilesetSummary, 0, len(want))
	for _, expected := range want {
		wantSummaries = append(wantSummaries, profilesetSummary(expected))
	}

	if !reflect.DeepEqual(gotSummaries, wantSummaries) {
		t.Fatalf("generateTopGearProfilesets() = %#v, want %#v", gotSummaries, wantSummaries)
	}
}

func TestGenerateTopGearProfilesetsAllowsDuplicateRingFromDifferentSources(t *testing.T) {
	t.Parallel()

	equipment := appendEquipment(
		mustParseEquippedEquipmentLines(t, baseSingletonLines()...),
		mustParseEquippedEquipmentLines(t, baseTrinketLines()...)...,
	)
	equipment = appendEquipment(
		equipment,
		mustParseEquipmentLine(t, api.Equipped, baseRingLines()[0]),
		mustParseEquipmentLine(
			t,
			api.Bag,
			"finger2=,id=256971,gem_id=240865,bonus_id=12769/6652/13668",
		),
	)

	got, err := generateTopGearManifest(equipment, "TALENTS")
	if err != nil {
		t.Fatalf("generateTopGearProfilesets() error = %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("len(generateTopGearProfilesets()) = %d, want 1", len(got))
	}
	if got[0].finger1 != baseRingLines()[0] ||
		got[0].finger2 != "finger2=,id=256971,gem_id=240865,bonus_id=12769/6652/13668" {
		t.Fatalf(
			"ring pair = (%q, %q), want the two distinct copies of the same ring",
			got[0].finger1,
			got[0].finger2,
		)
	}
}

func TestCountTopGearProfilesetsRejectsImpossibleLoadout(t *testing.T) {
	t.Parallel()

	equipment := appendEquipment(
		mustParseEquippedEquipmentLines(t, baseSingletonLines()...),
		mustParseEquippedEquipmentLines(t, baseTrinketLines()...)...,
	)
	equipment = appendEquipment(
		equipment,
		mustParseEquipmentLine(t, api.Equipped, baseRingLines()[0]),
		mustParseEquipmentLine(t, api.Equipped, baseRingLines()[0]),
	)

	count, err := countTopGearProfilesets(equipment)
	if err == nil {
		t.Fatalf("countTopGearProfilesets() = %d, want error", count)
	}
	if !strings.Contains(err.Error(), "ring candidates") {
		t.Fatalf("countTopGearProfilesets() error = %v, want ring-candidate error", err)
	}
}

func TestProcessTopGear(t *testing.T) {
	t.Parallel()

	worker := simulationWorker{
		runner: stubRunner{},
		store:  nil,
	}

	equipment := appendEquipment(
		mustParseEquippedEquipmentLines(t, baseSingletonLines()...),
		mustParseEquippedEquipmentLines(t, baseRingLines()...)...,
	)
	equipment = appendEquipment(
		equipment,
		mustParseEquippedEquipmentLines(t, baseTrinketLines()...)...)

	opts, err := topGearOptions(equipment...)
	if err != nil {
		t.Fatal(err)
	}

	err = worker.processTopGear(t.Context(), simulationRequest{
		id:      uuid.New(),
		options: opts,
	})
	if err != nil {
		t.Fatalf("processTopGear() error = %v", err)
	}
}

func TestProcessTopGearReturnsErrorWhenCombinationCountExceedsLimit(t *testing.T) {
	t.Parallel()

	worker := simulationWorker{
		runner: stubRunner{},
		store:  nil,
	}

	headLines := make([]string, 0, maxGeneratedProfilesets+1)
	for headIndex := range maxGeneratedProfilesets + 1 {
		headLines = append(
			headLines,
			"head=,id="+intString(300000+headIndex)+",bonus_id=6652/12667/13577/13333/12787",
		)
	}

	equipment := appendEquipment(
		mustParseEquippedEquipmentLines(t, headLines...),
		mustParseEquippedEquipmentLines(t, baseSingletonLines()[1:]...)...,
	)
	equipment = appendEquipment(
		equipment,
		mustParseEquippedEquipmentLines(t, baseRingLines()...)...)
	equipment = appendEquipment(
		equipment,
		mustParseEquippedEquipmentLines(t, baseTrinketLines()...)...)

	opts, err := topGearOptions(equipment...)
	if err != nil {
		t.Fatal(err)
	}

	err = worker.processTopGear(t.Context(), simulationRequest{
		id:      uuid.New(),
		options: opts,
	})
	if err == nil {
		t.Fatal("processTopGear() error = nil, want max-combination error")
	}
	if !strings.Contains(err.Error(), "exceeds max") {
		t.Fatalf("processTopGear() error = %v, want max-combination error", err)
	}
}

func mustParseEquippedEquipmentLines(
	t *testing.T,
	lines ...string,
) []api.EquipmentItem {
	t.Helper()

	items := make([]api.EquipmentItem, 0, len(lines))
	for _, line := range lines {
		items = append(items, mustParseEquipmentLine(t, api.Equipped, line))
	}

	return items
}

func mustParseEquipmentLine(
	t *testing.T,
	source api.AddonExportEquipmentSource,
	line string,
) api.EquipmentItem {
	t.Helper()

	item, ok := ParseEquipmentItem("", line, source)
	if !ok {
		t.Fatalf("failed to parse item line %q", line)
	}

	return item
}

func appendEquipment(
	base []api.EquipmentItem,
	items ...api.EquipmentItem,
) []api.EquipmentItem {
	return append(base, items...)
}

func topGearOptions(equipment ...api.EquipmentItem) (api.SimulationOptions, error) {
	topGear := api.SimulationOptionsTopGear{
		Kind:          api.TopGear,
		CharacterName: "Dom",
		Class:         api.Deathknight,
		Spec:          "unholy",
		Role:          "attack",
		Equipment:     equipment,
		TalentLoadout: api.AddonExportTalentLoadout{
			Name:    nil,
			Talents: "CwPAAAAAAAAAAAAAAAAAAAAAA",
		},
	}

	var opts api.SimulationOptions
	if err := opts.FromSimulationOptionsTopGear(topGear); err != nil {
		return api.SimulationOptions{}, fmt.Errorf("encode top gear options: %w", err)
	}

	return opts, nil
}

func intString(value int) string {
	return strconv.Itoa(value)
}

func baseSingletonLines() []string {
	return []string{
		"head=,id=250458,bonus_id=6652/12667/13577/13333/12787",
		"neck=,id=249626,gem_id=213494,bonus_id=12793/6652/13668",
		"shoulder=,id=249968,bonus_id=13574/13340/6652/13574/12793",
		"back=,id=257021,bonus_id=12779/6652/13577",
		"chest=,id=249653,bonus_id=12786/6652/13577",
		"wrist=,id=256965,bonus_id=6652/12667/13578/12772",
		"hands=,id=249655,bonus_id=12786/6652/13577",
		"waist=,id=249659,bonus_id=6652/12667/13578/12770",
		"legs=,id=257213,bonus_id=13634",
		"feet=,id=256973,bonus_id=43/13578/12771",
		"main_hand=,id=249671,enchant_id=3368,bonus_id=12786/6652",
	}
}

func baseRingLines() []string {
	return []string{
		"finger1=,id=256971,gem_id=240865,bonus_id=12769/6652/13668",
		"finger2=,id=256985,gem_id=213491,bonus_id=12778/6652/13668",
	}
}

func baseTrinketLines() []string {
	return []string{
		"trinket1=,id=250226,bonus_id=12785/13439/6652/12699",
		"trinket2=,id=251787,bonus_id=12786/6652",
	}
}

func extraHeadLine() string {
	return "head=,id=249952,enchant_id=7961,bonus_id=6652/12667/13440/13338/13575/12798"
}
