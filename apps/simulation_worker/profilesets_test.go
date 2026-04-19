package main

import (
	"slices"
	"testing"
)

func TestProfileset(t *testing.T) {
	t.Parallel()

	loadout := profileset{
		name:     "Combo1",
		head:     "head=,id=250458,bonus_id=6652/12667/13577/13333/12787",
		neck:     "neck=,id=249626,gem_id=213494,bonus_id=12793/6652/13668",
		shoulder: "shoulder=,id=249968,bonus_id=13574/13340/6652/13574/12793",
		back:     "back=,id=257021,bonus_id=12779/6652/13577",
		chest:    "chest=,id=249653,bonus_id=12786/6652/13577",
		wrist:    "wrist=,id=256965,bonus_id=6652/12667/13578/12772",
		hands:    "hands=,id=249655,bonus_id=12786/6652/13577",
		waist:    "waist=,id=249659,bonus_id=6652/12667/13578/12770",
		legs:     "legs=,id=257213,bonus_id=13634",
		feet:     "feet=,id=256973,bonus_id=43/13578/12771",
		finger1:  "finger1=,id=256971,gem_id=240865,bonus_id=12769/6652/13668",
		finger2:  "finger2=,id=256985,gem_id=213491,bonus_id=12778/6652/13668",
		trinket1: "trinket1=,id=250226,bonus_id=12785/13439/6652/12699",
		trinket2: "trinket2=,id=251787,bonus_id=12786/6652",
		mainHand: "main_hand=,id=249671,enchant_id=3368,bonus_id=12786/6652",
		offHand:  "off_hand=,",
		talents:  "talents=CwPAAAAAAAAAAAAAAAAAAAAAAAYmhZmZmZYWmZmZaYmxMzYAAAAAAAAYmBzAglhZmtZmxwAsYWMMkBmNGasgBMDAzMzMMAzMYGD",
	}

	want := []string{
		`profileset."Combo1"+=head=,id=250458,bonus_id=6652/12667/13577/13333/12787`,
		`profileset."Combo1"+=neck=,id=249626,gem_id=213494,bonus_id=12793/6652/13668`,
		`profileset."Combo1"+=shoulder=,id=249968,bonus_id=13574/13340/6652/13574/12793`,
		`profileset."Combo1"+=back=,id=257021,bonus_id=12779/6652/13577`,
		`profileset."Combo1"+=chest=,id=249653,bonus_id=12786/6652/13577`,
		`profileset."Combo1"+=wrist=,id=256965,bonus_id=6652/12667/13578/12772`,
		`profileset."Combo1"+=hands=,id=249655,bonus_id=12786/6652/13577`,
		`profileset."Combo1"+=waist=,id=249659,bonus_id=6652/12667/13578/12770`,
		`profileset."Combo1"+=legs=,id=257213,bonus_id=13634`,
		`profileset."Combo1"+=feet=,id=256973,bonus_id=43/13578/12771`,
		`profileset."Combo1"+=finger1=,id=256971,gem_id=240865,bonus_id=12769/6652/13668`,
		`profileset."Combo1"+=finger2=,id=256985,gem_id=213491,bonus_id=12778/6652/13668`,
		`profileset."Combo1"+=trinket1=,id=250226,bonus_id=12785/13439/6652/12699`,
		`profileset."Combo1"+=trinket2=,id=251787,bonus_id=12786/6652`,
		`profileset."Combo1"+=main_hand=,id=249671,enchant_id=3368,bonus_id=12786/6652`,
		`profileset."Combo1"+=off_hand=,`,
		`profileset."Combo1"+=talents=CwPAAAAAAAAAAAAAAAAAAAAAAAYmhZmZmZYWmZmZaYmxMzYAAAAAAAAYmBzAglhZmtZmxwAsYWMMkBmNGasgBMDAzMzMMAzMYGD`,
	}

	profilesetLines := loadout.Profileset()
	if !slices.Equal(profilesetLines, want) {
		t.Fatalf("Profileset() = %v, want %v", profilesetLines, want)
	}
}
