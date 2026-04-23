//nolint:exhaustruct
package api_utils

import (
	"testing"

	api "github.com/DomNidy/saint_sim/internal/api"
)

func TestParseAddonExportToWowCharacter(t *testing.T) {
	t.Parallel()

	got := ParseAddonExport(`priest="Example"
level=80
race=void_elf
region=us
server=area-52
role=spell
professions=alchemy=100/herbalism=100
spec=shadow
talents=ACTIVE_TALENTS
# loot_spec=shadow
# Saved Loadout: Dungeon
# talents=DUNGEON_TALENTS
# Host Commander's Casque (684)
head=,id=250458,bonus_id=6652/12667,gem_id=240865,ilevel=684
# Gear from Bags
# Gnarlroot Spinecleaver (710)
main_hand=,id=249671,bonus_id=6652,ilevel=710
# Additional Character Info
# catalyst_currencies=2912:1/3116:2
# slot_high_watermarks=head:684:710/main_hand:710:720
# upgrade_achievements=1/2`)

	if got.CharacterClass != api.Priest {
		t.Fatalf("character_class = %q, want %q", got.CharacterClass, api.Priest)
	}
	if got.Level != 80 {
		t.Fatalf("level = %d, want 80", got.Level)
	}
	if got.Spec != "shadow" {
		t.Fatalf("spec = %q, want shadow", got.Spec)
	}
	if got.LootSpec == nil || *got.LootSpec != "shadow" {
		t.Fatalf("loot_spec = %v, want shadow", got.LootSpec)
	}
	if got.ActiveTalents == nil || got.ActiveTalents.Talents != "ACTIVE_TALENTS" {
		t.Fatalf("active_talents = %#v, want ACTIVE_TALENTS", got.ActiveTalents)
	}
	if got.TalentLoadouts == nil || len(*got.TalentLoadouts) != 1 {
		t.Fatalf("talent_loadouts = %#v, want one non-active loadout", got.TalentLoadouts)
	}
	if (*got.TalentLoadouts)[0].Name == nil || *(*got.TalentLoadouts)[0].Name != "Dungeon" {
		t.Fatalf("talent loadout name = %#v, want Dungeon", (*got.TalentLoadouts)[0].Name)
	}
	if (*got.TalentLoadouts)[0].Talents != "DUNGEON_TALENTS" {
		t.Fatalf(
			"talent loadout talents = %q, want DUNGEON_TALENTS",
			(*got.TalentLoadouts)[0].Talents,
		)
	}
	if len(got.EquippedItems) != 1 {
		t.Fatalf("equipped_items len = %d, want 1", len(got.EquippedItems))
	}
	if got.EquippedItems[0].Source != api.Equipped {
		t.Fatalf("equipped item source = %q, want %q", got.EquippedItems[0].Source, api.Equipped)
	}
	if got.BagItems == nil || len(*got.BagItems) != 1 {
		t.Fatalf("bag_items = %#v, want one bag item", got.BagItems)
	}
	if (*got.BagItems)[0].Source != api.Bag {
		t.Fatalf("bag item source = %q, want %q", (*got.BagItems)[0].Source, api.Bag)
	}
	if got.CatalystCurrencies == nil || len(*got.CatalystCurrencies) != 2 {
		t.Fatalf("catalyst_currencies = %#v, want two entries", got.CatalystCurrencies)
	}
	if (*got.CatalystCurrencies)[0].Id != 2912 || (*got.CatalystCurrencies)[0].Quantity != 1 {
		t.Fatalf("first catalyst currency = %#v, want 2912:1", (*got.CatalystCurrencies)[0])
	}
	if got.SlotHighWatermarks == nil || len(*got.SlotHighWatermarks) != 2 {
		t.Fatalf("slot_high_watermarks = %#v, want two entries", got.SlotHighWatermarks)
	}
	if (*got.SlotHighWatermarks)[0].Slot != api.Head {
		t.Fatalf(
			"first slot watermark slot = %q, want %q",
			(*got.SlotHighWatermarks)[0].Slot,
			api.Head,
		)
	}
	if got.UpgradeAchievements == nil || len(*got.UpgradeAchievements) != 2 {
		t.Fatalf("upgrade_achievements = %#v, want two entries", got.UpgradeAchievements)
	}
}

func TestNormalizeLineEndings(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "already normalized",
			input: "line1\nline2\nline3",
			want:  "line1\nline2\nline3",
		},
		{
			name:  "windows line endings",
			input: "line1\r\nline2\r\nline3",
			want:  "line1\nline2\nline3",
		},
		{
			name:  "classic mac line endings",
			input: "line1\rline2\rline3",
			want:  "line1\nline2\nline3",
		},
		{
			name:  "mixed line endings",
			input: "line1\r\nline2\rline3\nline4",
			want:  "line1\nline2\nline3\nline4",
		},
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
		{
			name:  "single trailing carriage return",
			input: "line1\r",
			want:  "line1\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := NormalizeLineEndings(tt.input)
			if got != tt.want {
				t.Fatalf("NormalizeLineEndings(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestStripAllComments(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "Removes only commented lines",
			input: "# Simc version 1.23\nwarrior=John\nlevel=90\n",
			want:  "warrior=John\nlevel=90\n",
		},
		{
			name:  "Only commented input returns empty str",
			input: "# This is a comment line1.\n#Comment line 2\n#asd\n",
			want:  "",
		},
		{
			name:  "Multiple consecutive #'s",
			input: "##Comment #Two\n## Hey\nwarrior=John\n###123",
			want:  "warrior=John",
		},
		{
			name:  "Preserves blank lines between non-comment lines",
			input: "# Simc version 1.23\n\nwarrior=John\n\nlevel=90\n# trailing comment",
			want:  "\nwarrior=John\n\nlevel=90",
		},
		{
			name:  "Preserves spacing on non-comment lines",
			input: "# comment\n  warrior=John  \n\tlevel=90\t\n",
			want:  "  warrior=John  \n\tlevel=90\t\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := StripAllComments(tt.input)
			if got != tt.want {
				t.Fatalf("StripAllComments(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestTrimLineWhitespace(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "Trims leading and trailing spaces on each line",
			input: "  warrior=John  \n  level=90  ",
			want:  "warrior=John\nlevel=90",
		},
		{
			name:  "Preserves blank lines while trimming spaces",
			input: "  warrior=John  \n   \n  level=90  ",
			want:  "warrior=John\n\nlevel=90",
		},
		{
			name:  "Normalizes line endings before trimming",
			input: "  warrior=John  \r\n  level=90  \r",
			want:  "warrior=John\nlevel=90\n",
		},
		{
			name:  "Trims tabs at the start and end of each line",
			input: "\twarrior=John\t\n \tlevel=90\t ",
			want:  "warrior=John\nlevel=90",
		},
		{
			name:  "Preserves internal tabs",
			input: "value\t=\t42\n\tname\t=\tvalue\t",
			want:  "value\t=\t42\nname\t=\tvalue",
		},
		{
			name:  "Empty string",
			input: "",
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := TrimLineWhitespace(tt.input)
			if got != tt.want {
				t.Fatalf("TrimLineWhitespace(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
