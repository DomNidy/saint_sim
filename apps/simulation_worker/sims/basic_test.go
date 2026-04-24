package sims

import (
	"log"
	"testing"

	"github.com/DomNidy/saint_sim/internal/api"
)

const basicProfile string = `# Gubulgi - Unholy - 2026-04-23 10:52 - US/Hydraxis
# SimC Addon 12.0.0-02
# WoW 12.0.5.67088, TOC 120005
# Requires SimulationCraft 1000-01 or newer

deathknight="Gubulgi"
level=90
race=maghar_orc
region=us
server=hydraxis
role=attack
professions=mining=39/
spec=unholy
# loot_spec=unholy

talents=CwPAkXBWxkyfx9CbGaHonEAhLBYmhZmZmZYWmZmZaYmxMzYAAAAAAAAYAGAsMYmtZmxYAWMLGGyAzGDNWwAmBAGDAMzgB

# Saved Loadout: M+
# talents=CwPAkXBWxkyfx9CbGaHonEAhLBYmhZmZmZYWmZmZMMzYmZMAAAAAAAAMzgZAwywMz2MzYMALmFDDMwsxgxCGwMAMzMzwAMzgxA
# Saved Loadout: AoE
# talents=CwPAkXBWxkyfx9CbGaHonEAhLBYmZMjZmZY2mZmZMMzYmZMAAAAAAAAMzwMDAWGmZ2mZGzMALmFDDMwsxgxCGwAgxMzwAMzMMG
# Saved Loadout: RAID
# talents=CwPAkXBWxkyfx9CbGaHonEAhLBwMjZMDDz2MzMjZzMzMmxAAAAAAAAwMPAzMAYZGzMbzMjZmBsZWMMwAzGDGLAYGAYmZMDwMzMGD

# Host Commander's Casque (253)
head=,id=250458,bonus_id=6652/12667/13577/13333/12787
# Nocturnal Thorncharm (259)
neck=,id=249626,gem_id=213494,bonus_id=12793/6652/13668
# Relentless Rider's Dreadthorns (259)
shoulder=,id=249968,bonus_id=13340/6652/13574/12793
# Verdant Tracker's Cover (240)
back=,id=257021,bonus_id=12779/6652/13577
# Rampant Brambleplate (250)
chest=,id=249653,bonus_id=12786/6652/13577
# Steelbark Vambraces (230)
wrist=,id=256965,bonus_id=6652/12667/13578/12772
# Rampant Creepers (250)
hands=,id=249655,bonus_id=12786/6652/13577
# Rampant Thornstrap (224)
waist=,id=249659,bonus_id=6652/12667/13578/12770
# Voidbreaker's Faulds (214)
legs=,id=257213,bonus_id=13634
# Preyseeker's Polished Greatboots (233)
feet=,id=259942,bonus_id=12777/6652/13577
# Radiant Phoenix Band (220)
finger1=,id=256971,gem_id=240865,bonus_id=12769/6652/13668
# Forest Hunter's Hoop (237)
finger2=,id=256985,gem_id=213491,bonus_id=12778/6652/13668
# Latch's Crooked Hook (246)
trinket1=,id=250226,bonus_id=12785/13439/6652/12699
# Sealed Chaos Urn (250)
trinket2=,id=251787,bonus_id=12786/6652
# Gnarlroot Spinecleaver (250)
main_hand=,id=249671,enchant_id=3368,bonus_id=12786/6652

### Gear from Bags
#
# Frayed Guise (201)
# head=,id=258876,bonus_id=13611,drop_level=90
#
# Silvermoon Suncrest (246)
# head=,id=266432,bonus_id=13577/12785/12667
#
# Steelbark Casque (233)
# head=,id=256994,bonus_id=12777/6652/12667/13577
#
# Zedling Summoning Collar (220)
# head=,id=264536,bonus_id=12769/6652/13534/13578/11311
#
# Tarnished Dawnlit Pendant (198)
# neck=,id=258911,bonus_id=13730/41/13668,drop_level=90
#
# Pendant of Siphoned Vitality (220)
# neck=,id=264611,bonus_id=12769/6652/13668/11311
#
# Loa-Blessed Beads (224)
# neck=,id=256970,gem_id=240903,bonus_id=12770/6652/13668
#
# Preyseeker's Polished Pauldrons (220)
# shoulder=,id=259946,bonus_id=12769/6652/13578
#
# Steelbark Shoulderguards (233)
# shoulder=,id=256998,bonus_id=6652/13578/12773
#
# Rampant Thornmantles (227)
# shoulder=,id=249658,bonus_id=12771/6652/13578
#
# Voidbreaker's Shoulderplates (214)
# shoulder=,id=257182,bonus_id=13634
#
# Preyseeker's Refined Shawl (233)
# back=,id=259909,bonus_id=12777/42/13577
#
# Deepvine Shroud (230)
# back=,id=257022,bonus_id=12772/6652/13578
#
# Steelbark Cloak (220)
# back=,id=259360,bonus_id=12769/6652/13578
#
# Netherscale Cloak (220)
# back=,id=264594,bonus_id=12769/6652/13578/11311
#
# Lynxhide Shawl (220)
# back=,id=264595,bonus_id=12769/6652/13578/11311
#
# Sentinel Challenger's Prize (214)
# chest=,id=251151,bonus_id=13442/13437/41/13578
#
# Rampant Brambleplate (220)
# chest=,id=249653,bonus_id=12769/6652/13578
#
# Preyseeker's Polished Brigandine (233)
# chest=,id=259941,bonus_id=12777/6652/13577
#
# Steelbark Gauntlets (224)
# hands=,id=256977,bonus_id=12770/6652/13578
#
# Steelbark Gauntlets (224)
# hands=,id=256977,bonus_id=12770/6652/13578
#
# Riphook Defender (207)
# waist=,id=251086,bonus_id=13441/13436/13578,drop_level=90
#
# Steelbark Girdle (230)
# waist=,id=257017,bonus_id=12772/6652/12667/13578
#
# Devouring Outrider's Chausses (253)
# legs=,id=250457,bonus_id=6652/13578/13333/12787/13577
#
# Snapdragon Pantaloons (220)
# legs=,id=264543,bonus_id=12769/6652/13578/11311
#
# Steelbark Sabatons (227)
# feet=,id=256973,bonus_id=43/13578/12771
#
# Snapper Steppers (220)
# feet=,id=264585,bonus_id=12769/6652/13578/11311
#
# Vibrant Wilderloop (171)
# finger1=,id=249620,bonus_id=13648/6652/13668,drop_level=86
#
# Omission of Light (198)
# finger1=,id=251093,gem_id=240865,bonus_id=13441/13436/13668,drop_level=89
#
# Preyseeker's Ring (220)
# finger1=,id=259913,gem_id=213500,bonus_id=12769/6652/13668
#
# Evertwisting Swiftvine (220)
# finger1=,id=256972,bonus_id=12769/6652/13668
#
# Glorious Crusader's Keepsake (220)
# trinket1=,id=251792,bonus_id=12769/6652
#
# Lost Idol of the Hash'ey (224)
# trinket1=,id=251783,bonus_id=12770/6652
#
# Lost Idol of the Hash'ey (220)
# trinket1=,id=251783,bonus_id=6652/12769
#
# Preyseeker's Falchion (220)
# main_hand=,id=259964,bonus_id=12769/6652
#
# Vine-Rending Claymore (220)
# main_hand=,id=257006,bonus_id=12769/6652
#
# Resinous Blossomblade (237)
# main_hand=,id=249610,bonus_id=12778/6652/11215
#
# Tarnished Dawnlit Longsword (198)
# main_hand=,id=258952,bonus_id=13730/6652,drop_level=90
#
# Preyseeker's Falchion (233)
# main_hand=,id=259964,bonus_id=12777/6652
#
# Sharpened Borer Claw (220)
# main_hand=,id=264640,bonus_id=12769/6652/11311
#
# Strangely Eelastic Blade (220)
# main_hand=,id=264618,bonus_id=12769/6652/11311
#
# Coralfang's Hefty Fin (220)
# main_hand=,id=264629,bonus_id=12769/6652/11311
#
# Underbrush-Clearing Cleaver (237)
# main_hand=,id=259358,enchant_id=3368,bonus_id=6652/12774

### Additional Character Info
#
# catalyst_currencies=3269:8/2813:8/3116:8
#
# upgrade_currencies=
#
# slot_high_watermarks=0:253:253/1:259:259/2:259:259/3:250:250/4:230:230/5:214:214/6:233:233/7:230:230/8:250:250/9:220:220/10:246:246/11:240:240/12:250:250/13:10:74/14:133:133/15:102:102/16:0:77
#
# upgrade_achievements=19326/40107/40115/40942

# Checksum: 17b399b8`

func ptr[T any](s T) *T {
	return &s
}

func TestBasic(t *testing.T) {
	mani, _ := NewBasicSimManifest(
		api.SimulationConfigBasic{
			Character: api.WowCharacter{
				Name: ptr("Gubulgi"),
				ActiveTalents: api.CharacterTalentLoadout{
					Name:    ptr("Active"),
					Talents: "123",
				},
				Level: 90,
			},
			CoreConfig: api.SimulationCoreConfig{
				FightStyle: ptr(api.DungeonSlice),
			},
			Kind:            api.SimulationConfigBasicKindBasic,
			SimcAddonExport: api.SimcAddonExport(basicProfile),
		},
	)
	profileText, _ := mani.buildSimcProfile()
	log.Print(profileText)
}
