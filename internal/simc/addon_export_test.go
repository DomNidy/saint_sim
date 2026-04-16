package simc

import (
	"reflect"
	"testing"
)

func TestParse(t *testing.T) {
	t.Parallel()

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
head=,id=250458,bonus_id=6652/12667/13577/13333/12787
# Gnarlroot Spinecleaver (250)
main_hand=,id=249671,enchant_id=3368,bonus_id=12786/6652

### Gear from Bags
#
# Frayed Guise (201)
# head=,id=258876,bonus_id=13611,drop_level=90
#
# Silvermoon Suncrest (246)
# head=,id=266432,bonus_id=13577/12785

### Additional Character Info
#
	# catalyst_currencies=3269:8/2813:8/3116:8

# Checksum: 6dda4018`

	got := Parse(input)

	want := AddonExport{
		class:         DeathKnight,
		level:         "90",
		race:          "maghar_orc",
		region:        "us",
		server:        "hydraxis",
		role:          "attack",
		professions:   "mining=34/",
		spec:          "unholy",
		activeTalents: "ACTIVE_TALENTS",
		alternateTalentLoadout: []alternateTalentLoadout{
			{name: "M+", talents: "MPLUS_TALENTS"},
			{name: "RAID", talents: "RAID_TALENTS"},
		},
		equipmentItems: []equipmentItem{
			{
				name:      "Host Commander's Casque (253)",
				equipment: "head=,id=250458,bonus_id=6652/12667/13577/13333/12787",
				bagItem:   false,
			},
			{
				name:      "Gnarlroot Spinecleaver (250)",
				equipment: "main_hand=,id=249671,enchant_id=3368,bonus_id=12786/6652",
				bagItem:   false,
			},
			{
				name:      "Frayed Guise (201)",
				equipment: "head=,id=258876,bonus_id=13611,drop_level=90",
				bagItem:   true,
			},
			{
				name:      "Silvermoon Suncrest (246)",
				equipment: "head=,id=266432,bonus_id=13577/12785",
				bagItem:   true,
			},
		},
		checksum: "6dda4018",
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Parse() mismatch:\n got: %#v\nwant: %#v", got, want)
	}
}

func TestParseDoesNotAcceptUnderscoreClassAliases(t *testing.T) {
	t.Parallel()

	got := Parse("death_knight=\"Example\"\nspec=unholy\n")

	if got.class != "" {
		t.Fatalf("class = %q, want empty class identifier", got.class)
	}
}
