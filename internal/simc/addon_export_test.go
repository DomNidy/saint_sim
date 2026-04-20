package simc

import (
	"reflect"
	"sort"
	"testing"

	api "github.com/DomNidy/saint_sim/internal/api"
)

func TestParse(t *testing.T) {
	t.Parallel()

	talentLoadouts := []api.AddonExportTalentLoadout{
		{Name: strPtr("Active"), Talents: "ACTIVE_TALENTS"},
		{Name: strPtr("M+"), Talents: "MPLUS_TALENTS"},
		{Name: strPtr("RAID"), Talents: "RAID_TALENTS"},
	}
	equipment := []api.EquipmentItem{
		{
			Fingerprint: fingerprintForItem(
				"head=,id=250458,bonus_id=6652/12667/13577/13333/12787,crafted_stats=40/49,crafting_quality=5",
				"equipped",
			),
			Slot:            "head",
			Name:            "Host Commander's Casque",
			DisplayName:     "Host Commander's Casque",
			ItemId:          250458,
			ItemLevel:       intPtr(253),
			CraftingQuality: intPtr(5),
			BonusIds:        intSlicePtr([]int{6652, 12667, 13577, 13333, 12787}),
			CraftedStats:    intSlicePtr([]int{40, 49}),
			Source:          api.Equipped,
			RawLine:         "head=,id=250458,bonus_id=6652/12667/13577/13333/12787,crafted_stats=40/49,crafting_quality=5",
			EnchantId:       nil,
			GemIds:          nil,
		},
		{
			Fingerprint: fingerprintForItem(
				"main_hand=,id=249671,enchant_id=3368,bonus_id=12786/6652",
				"equipped",
			),
			Slot:            "main_hand",
			Name:            "Gnarlroot Spinecleaver",
			DisplayName:     "Gnarlroot Spinecleaver",
			ItemId:          249671,
			ItemLevel:       intPtr(250),
			EnchantId:       intPtr(3368),
			BonusIds:        intSlicePtr([]int{12786, 6652}),
			GemIds:          nil,
			CraftedStats:    nil,
			CraftingQuality: nil,
			Source:          api.Equipped,
			RawLine:         "main_hand=,id=249671,enchant_id=3368,bonus_id=12786/6652",
		},
		{
			Fingerprint: fingerprintForItem(
				"head=,id=258876,bonus_id=13611,drop_level=90",
				"bag",
			),
			Slot:            "head",
			Name:            "Frayed Guise",
			DisplayName:     "Frayed Guise",
			ItemId:          258876,
			ItemLevel:       intPtr(201),
			EnchantId:       nil,
			CraftingQuality: nil,
			BonusIds:        intSlicePtr([]int{13611}),
			GemIds:          nil,
			CraftedStats:    nil,
			Source:          api.Bag,
			RawLine:         "head=,id=258876,bonus_id=13611,drop_level=90",
		},
		{
			Fingerprint: fingerprintForItem(
				"head=,id=266432,bonus_id=13577/12785,gem_id1=213482,gem_id2=213743",
				"bag",
			),
			Slot:            "head",
			Name:            "Silvermoon Suncrest",
			DisplayName:     "Silvermoon Suncrest",
			ItemId:          266432,
			ItemLevel:       intPtr(246),
			EnchantId:       nil,
			CraftingQuality: nil,
			BonusIds:        intSlicePtr([]int{13577, 12785}),
			GemIds:          intSlicePtr([]int{213482, 213743}),
			CraftedStats:    nil,
			Source:          api.Bag,
			RawLine:         "head=,id=266432,bonus_id=13577/12785,gem_id1=213482,gem_id2=213743",
		},
	}
	catalystCurrencies := map[string]int{
		"3269": 8,
		"2813": 8,
		"3116": 8,
	}
	slotHighWatermarks := map[string]api.AddonExportSlotHighWatermark{
		"head": {
			CurrentItemLevel: 639,
			MaxItemLevel:     652,
		},
		"off_hand": {
			CurrentItemLevel: 636,
			MaxItemLevel:     649,
		},
	}
	upgradeAchievements := []int{123, 456}

	const input = `# Gubulgi - Unholy - 2026-03-28 12:47 - US/Hydraxis
# SimC Addon 12.0.0-02
# WoW 12.0.1.66709, TOC 120001
# Requires SimulationCraft 1000-01 or newer

deathknight="Gubulgi"
level=90
race=maghar_orc
region=us
server=hydraxis
role=attack
professions=mining=34/
spec=unholy
# loot_spec=unholy

talents=ACTIVE_TALENTS

# Saved Loadout: M+
# talents=MPLUS_TALENTS
# Saved Loadout: RAID
# talents=RAID_TALENTS

# Host Commander's Casque (253)
head=,id=250458,bonus_id=6652/12667/13577/13333/12787,crafted_stats=40/49,crafting_quality=5
# Gnarlroot Spinecleaver (250)
main_hand=,id=249671,enchant_id=3368,bonus_id=12786/6652

### Gear from Bags
#
# Frayed Guise (201)
# head=,id=258876,bonus_id=13611,drop_level=90
#
# Silvermoon Suncrest (246)
# head=,id=266432,bonus_id=13577/12785,gem_id1=213482,gem_id2=213743

### Additional Character Info
#
	# catalyst_currencies=3269:8/2813:8/3116:8
	# slot_high_watermarks=head:639:652/off_hand:636:649
	# upgrade_achievements=123/456

# Checksum: 6dda4018`

	got := Parse(input)

	want := api.AddonExport{
		CharacterName:       strPtr("Gubulgi"),
		Class:               classPtr(api.Deathknight),
		Level:               strPtr("90"),
		Race:                strPtr("maghar_orc"),
		Region:              strPtr("us"),
		Server:              strPtr("hydraxis"),
		Role:                strPtr("attack"),
		Professions:         strPtr("mining=34/"),
		Spec:                strPtr("unholy"),
		TalentLoadouts:      &talentLoadouts,
		HeaderComment:       strPtr("Gubulgi - Unholy - 2026-03-28 12:47 - US/Hydraxis"),
		SimcAddonComment:    strPtr("SimC Addon 12.0.0-02"),
		WowBuildComment:     strPtr("WoW 12.0.1.66709, TOC 120001"),
		RequiredSimcComment: strPtr("Requires SimulationCraft 1000-01 or newer"),
		LootSpec:            strPtr("unholy"),
		Checksum:            strPtr("6dda4018"),
		Equipment:           &equipment,
		CatalystCurrencies:  &catalystCurrencies,
		SlotHighWatermarks:  &slotHighWatermarks,
		UpgradeAchievements: &upgradeAchievements,
	}

	sortTalentLoadouts(got.TalentLoadouts)
	sortTalentLoadouts(want.TalentLoadouts)

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Parse() mismatch:\n got: %#v\nwant: %#v", got, want)
	}
}

func TestParseDoesNotAcceptUnderscoreClassAliases(t *testing.T) {
	t.Parallel()

	got := Parse("death_knight=\"Example\"\nspec=unholy\n")

	if got.Class != nil {
		t.Fatalf("class = %q, want nil class identifier", *got.Class)
	}
}

func TestHasRecognizedData(t *testing.T) {
	t.Parallel()

	if HasRecognizedData(Parse("### comments only")) {
		t.Fatal("HasRecognizedData returned true for unparseable input")
	}

	if !HasRecognizedData(Parse("priest=\"Example\"\nspec=shadow\n")) {
		t.Fatal("HasRecognizedData returned false for recognizable input")
	}
}

func intPtr(value int) *int {
	return &value
}

func classPtr(value api.CharacterClass) *api.CharacterClass {
	return &value
}

// Parse does not guarantee loadout ordering, so normalize before comparing.
func sortTalentLoadouts(loadouts *[]api.AddonExportTalentLoadout) {
	if loadouts == nil {
		return
	}

	sort.Slice(*loadouts, func(i, j int) bool {
		left := (*loadouts)[i]
		right := (*loadouts)[j]
		leftName := ""
		if left.Name != nil {
			leftName = *left.Name
		}
		rightName := ""
		if right.Name != nil {
			rightName = *right.Name
		}

		if leftName != rightName {
			return leftName < rightName
		}

		return left.Talents < right.Talents
	})
}
