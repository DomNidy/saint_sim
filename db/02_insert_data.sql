INSERT INTO
    public.simulation_data (sim_result)
VALUES
    (
        '
Raid Events:
  raid_event=/heal,name=tank_heal,amount=60000,cooldown=0.5,duration=0,player_if=role.tank


DPS Ranking:
 331910 100.0%  Raid
 331910  100.0%  Ishton

HPS Ranking:
  33535 100.0%%  Raid
  33331   0.0%  Ishton

Player: Ishton human warrior protection 80
  DPS=331909.6887164972 DPS-Error=652.3742168808973/0.20% DPS-Range=34421.47978467919/10.37%
  DTPS=86661.06820525817 DTPS-Error=343.11713617091783/0.40% DTPS-Range=20563.391750301053/23.73%
  DPR=20934.154791857505 RPS-Out=15.567136781296263 RPS-In=15.828571534305413 Resource=rage Waiting=1.4310013861174489 ApM=69.90831332503677
  Origin: https://worldofwarcraft.com/en-us/character/hydraxis/ishton
  Talents: CkEAmidFBOBFf5oKuZ7r/WeW7YEDAAAAzMzMmxMmZbmlZmZWYGDTjxDYMDglBG2YmZMDzYYAAAAAAAzMAALbbAGGYDWWMaMDgZL2Y2A
  Covenant: invalid
  Core Stats:    strength=41179|39895(34099)  agility=12176|12176(12176)  stamina=297387|283226(169904)  intellect=15258|14814(14814)  spirit=0|0(0)  health=6304604|6004391  mana=0|0
  Generic Stats: mastery=28.87%|25.87%(6474)  versatility=6.13%|3.51%(2739)  leech=3.00%|3.00%(0)  runspeed=7.00%|7.00%(0)
  Spell Stats:   power=15258|14814(0)  hit=18.00%|18.00%(0)  crit=14.64%|15.06%(6340)  haste=19.52%|13.94%(6280)  speed=19.52%|13.94%  manareg=0|0(0)
  Attack Stats:  power=51560|46776(0)  hit=7.50%|7.50%(0)  crit=14.64%|15.06%(6340)  expertise=10.50%|10.50%(0)  haste=19.52%|13.94%(6280)  speed=19.52%|13.94%
  Defense Stats: armor=98529|97408(50167) miss=3.00%|3.00%  dodge=3.00%|3.00%(0)  parry=16.67%|16.79%(0)  block=38.53%|37.69%(0) crit=-6.00%|-6.00%  versatility=3.07%|1.76%(2739)
  Priorities (actions.precombat):
    flask/food/augmentation/snapshot_stats/battle_stance,toggle=on
  Priorities (actions.default):
    auto_attack/charge,if=time=0/use_items/avatar/shield_wall,if=talent.immovable_object.enabled&buff.avatar.down/blood_fury/berserking/arcane_torrent
    lights_judgment/fireblood/ancestral_call/bag_of_tricks/potion,if=buff.avatar.up|buff.avatar.up&target.health.pct<=20
    ignore_pain,if=target.health.pct>=20&(rage.deficit<=15&cooldown.shield_slam.ready|rage.deficit<=40&cooldown.shield_charge.ready&talent.champions_bulwark.enabled|rage.deficit<=20&cooldown.shield_charge.ready|rage.deficit<=30&cooldown.demoralizing_shout.ready&talent.booming_voice.enabled|rage.deficit<=20&cooldown.avatar.ready|rage.deficit<=45&cooldown.demoralizing_shout.ready&talent.booming_voice.enabled&buff.last_stand.up&talent.unnerving_focus.enabled|rage.deficit<=30&cooldown.avatar.ready&buff.last_stand.up&talent.unnerving_focus.enabled|rage.deficit<=20|rage.deficit<=40&cooldown.shield_slam.ready&buff.violent_outburst.up&talent.heavy_repercussions.enabled&talent.impenetrable_wall.enabled|rage.deficit<=55&cooldown.shield_slam.ready&buff.violent_outburst.up&buff.last_stand.up&talent.unnerving_focus.enabled&talent.heavy_repercussions.enabled&talent.impenetrable_wall.enabled|rage.deficit<=17&cooldown.shield_slam.ready&talent.heavy_repercussions.enabled|rage.deficit<=18&cooldown.shield_slam.ready&talent.impenetrable_wall.enabled)|(rage>=70|buff.seeing_red.stack=7&rage>=35)&cooldown.shield_slam.remains<=1&buff.shield_block.remains>=4&set_bonus.tier31_2pc,use_off_gcd=1
    last_stand,if=(target.health.pct>=90&talent.unnerving_focus.enabled|target.health.pct<=20&talent.unnerving_focus.enabled)|talent.bolster.enabled|set_bonus.tier30_2pc|set_bonus.tier30_4pc
    ravager/demoralizing_shout,if=talent.booming_voice.enabled/champions_spear/demolish,if=buff.colossal_might.stack>=3/thunderous_roar
    shockwave,if=talent.rumbling_earth.enabled&spell_targets.shockwave>=3/shield_charge/shield_block,if=buff.shield_block.duration<=10
    run_action_list,name=aoe,if=spell_targets.thunder_clap>=3/call_action_list,name=generic
  Priorities (actions.generic):
    thunder_blast,if=(buff.thunder_blast.stack=2&buff.burst_of_power.stack<=1&buff.avatar.up&talent.unstoppable_force.enabled)|rage<=70&talent.demolish.enabled
    shield_slam,if=(buff.burst_of_power.stack=2&buff.thunder_blast.stack<=1|buff.violent_outburst.up)|rage<=70&talent.demolish.enabled
    execute,if=rage>=70|(rage>=40&cooldown.shield_slam.remains&talent.demolish.enabled|rage>=50&cooldown.shield_slam.remains)|buff.sudden_death.up&talent.sudden_death.enabled
    shield_slam/thunder_blast,if=dot.rend.remains<=2&buff.violent_outburst.down/thunder_blast
    thunder_clap,if=dot.rend.remains<=2&buff.violent_outburst.down
    thunder_blast,if=(spell_targets.thunder_clap>1|cooldown.shield_slam.remains&!buff.violent_outburst.up)
    thunder_clap,if=(spell_targets.thunder_clap>1|cooldown.shield_slam.remains&!buff.violent_outburst.up)
    revenge,if=(rage>=80&target.health.pct>20|buff.revenge.up&target.health.pct<=20&rage<=18&cooldown.shield_slam.remains|buff.revenge.up&target.health.pct>20)|(rage>=80&target.health.pct>35|buff.revenge.up&target.health.pct<=35&rage<=18&cooldown.shield_slam.remains|buff.revenge.up&target.health.pct>35)&talent.massacre.enabled
    execute/revenge,if=target.health>20/thunder_blast,if=(spell_targets.thunder_clap>=1|cooldown.shield_slam.remains&buff.violent_outburst.up)
    thunder_clap,if=(spell_targets.thunder_clap>=1|cooldown.shield_slam.remains&buff.violent_outburst.up)/devastate
  Actions:
    auto_attack_mh                                                 Count= 190.5|  1.892sec  DPE= 77159| 6.21%  DPET= 41217  DPR=     0  pDPS=20607  Miss= 0.00%  Hit= 26846| 21166| 36865  Crit= 53769| 42332| 73730|20.71%
    avatar                                                         Count=   8.9| 36.067sec  DPE=     0| 0.00%  DPET=     0  DPR=     0  pDPS=    0
    champions_spear                                                Count=   3.7| 90.412sec  DPE=959790| 0.00%  DPET=788415  DPR=     0  pDPS=    0
    champions_spear_damage                                         Count=   0.0|  0.000sec  DPE=     0| 3.58%  DPET=     0  DPR=     0  pDPS=11909  Miss= 0.00%  Hit=218850|172093|281831  Crit=442646|344186|563662|20.64%  TickCount=    29  MissTick= 0.00%  Tick= 68722|   599| 91503  CritTick=157293|  1547|205883|21.23%  UpTime=  7.37%
    charge_impact                                                  Count=   1.0|  0.000sec  DPE= 12347| 0.01%  DPET=     0  DPR=     0  pDPS=   42  Miss= 0.00%  Hit= 10526| 10257| 10770  Crit= 21014| 20514| 21540|17.35%
    deep_wounds                                                    Count= 235.8|  1.464sec  DPE= 13305| 3.15%  DPET=     0  DPR=     0  pDPS=10464  TickCount=   120  MissTick= 0.00%  Tick= 21599| 17019| 28743  CritTick= 43710| 34037| 64671|20.61%  UpTime= 99.57%
    demoralizing_shout                                             Count=   9.4| 33.701sec  DPE=     0| 0.00%  DPET=     0  DPR=     0  pDPS=    0
    devastator                                                     Count= 190.5|  1.892sec  DPE= 44735| 8.57%  DPET=     0  DPR=     0  pDPS=28427  Miss= 0.00%  Hit= 36895| 29095| 50676  Crit= 74748| 58190|110562|20.71%
    earthen_ire                                                    Count=  27.1|  5.948sec  DPE= 33319| 0.91%  DPET=     0  DPR=     0  pDPS= 3008  Miss= 0.00%  Hit= 27568| 26848| 30405  Crit= 55151| 53697| 60810|20.85%
    execute                                                        Count=  10.7|  5.697sec  DPE=628086| 0.00%  DPET=483089  DPR= 15758  pDPS=    0
    execute_damage                                                 Count=  10.7|  5.697sec  DPE=595699| 6.41%  DPET=     0  DPR=     0  pDPS=21191  Miss= 0.00%  Hit=491627|210663|665607  Crit=991375|401263|1407819|20.78%
    ground_current_lightning_strike_execute                        Count=   3.0| 15.572sec  DPE=     0| 0.00%  DPET=     0  DPR=     0  pDPS=    0
    ground_current_lightning_strike_revenge                        Count=  13.0| 19.130sec  DPE=     0| 0.00%  DPET=     0  DPR=     0  pDPS=    0
    ground_current_lightning_strike_thunder_blast                  Count=  56.6|  6.759sec  DPE=     0| 0.00%  DPET=     0  DPR=     0  pDPS=    0
    ground_current_lightning_strike_thunder_clap                   Count=   8.1| 33.266sec  DPE=     0| 0.00%  DPET=     0  DPR=     0  pDPS=    0
    ignore_pain                                                    Count=  77.9|  3.080sec  DPE=128970|101.08%  DPET=     0  DPR=  3677  pDPS=33535  Miss= 0.00%  Hit= 66601|    -0|309284
    lightning_strike_execute                                       Count=   3.0| 15.572sec  DPE=115043| 0.35%  DPET=     0  DPR=     0  pDPS= 1153  Miss= 0.00%  Hit= 94718| 76240|126465  Crit=190343|153019|267485|21.22%
    lightning_strike_revenge                                       Count=  13.0| 19.130sec  DPE=115423| 1.51%  DPET=     0  DPR=     0  pDPS= 5042  Miss= 0.00%  Hit= 95578| 76240|128760  Crit=191445|152479|289710|20.70%
    lightning_strike_thunder_blast                                 Count=  56.6|  6.759sec  DPE=123863| 7.05%  DPET=     0  DPR=     0  pDPS=23382  Miss= 0.00%  Hit=101861| 76240|128760  Crit=207224|152479|289710|20.88%
    lightning_strike_thunder_clap                                  Count=   8.1| 33.266sec  DPE=118446| 0.97%  DPET=     0  DPR=     0  pDPS= 3207  Miss= 0.00%  Hit= 97795| 76240|128760  Crit=197252|152479|289661|20.74%
    potion                                                         Count=   1.4|  0.000sec  DPE=     0| 0.00%  DPET=     0  DPR=     0  pDPS=    0
    ravager                                                        Count=   3.7| 90.296sec  DPE=697558| 0.00%  DPET=572863  DPR=     0  pDPS=    0  TickCount=    22  UpTime= 11.90%
    ravager_tick                                                   Count=   0.0|  0.000sec  DPE=     0| 2.61%  DPET=     0  DPR=     0  pDPS= 8689  Miss= 0.00%  Hit= 92462| 71641|117324  Crit=204527|143281|263978|21.14%
    rend                                                           Count=  70.1|  4.241sec  DPE= 65027| 4.59%  DPET=     0  DPR=     0  pDPS=15213  TickCount=   118  MissTick= 0.00%  Tick= 31963|  1908| 42736  CritTick= 64519| 15193| 96156|20.61%  UpTime= 98.26%
    revenge                                                        Count=  45.3|  5.808sec  DPE=107491| 3.38%  DPET= 86150  DPR=  6108  pDPS=11274  Miss= 0.00%  Hit= 61436| 49807| 84119  Crit=123614| 99615|189267|20.63%
    shadowed_essence                                               Count=  39.3|  7.459sec  DPE=103443| 0.00%  DPET=     0  DPR=     0  pDPS=    0
    shadowed_essence_damage                                        Count=  39.3|  7.457sec  DPE=103567| 4.09%  DPET=     0  DPR=     0  pDPS=13556  Miss= 0.00%  Hit= 85916| 83806| 94909  Crit=171912|167612|189817|20.67%
    shield_block                                                   Count=  23.6| 12.752sec  DPE=     0| 0.00%  DPET=     0  DPR=     0  pDPS=    0
    shield_charge                                                  Count=   7.0| 45.359sec  DPE=525404| 0.00%  DPET=417468  DPR=     0  pDPS=    0
    shield_charge_main                                             Count=   7.0| 45.359sec  DPE=525404| 3.71%  DPET=     0  DPR=     0  pDPS=12326  Miss= 0.00%  Hit=426542|334794|565429  Crit=918647|669588|1272215|20.09%
    shield_slam                                                    Count=  85.6|  3.462sec  DPE=302502|26.03%  DPET=242010  DPR=     0  pDPS=86430  Miss= 0.00%  Hit=246466|151576|754296  Crit=517675|318310|1677745|20.58%
    sidearm                                                        Count=  38.2|  7.910sec  DPE= 25491| 0.98%  DPET=     0  DPR=     0  pDPS= 3245  Miss= 0.00%  Hit= 21073| 16607| 28924  Crit= 42506| 33214| 63106|20.68%
    sigil_of_algari_concordance                                    Count=   2.8| 82.953sec  DPE=956785| 0.00%  DPET=     0  DPR=     0  pDPS=    0
    thunder_blast                                                  Count=  43.2|  6.759sec  DPE=422770|11.30%  DPET=340256  DPR=     0  pDPS=37510  Miss= 0.00%  Hit=214232| 91008|691660  Crit=435051|182016|1436581|20.95%
    thunder_clap                                                   Count=  27.0| 10.694sec  DPE=140672| 2.85%  DPET=112575  DPR=     0  pDPS= 9432  Miss= 0.00%  Hit= 86864| 45662|339680  Crit=174752| 91324|670482|20.76%
   boulderbane (DPS=42751.94269740956)
    earthen_ire_buff                                               Count=   2.8| 82.245sec  DPE=     0| 0.00%  DPET=     0  DPR=     0  pDPS=    0
    mighty_smash                                                   Count=  11.0| 15.938sec  DPE=159885| 1.77%  DPET=     0  DPR=     0  pDPS=42752  Miss= 0.00%  Hit=133337|129791|146986  Crit=266709|259583|293973|19.91%

  Constant Buffs:
    arcane_intellect/battle_shout/battle_stance/crystallization/flask_of_alchemical_chaos/mark_of_the_wild/power_word_fortitude/skyfury/well_fed
  Dynamic Buffs:
    avatar                            : start= 10.4 refresh=  0.0 interval= 29.8 trigger= 30.1 duration= 17.4 uptime= 60.42%  benefit= 62.25%
    bloodlust                         : start=  1.0 refresh=  0.0 interval=  0.0 trigger=  0.0 duration= 40.0 uptime= 13.53%
    brace_for_impact                  : start=  1.0 refresh= 84.6 interval=  0.0 trigger=  3.5 duration=295.0 uptime= 98.36%  benefit= 98.81%
    burst_of_power                    : start= 10.7 refresh=  1.0 interval= 26.3 trigger= 23.9 duration=  4.4 uptime= 15.59%  benefit= 24.93%
    dark_embrace                      : start= 10.3 refresh=  1.0 interval= 30.5 trigger= 30.0 duration= 27.2 uptime= 93.75%
    earthen_ire                       : start=  2.8 refresh=  0.0 interval= 78.7 trigger= 78.4 duration=  9.8 uptime=  9.05%
    flask_of_alchemical_chaos_crit    : start=  2.1 refresh=  0.6 interval=111.3 trigger= 76.0 duration= 35.3 uptime= 25.09%
    flask_of_alchemical_chaos_haste   : start=  2.1 refresh=  0.6 interval=113.4 trigger= 77.1 duration= 35.4 uptime= 24.64%
    flask_of_alchemical_chaos_mastery : start=  2.1 refresh=  0.7 interval=112.0 trigger= 74.7 duration= 36.1 uptime= 25.55%
    flask_of_alchemical_chaos_vers    : start=  2.1 refresh=  0.6 interval=112.5 trigger= 78.3 duration= 35.1 uptime= 24.72%
    ignore_pain                       : start= 19.7 refresh= 77.1 interval= 15.3 trigger=  3.0 duration= 10.7 uptime= 69.86%
    revenge                           : start=  8.8 refresh=  0.2 interval= 35.1 trigger= 34.2 duration=  4.5 uptime= 13.32%  benefit= 11.02%
    seeing_red                        : start= 19.8 refresh=135.4 interval= 15.2 trigger=  1.9 duration= 13.6 uptime= 90.07%
    shadowed_essence                  : start=  9.8 refresh=  0.0 interval= 30.0 trigger= 30.0 duration= 26.7 uptime= 87.34%
    shield_block                      : start=  1.0 refresh=  0.0 interval=  0.0 trigger=  0.0 duration=297.0 uptime= 99.01%
    tempered_potion                   : start=  1.4 refresh=  0.0 interval=308.5 trigger=  0.0 duration= 27.3 uptime= 12.74%
    thunder_blast                     : start= 24.6 refresh= 14.2 interval= 12.2 trigger=  7.7 duration=  4.9 uptime= 40.15%
    violent_outburst                  : start= 19.0 refresh=  0.0 interval= 15.5 trigger= 15.5 duration=  1.9 uptime= 10.96%
    wild_strikes                      : start= 14.3 refresh=  4.8 interval= 21.2 trigger= 15.8 duration= 12.8 uptime= 61.40%
  Up-Times:
      5.10% : Rage Cap                      
  Procs:
     31.57353 |   9.32209sec : Skyfury (Main Hand)
      8.37941 |  32.73085sec : parry_haste
  Gains:
        86.6 : avatar                 (rage)           (overflow=2.18%)
       112.0 : bloodsurge             (rage)           (overflow=5.60%)
       260.7 : booming_voice          (rage)           (overflow=7.34%)
        69.7 : champions_spear        (rage)           (overflow=6.12%)
        20.0 : charge_impact          (rage)         
        39.8 : frothing_berserker     (rage)         
    10057853.3 : ignore_pain            (health)       
       542.9 : melee_main_hand        (rage)           (overflow=4.97%)
       189.7 : rage_from_damage_taken (rage)           (overflow=4.58%)
       206.3 : ravager                (rage)           (overflow=7.96%)
       249.7 : shield_charge          (rage)           (overflow=11.09%)
      1977.3 : shield_slam            (rage)           (overflow=3.55%)
       383.4 : thorims_might          (rage)           (overflow=4.91%)
       464.9 : thunder_blast          (rage)           (overflow=4.17%)
       140.4 : thunder_clap           (rage)           (overflow=1.52%)
    Pet "boulderbane" Gains:
  Waiting:  1.43%


 *** Targets *** 

Target: Fluffy_Pillow humanoid tank_dummy unknown 83
  DPS=86661.0682052581 DPS-Error=0/0.00% DPS-Range=20563.391750301067/23.73%
  DTPS=331909.68871649704 DTPS-Error=652.374216880899/0.20% DTPS-Range=34421.4797846787/10.37%
  DPR=0.2905933777492882 RPS-Out=298123.57031870814 RPS-In=0 Resource=health Waiting=46.448230220757694 ApM=3.3702427856970383
  Core Stats:    strength=0|0(0)  agility=0|0(0)  stamina=0|0(0)  intellect=0|0(0)  spirit=0|0(0)  health=0|84546167  mana=0|0
  Generic Stats: mastery=0.00%|0.00%(0)  versatility=0.00%|0.00%(0)  leech=0.00%|0.00%(0)  runspeed=7.00%|7.00%(0)
  Spell Stats:   power=0|0(0)  hit=0.00%|0.00%(0)  crit=0.00%|0.00%(0)  haste=0.00%|0.00%(0)  speed=0.00%|0.00%  manareg=0|0(0)
  Attack Stats:  power=0|0(0)  hit=0.00%|0.00%(0)  crit=5.00%|5.00%(0)  expertise=0.00%|0.00%(0)  haste=0.00%|0.00%(0)  speed=0.00%|0.00%
  Defense Stats: armor=42857|42857(42857) miss=3.00%|3.00%  dodge=3.00%|3.00%(0)  parry=3.00%|3.00%(0)  block=3.00%|3.00%(0) crit=0.00%|0.00%  versatility=0.00%|0.00%(0)
  Priorities (actions.precombat):
    snapshot_stats
  Priorities (actions.default):
    auto_attack,damage=1500000,range=30000,attack_speed=2,aoe_tanks=1
    melee_nuke,damage=3000000,range=60000,attack_speed=2,cooldown=30,aoe_tanks=1
    spell_dot,damage=60000,range=1200,tick_time=2,cooldown=60,aoe_tanks=1,dot_duration=60,bleed=1
    pause_action,duration=30,cooldown=30,if=time>=30
  Actions:
    melee_main_hand_Ishton  Count=  77.5|  3.760sec  DPE=222949|66.47%  DPET=111475  DPR=  0  pDPS=57562  Miss= 0.00%  Parry=14.45%
    melee_nuke_Ishton       Count=   5.5| 59.564sec  DPE=647770|13.65%  DPET=323178  DPR=  0  pDPS=11874  Miss= 0.00%
    pause_action            Count=   4.5| 60.014sec  DPE=     0| 0.00%  DPET=     0  DPR=  0  pDPS=    0
    spell_dot_Ishton        Count=   5.4| 61.014sec  DPE=964440|19.88%  DPET=960116  DPR=  0  pDPS=17225  TickCount=   145  MissTick= 0.00%  Tick= 35672| 21818| 58160  UpTime= 96.53%
    tank_heal               Count= 600.3|  0.500sec  DPE=     0| 0.00%  DPET=     0  DPR=  0  pDPS=    0

  Constant Buffs:
    arcane_intellect/battle_shout/bleeding/chaos_brand/hunters_mark/mark_of_the_wild/mortal_wounds/mystic_touch/power_word_fortitude/skyfury
  Dynamic Buffs:
    champions_might           : start=  3.7 refresh=  0.0 interval= 90.0 trigger= 90.4 duration=  6.0 uptime=  7.38%
    demoralizing_shout_debuff : start=  9.4 refresh=  0.0 interval= 33.6 trigger= 33.6 duration=  7.9 uptime= 24.76%  benefit= 23.42%
    punish                    : start=  2.2 refresh= 83.4 interval=128.1 trigger=  3.5 duration=131.8 uptime= 97.99%  benefit= 99.22%
  Procs:
      4.98431 |  60.01349sec : delayed_aa_cast
  Waiting: 46.45%


Baseline Performance:
  Networking    = enabled
  RNG Engine    = xoshiro256+
  Iterations    = 1032 (94, 88, 89, 69, 97, 80, 98, 75, 94, 81, 96, 71)
  TotalEvents   = 6668449
  MaxEventQueue = 79
  TargetHealth  = 84546167
  SimSeconds    = 305888.942
  CpuSeconds    = 14.6219925
  WallSeconds   = 4.361831661
  InitSeconds   = 0.201220454
  MergeSeconds  = 0.010195036
  AnalyzeSeconds= 0.001260423
  SpeedUp       = 21166
  EndTime       = 2024-09-09 15:26:45+0000 (1725895605)


Waiting:
     1.43% : Ishton

'
    );