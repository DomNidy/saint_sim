package sims

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/DomNidy/saint_sim/internal/api"
)

func TestGenerateTopGearProfilesetsDeterministic(t *testing.T) {

}

func TestGenerateTopGearProfilesetsAllowsDuplicateRingFromDifferentSources(t *testing.T) {

}

func TestCountProfilesetsAllowsDuplicateRingFromDifferentSources(t *testing.T) {

}

func TestCountProfilesetsCanReportLargeCombinationCount(t *testing.T) {

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
			EquippedItems:       api.CharacterEquippedItems{},
			Level:               80,
			Race:                "human",
			Role:                stringPtr("attack"),
			Spec:                "unholy",
			ActiveTalents:       api.CharacterTalentLoadout{Talents: talents},
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

func offHandSummaryLine(plan TopGearSimPlan, offHandIndex int) string {
	if offHandIndex == noItemIndex {
		return emptyOffHandLine
	}

	return retargetEquipmentLine(plan.equipment[offHandIndex].RawLine, api.OffHand)
}

func countProfilesets(equipment []api.EquipmentItem) (int, error) {
	plan := TopGearSimPlan{
		equipment: equipment,
	}

	return plan.CountProfilesets()
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
