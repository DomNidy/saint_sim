# SimC Tank Simulation Reference

SimulationCraft supports tank simulation. The primary output metric shifts from DPS to survivability and DTPS (damage taken per second), though DPS is still tracked.

---

## Basic Tank Profile Setup

```
# Declare the character as a tank
warrior=MyTank
level=80
race=human
spec=protection
role=tank
position=front          # Tank stands in front (melee range, front arc)

# Target options for tanking
target_level=+3         # Boss-level enemy (standard)
desired_targets=1
```

---

## Tank-Specific Options

| Option | Default | Description |
|--------|---------|-------------|
| `role=tank` | `attack` | Sets character role; changes APL and metrics |
| `position=front` | `back` | Melee tanks stand at the front |
| `target_adds=1` | `0` | Tank picks up all adds |
| `tanks_consume_bloodlust=0` | `1` | Whether tank receives bloodlust |
| `override.target_health` | `0` | Set boss HP pool directly |

---

## Survivability Metrics in the Report

The HTML report for tanks includes additional sections:

- **DTPS (Damage Taken Per Second)**: Primary survivability metric
- **TMI (Theck-Meloree Index)**: Spike damage metric — lower is better. Models how much you challenge your healer. More useful than raw DTPS for comparing builds.
- **Effective Health**: Raw EHP after mitigation
- **Absorption**: From shields, blocks, etc.

---

## TMI (Theck-Meloree Index)

TMI measures damage spikiness — how often large damage windows occur. It's the gold standard for tank survivability in SimC.

```
# Enable TMI calculation (auto-enabled for tanks)
tmi_window=6    # Rolling window in seconds (default: 6)
```

Lower TMI = smoother damage intake = easier to heal.

---

## Custom Tank Encounter

```
fight_style=Patchwerk
max_time=300

warrior=MyTank
level=80
spec=protection
role=tank
position=front

# Simulate incoming tank damage pattern
# (SimC applies boss auto-attack based on target_level)
target_level=+3

# Add some magic damage
raid_events+=/damage,cooldown=10,duration=0,amount=80000,type=magic

# Simulate healer keeping you up
# (Tank sim assumes a healer; no explicit configuration needed)
```

---

## Self-Healing / Absorption

Tank APLs often include self-healing. These work identically to DPS APLs:

```
actions+=/victory_rush,if=target.health.pct<20&health.pct<70
actions+=/ignore_pain,if=rage>=40&incoming_damage_1500ms>0
```

**Relevant APL tokens for tanks:**
| Token | Description |
|-------|-------------|
| `health.pct` | Player's current health percentage |
| `health.deficit` | Missing health in absolute value |
| `incoming_damage_1500ms` | Damage taken in the last 1.5 seconds |
| `incoming_damage_3s` | Damage taken in the last 3 seconds |
| `absorb_pct` | Current absorb shield as % of max health |

---

## Tank Profile Comparison Example

```
# Compare two tank trinkets for survivability

warrior=MyTank
armory=us,illidan,mytank
role=tank

# Baseline (currently equipped)

profileset.defensive_trinket=trinket1=,id=111111,bonus_id=1234
profileset.offensive_trinket=trinket1=,id=222222,bonus_id=5678
```

Run with:
```bash
simc tankprofile.simc json2=tank_compare.json html=tank_compare.html
```

The JSON output includes both DPS and TMI for each profileset, allowing survivability vs. output tradeoffs to be assessed.

---

## Notes

- Tank simulation is less mature than DPS simulation in SimC. Results are directionally correct but may not perfectly model all mechanics.
- TMI is highly sensitive to fight length and `vary_combat_length`. Use consistent settings when comparing.
- Self-healing effectiveness depends heavily on APL quality.
- See https://github.com/simulationcraft/simc/wiki/SimcForTanks for the full official reference.
