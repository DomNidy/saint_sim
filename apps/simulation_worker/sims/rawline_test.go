package sims //nolint:testpackage // Tests cover unexported rawline sanitizer helpers.

import (
	"errors"
	"strings"
	"testing"

	"github.com/DomNidy/saint_sim/internal/api"
)

const unsafeProxyValue = "bad\nproxy=evil"

func TestEquipmentRawlineDoesNotEmitUntrustedStringFields(t *testing.T) {
	t.Parallel()

	got, err := equipmentRawline(testEquipmentItem(api.Head, 250015, func(item *api.EquipmentItem) {
		item.Name = `bad",proxy=evil`
		item.DisplayName = unsafeProxyValue
		item.Source = api.Equipped
		item.RawLine = "head=,id=250015\nproxy=evil"
	}))
	if err != nil {
		t.Fatal(err)
	}

	const want = "head=,id=250015"
	if got != want {
		t.Fatalf("equipmentRawline() = %q, want %q", got, want)
	}

	for _, dangerous := range []string{"proxy", "\n", "\r", `"`, "evil"} {
		if strings.Contains(got, dangerous) {
			t.Fatalf("equipmentRawline() emitted dangerous content %q in %q", dangerous, got)
		}
	}
}

func TestEquipmentRawlineRejectsDisallowedSlots(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		slot api.EquipmentSlot
	}{
		{name: "unknown slot", slot: api.EquipmentSlot("unknown")},
		{name: "proxy sim option", slot: api.EquipmentSlot("proxy")},
		{name: "shirt cosmetic slot", slot: api.Shirt},
		{name: "tabard cosmetic slot", slot: api.Tabard},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := equipmentRawline(testEquipmentItem(tt.slot, 1))
			if !errors.Is(err, errInvalidEquipmentSlot) {
				t.Fatalf("equipmentRawline() error = %v, want %v", err, errInvalidEquipmentSlot)
			}
		})
	}
}

func TestEquipmentRawlineFormatsFullCraftedJewelryLine(t *testing.T) {
	t.Parallel()

	got, err := equipmentRawline(testEquipmentItem(api.Neck, 240950, func(item *api.EquipmentItem) {
		item.GemIds = intSlicePointer([]int{240909})
		item.BonusIds = intSlicePointer(
			[]int{12214, 12497, 12066, 13454, 8960, 8795, 13622, 13667},
		)
		item.CraftedStats = intSlicePointer([]int{36, 49})
		item.CraftingQuality = intPointer(5)
	}))
	if err != nil {
		t.Fatal(err)
	}

	const want = "neck=,id=240950,gem_id=240909,bonus_id=12214/12497/12066/13454/8960/8795/13622/13667,crafted_stats=36/49,crafting_quality=5"
	if got != want {
		t.Fatalf("equipmentRawline() = %q, want %q", got, want)
	}
}

func TestEquipmentRawlineFormatsMinimalItem(t *testing.T) {
	t.Parallel()

	got, err := equipmentRawline(testEquipmentItem(api.OffHand, 3419))
	if err != nil {
		t.Fatal(err)
	}

	const want = "off_hand=,id=3419"
	if got != want {
		t.Fatalf("equipmentRawline() = %q, want %q", got, want)
	}
}

func TestEquipmentRawlineOmitsNilAndEmptySlices(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		item api.EquipmentItem
	}{
		{
			name: "nil slices",
			item: testEquipmentItem(api.Head, 250015),
		},
		{
			name: "empty slices",
			item: testEquipmentItem(api.Head, 250015, func(item *api.EquipmentItem) {
				item.GemIds = intSlicePointer([]int{})
				item.BonusIds = intSlicePointer([]int{})
				item.CraftedStats = intSlicePointer([]int{})
			}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := equipmentRawline(tt.item)
			if err != nil {
				t.Fatal(err)
			}

			const want = "head=,id=250015"
			if got != want {
				t.Fatalf("equipmentRawline() = %q, want %q", got, want)
			}
		})
	}
}

func TestCharacterClassRawline(t *testing.T) {
	t.Parallel()

	got, err := characterClassRawline(api.Monk)
	if err != nil {
		t.Fatal(err)
	}

	if got != "monk" {
		t.Fatalf("characterClassRawline() = %q, want %q", got, "monk")
	}
}

func TestCharacterClassRawlineRejectsUnsafeValues(t *testing.T) {
	t.Parallel()

	for _, class := range []api.CharacterClass{
		api.CharacterClass("proxy"),
		api.CharacterClass("monk\nproxy=evil"),
	} {
		t.Run(string(class), func(t *testing.T) {
			t.Parallel()

			_, err := characterClassRawline(class)
			if !errors.Is(err, errInvalidCharacterClass) {
				t.Fatalf(
					"characterClassRawline() error = %v, want %v",
					err,
					errInvalidCharacterClass,
				)
			}
		})
	}
}

func TestProfileFieldRawlinesRejectUnsafeValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		run  func() (string, error)
		want error
	}{
		{
			name: "character name quote injection",
			run: func() (string, error) {
				name := `bad",proxy=evil`

				return characterNameRawline(api.Monk, &name)
			},
			want: errInvalidProfileName,
		},
		{
			name: "character name newline injection",
			run: func() (string, error) {
				name := unsafeProxyValue

				return characterNameRawline(api.Monk, &name)
			},
			want: errInvalidProfileName,
		},
		{
			name: "invalid level",
			run: func() (string, error) {
				return levelRawline(0)
			},
			want: errInvalidLevel,
		},
		{
			name: "unsafe race",
			run: func() (string, error) {
				return identifierRawline("race", "void_elf\nproxy=evil")
			},
			want: errInvalidIdentifier,
		},
		{
			name: "unsafe spec",
			run: func() (string, error) {
				return identifierRawline("spec", "brewmaster,proxy=evil")
			},
			want: errInvalidIdentifier,
		},
		{
			name: "unsafe talents",
			run: func() (string, error) {
				return talentsRawline("abc\nproxy=evil")
			},
			want: errInvalidTalents,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			_, err := testCase.run()
			if !errors.Is(err, testCase.want) {
				t.Fatalf("rawline helper error = %v, want %v", err, testCase.want)
			}
		})
	}
}

func TestBasicSimProfileRejectsUnsafeBaseFields(t *testing.T) {
	t.Parallel()

	name := unsafeProxyValue
	manifest := BasicSimManifest{
		simConfig: api.SimulationConfigBasic{
			Character: api.WowCharacter{
				ActiveTalents:       api.CharacterTalentLoadout{Name: nil, Talents: "abc"},
				BagItems:            nil,
				CatalystCurrencies:  nil,
				CharacterClass:      api.Monk,
				EquippedItems:       testCharacterEquippedItems(),
				Level:               90,
				LootSpec:            nil,
				Name:                &name,
				Professions:         nil,
				Race:                "nightborne",
				Region:              nil,
				Role:                nil,
				Server:              nil,
				SlotHighWatermarks:  nil,
				Spec:                "brewmaster",
				TalentLoadouts:      nil,
				UpgradeAchievements: nil,
			},
			CoreConfig: api.SimulationCoreConfig{FightStyle: nil},
			Kind:       api.SimulationConfigBasicKindBasic,
		},
	}

	_, err := manifest.buildSimcProfile()
	if !errors.Is(err, errInvalidProfileName) {
		t.Fatalf("buildSimcProfile() error = %v, want %v", err, errInvalidProfileName)
	}
}

func TestTopGearSimcLinesRejectsUnsafeTalents(t *testing.T) {
	t.Parallel()

	manifest := testTopGearManifest()
	manifest.Profilesets[0].Talents = "abc\nproxy=evil"

	_, err := manifest.SimcLines()
	if !errors.Is(err, errInvalidTalents) {
		t.Fatalf("SimcLines() error = %v, want %v", err, errInvalidTalents)
	}
}

func TestTopGearSimcLinesDoesNotEmitEquipmentRawLine(t *testing.T) {
	t.Parallel()

	manifest := testTopGearManifest()

	lines, err := manifest.SimcLines()
	if err != nil {
		t.Fatal(err)
	}

	for _, line := range lines {
		if strings.Contains(line, "proxy") || strings.ContainsAny(line, "\r") {
			t.Fatalf("SimcLines() emitted unsafe line %q", line)
		}
	}
}

func testTopGearManifest() TopGearManifest {
	equipment, profileset := testTopGearEquipmentAndProfileset()

	return TopGearManifest{
		characterName:  "Celinka",
		level:          90,
		race:           "nightborne",
		characterClass: api.Monk,
		spec:           "brewmaster",
		equipment:      equipment,
		Profilesets:    []Profileset{profileset},
	}
}

func testTopGearEquipmentAndProfileset() ([]api.EquipmentItem, Profileset) {
	slots := []api.EquipmentSlot{
		api.Head, api.Neck, api.Shoulder, api.Back, api.Chest, api.Wrist,
		api.Hands, api.Waist, api.Legs, api.Feet, api.Finger1, api.Finger2,
		api.Trinket1, api.Trinket2, api.MainHand, api.OffHand,
	}

	equipment := make([]api.EquipmentItem, 0, len(slots))
	for idx, slot := range slots {
		itemID := 250000 + idx
		equipment = append(
			equipment,
			testEquipmentItem(slot, itemID, func(item *api.EquipmentItem) {
				item.RawLine = string(slot) + "=,id=1\nproxy=evil"
			}),
		)
	}

	profileset := Profileset{
		Name:     "Combo1",
		Head:     0,
		Neck:     1,
		Shoulder: 2,
		Back:     3,
		Chest:    4,
		Wrist:    5,
		Hands:    6,
		Waist:    7,
		Legs:     8,
		Feet:     9,
		Finger1:  10,
		Finger2:  11,
		Trinket1: 12,
		Trinket2: 13,
		MainHand: 14,
		OffHand:  15,
		Talents:  "abc",
	}

	return equipment, profileset
}

func testEquipmentItem(
	slot api.EquipmentSlot,
	itemID int,
	options ...func(*api.EquipmentItem),
) api.EquipmentItem {
	item := api.EquipmentItem{
		BonusIds:        nil,
		CraftedStats:    nil,
		CraftingQuality: nil,
		DisplayName:     "",
		EnchantId:       nil,
		GemIds:          nil,
		ItemId:          itemID,
		ItemLevel:       nil,
		Name:            "",
		RawLine:         "",
		Slot:            slot,
		Source:          api.Equipped,
	}

	for _, option := range options {
		option(&item)
	}

	return item
}

func testCharacterEquippedItems() api.CharacterEquippedItems {
	return api.CharacterEquippedItems{
		Back:     testEquipmentItem(api.Back, 1),
		Chest:    testEquipmentItem(api.Chest, 2),
		Feet:     testEquipmentItem(api.Feet, 3),
		Finger1:  testEquipmentItem(api.Finger1, 4),
		Finger2:  testEquipmentItem(api.Finger2, 5),
		Hands:    testEquipmentItem(api.Hands, 6),
		Head:     testEquipmentItem(api.Head, 7),
		Legs:     testEquipmentItem(api.Legs, 8),
		MainHand: testEquipmentItem(api.MainHand, 193723),
		Neck:     testEquipmentItem(api.Neck, 9),
		OffHand:  nil,
		Shoulder: testEquipmentItem(api.Shoulder, 10),
		Trinket1: testEquipmentItem(api.Trinket1, 11),
		Trinket2: testEquipmentItem(api.Trinket2, 12),
		Waist:    testEquipmentItem(api.Waist, 13),
		Wrist:    testEquipmentItem(api.Wrist, 14),
	}
}

func intPointer(value int) *int {
	return &value
}

func intSlicePointer(values []int) *[]int {
	return &values
}
