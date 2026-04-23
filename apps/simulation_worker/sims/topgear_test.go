package sims

import (
	"fmt"
	"reflect"
	"strconv"
	"testing"

	api_utils "github.com/DomNidy/saint_sim/apps/api/api_utils"
	"github.com/DomNidy/saint_sim/internal/api"
)

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

	gotCount, err := countProfilesets(equipment)
	if err != nil {
		t.Fatalf("CountProfilesets() error = %v", err)
	}
	if gotCount != 6 {
		t.Fatalf("CountProfilesets() = %d, want 6", gotCount)
	}

	got, err := NewTopGearManifest(topGearConfig(equipment, "TALENTS"))
	if err != nil {
		t.Fatalf("generateTopGearProfilesets() error = %v", err)
	}
	if got.Len() != gotCount {
		t.Fatalf(
			"len(generateTopGearProfilesets()) = %d, want %d",
			got.Len(),
			gotCount,
		)
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
			talents: "TALENTS",
		},
		{
			name:    "Combo2",
			head:    baseSingletonLines()[0],
			finger1: baseRingLines()[0],
			finger2: "finger2=,id=151311,enchant_id=8021,gem_id=240873,bonus_id=13439/6652/13668/12699/12790",
			offHand: "off_hand=,",
			talents: "TALENTS",
		},
		{
			name:    "Combo3",
			head:    baseSingletonLines()[0],
			finger1: "finger1=,id=256985,gem_id=213491,bonus_id=12778/6652/13668",
			finger2: "finger2=,id=151311,enchant_id=8021,gem_id=240873,bonus_id=13439/6652/13668/12699/12790",
			offHand: "off_hand=,",
			talents: "TALENTS",
		},
		{
			name:    "Combo4",
			head:    extraHeadLine(),
			finger1: baseRingLines()[0],
			finger2: baseRingLines()[1],
			offHand: "off_hand=,",
			talents: "TALENTS",
		},
		{
			name:    "Combo5",
			head:    extraHeadLine(),
			finger1: baseRingLines()[0],
			finger2: "finger2=,id=151311,enchant_id=8021,gem_id=240873,bonus_id=13439/6652/13668/12699/12790",
			offHand: "off_hand=,",
			talents: "TALENTS",
		},
		{
			name:    "Combo6",
			head:    extraHeadLine(),
			finger1: "finger1=,id=256985,gem_id=213491,bonus_id=12778/6652/13668",
			finger2: "finger2=,id=151311,enchant_id=8021,gem_id=240873,bonus_id=13439/6652/13668/12699/12790",
			offHand: "off_hand=,",
			talents: "TALENTS",
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

	gotSummaries := make([]profilesetSummary, 0, got.Len())
	for _, profile := range got.Profilesets {
		gotSummaries = append(gotSummaries, profilesetSummary{
			name:    profile.Name,
			head:    got.equipment[profile.Head].RawLine,
			finger1: retargetEquipmentLine(got.equipment[profile.Finger1].RawLine, api.Finger1),
			finger2: retargetEquipmentLine(got.equipment[profile.Finger2].RawLine, api.Finger2),
			offHand: offHandSummaryLine(got, profile.OffHand),
			talents: profile.Talents,
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

	got, err := NewTopGearManifest(topGearConfig(equipment, "TALENTS"))
	if err != nil {
		t.Fatalf("generateTopGearProfilesets() error = %v", err)
	}
	if got.Len() != 1 {
		t.Fatalf("len(generateTopGearProfilesets()) = %d, want 1", got.Len())
	}
	if retargetEquipmentLine(
		got.equipment[got.Profilesets[0].Finger1].RawLine,
		api.Finger1,
	) != baseRingLines()[0] ||
		retargetEquipmentLine(
			got.equipment[got.Profilesets[0].Finger2].RawLine,
			api.Finger2,
		) != "finger2=,id=256971,gem_id=240865,bonus_id=12769/6652/13668" {
		t.Fatalf(
			"ring pair = (%q, %q), want the two distinct copies of the same ring",
			retargetEquipmentLine(got.equipment[got.Profilesets[0].Finger1].RawLine, api.Finger1),
			retargetEquipmentLine(got.equipment[got.Profilesets[0].Finger2].RawLine, api.Finger2),
		)
	}
}

func TestCountProfilesetsAllowsDuplicateRingFromDifferentSources(t *testing.T) {
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

	count, err := countProfilesets(equipment)
	if err != nil {
		t.Fatalf("CountProfilesets() error = %v", err)
	}
	if count != 1 {
		t.Fatalf("CountProfilesets() = %d, want 1", count)
	}
}

func TestCountProfilesetsCanReportLargeCombinationCount(t *testing.T) {
	t.Parallel()

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

	count, err := countProfilesets(equipment)
	if err != nil {
		t.Fatalf("countProfilesets() error = %v", err)
	}
	if count <= maxGeneratedProfilesets {
		t.Fatalf("countProfilesets() = %d, want more than %d", count, maxGeneratedProfilesets)
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
	source api.EquipmentSource,
	line string,
) api.EquipmentItem {
	t.Helper()

	item, ok := api_utils.ParseEquipmentItem("", line, source)
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

func topGearConfig(
	equipment []api.EquipmentItem,
	talents string,
) api.SimulationConfigTopGear {
	name := "Dom"

	return api.SimulationConfigTopGear{
		Kind:       api.SimulationConfigTopGearKindTopGear,
		CoreConfig: api.SimulationCoreConfig{},
		Equipment:  equipment,
		Character: api.WowCharacter{
			Name:                &name,
			CharacterClass:      api.Deathknight,
			EquippedItems:       []api.EquipmentItem{},
			Level:               80,
			Race:                "human",
			Role:                stringPtr("attack"),
			Spec:                "unholy",
			ActiveTalents:       &api.CharacterTalentLoadout{Talents: talents},
			TalentLoadouts:      nil,
			BagItems:            nil,
			LootSpec:            nil,
			Professions:         nil,
			Region:              nil,
			Server:              nil,
			CatalystCurrencies:  nil,
			SlotHighWatermarks:  nil,
			UpgradeAchievements: nil,
		},
	}
}

func topGearOptions(equipment ...api.EquipmentItem) (api.SimulationOptions, error) {
	topGear := topGearConfig(equipment, "CwPAAAAAAAAAAAAAAAAAAAAAA")

	var opts api.SimulationOptions
	if err := opts.FromSimulationConfigTopGear(topGear); err != nil {
		return api.SimulationOptions{}, fmt.Errorf("encode top gear options: %w", err)
	}

	return opts, nil
}

func intString(value int) string {
	return strconv.Itoa(value)
}

func offHandSummaryLine(manifest TopGearManifest, offHandIndex int) string {
	if offHandIndex == noItemIndex {
		return emptyOffHandLine
	}

	return retargetEquipmentLine(manifest.equipment[offHandIndex].RawLine, api.OffHand)
}

func countProfilesets(equipment []api.EquipmentItem) (int, error) {
	manifest := TopGearManifest{
		equipment: equipment,
	}

	return manifest.CountProfilesets()
}

func stringPtr(value string) *string {
	return &value
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
