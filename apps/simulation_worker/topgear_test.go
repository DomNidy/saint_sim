package main

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/DomNidy/saint_sim/internal/api"
	"github.com/DomNidy/saint_sim/internal/simc"
)

type stubRunner struct{}

func (sr stubRunner) Run(ctx context.Context, profilePath string) ([]byte, error) {
	return nil, nil
}
func TestProcessTopGear(t *testing.T) {
	worker := simulationWorker{
		runner: stubRunner{},
	}

	// Define simulation options
	topGearOptions := api.SimulationOptionsTopGear{
		Kind:          api.SimulationOptionsTopGearKind(api.SimulationOptionsKindTopGear),
		CharacterName: "Dom",
		Class:         string(simc.Warrior),
		Equipment: []api.AddonExportEquipmentItem{
			{},
		},
	}
	var opts api.SimulationOptions

	// convert the top gear type into the union type
	err := opts.FromSimulationOptionsTopGear(topGearOptions)
	if err != nil {
		t.Fatal(err)
	}

	// TODO: do some testing ...
	worker.processTopGear(context.Background(), simulationRequest{
		id:      uuid.New(),
		options: opts,
	})
}

func TestGenerateProfileSets(t *testing.T) {
	lines := []string{
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
		"finger1=,id=256971,gem_id=240865,bonus_id=12769/6652/13668",
		"finger2=,id=256985,gem_id=213491,bonus_id=12778/6652/13668",
		"trinket1=,id=250226,bonus_id=12785/13439/6652/12699",
		"trinket2=,id=251787,bonus_id=12786/6652",
		"head=,id=249952,enchant_id=7961,bonus_id=6652/12667/13440/13338/13575/12798",
		"neck=,id=249626,gem_id=240983,bonus_id=6652/13668/12794",
		"shoulder=,id=249950,enchant_id=8031,bonus_id=6652/13440/13340/13574/12798",
		"back=,id=257175,bonus_id=6652/13577/12790",
		"chest=,id=249955,enchant_id=7987,bonus_id=6652/13440/13336/13575/12798",
		"shirt=,id=3428",
		"tabard=,id=69210",
		"wrist=,id=249326,gem_id=240869,bonus_id=6652/13534/13577/13334/12794",
		"hands=,id=249953,bonus_id=13334/13337/6652/13574/12795",
		"waist=,id=249331,bonus_id=6652/12667/13577/13334/12795",
		"legs=,id=249951,enchant_id=8163,bonus_id=6652/12795/13440/13339/13575/3151",
		"feet=,id=251169,enchant_id=7963,bonus_id=13440/6652/13577/12699/12798",
		"finger1=,id=151308,enchant_id=7969,gem_id=240900,bonus_id=12795/13440/42/13668/12699",
		"finger2=,id=251115,enchant_id=7969,gem_id=240892,bonus_id=12795/13440/6652/13668/12699",
		"trinket1=,id=252420,bonus_id=12795/13440/42/12699",
		"trinket2=,id=249342,bonus_id=6652/13333/12790",
		"main_hand=,id=237846,enchant_id=8039,bonus_id=12214/12497/12066/12693/8960/8793/13622/13667,crafted_stats=36/49,crafting_quality=5",
		"off_hand=,id=251078,enchant_id=7983,bonus_id=13440/6652/12701/12798",
		"back=,id=258575,bonus_id=13439/6652/13577/12699/12786",
		"back=,id=235499,enchant_id=7403,gem_id=238046,bonus_id=12401/9893/12256",
		"back=,id=193712,bonus_id=12795/13440/6652/13577/12699",
		"chest=,id=193705,enchant_id=7957,bonus_id=13439/41/13577/12699/12790",
		"chest=,id=193705,bonus_id=12795/13440/6652/13577/12699",
		"wrist=,id=151328,bonus_id=13439/6652/12667/13577/12699/12790",
		"wrist=,id=249660,bonus_id=6652/12667/13577/12790",
		"hands=,id=251081,bonus_id=12795/13440/6652/13577/12699",
		"waist=,id=251086,bonus_id=12795/13440/6652/13534/13577/12699",
		"legs=,id=250457,bonus_id=41/13577/13333/12787",
		"feet=,id=237917,bonus_id=12249/12248/4785/12496/8960/12384/8793/13620,crafted_stats=49/32,crafting_quality=4",
		"finger1=,id=151311,enchant_id=8021,gem_id=240873,bonus_id=13439/6652/13668/12699/12790",
		"finger1=,id=251194,enchant_id=8021,gem_id=240885,bonus_id=13439/6652/13668/12699/12782",
		"finger1=,id=265814,bonus_id=13639",
		"finger1=,id=251115,enchant_id=8021,gem_id=240885,bonus_id=13439/6652/13668/12699/12782",
		"trinket1=,id=250227,bonus_id=12795/13440/6652/12699",
	}

	equipment := make([]api.AddonExportEquipmentItem, 0, len(lines))
	for _, line := range lines {
		item, ok := simc.ParseEquipmentItem("", line, api.Equipped)
		if !ok {
			t.Fatalf("failed to parse item line: %s", line)
		}
		equipment = append(equipment, item)
	}

	_ = equipment
}
