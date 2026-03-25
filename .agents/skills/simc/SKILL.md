---
name: simc
description: >
  Expert guide for SimulationCraft (SimC) — the World of Warcraft combat simulator.
  Use this skill whenever the user asks about simc, SimulationCraft, simming a character,
  APL (Action Priority Lists), stat weights, fight styles, .simc files, profile sets,
  raid events, TCI (Textual Configuration Interface), or running WoW DPS/tank simulations.
  Trigger on phrases like "how do I sim", "write me a simc profile", "explain the APL",
  "fight style options", "scale factors", "profileset", "simc options", "what does X mean
  in simc", or any question about SimulationCraft configuration, output interpretation,
  or advanced simulation techniques.
---

# SimulationCraft (SimC) Skill

SimulationCraft is an open-source, multi-player, event-driven combat simulator for World of Warcraft. It models player DPS/HPS/survival in raid and dungeon scenarios through statistical Monte Carlo iteration — running a fight hundreds or thousands of times and averaging the results. Source: https://github.com/simulationcraft/simc

---

## Quick Reference: Common Tasks

| Goal | What to do |
|------|-----------|
| Sim your character | Use `/simc` addon in-game → paste output into SimC GUI or Raidbots |
| Run from CLI | `simc myprofile.simc html=output.html` |
| Change fight style | `fight_style=Patchwerk` / `DungeonSlice` / `HecticAddCleave` |
| Set iterations | `iterations=10000` |
| Generate stat weights | `calculate_scale_factors=1` |
| Compare gear | Use `profileset` feature (CLI only) |
| Custom APL | `actions=spell_name,if=condition` |
| Output HTML report | `html=report.html` |
| Import from Armory | `armory=us,realm,charactername` |

---

## 1. Installation & Executables

- **Download**: https://www.simulationcraft.org/download.html — Windows installer or ZIP. Linux users build from source.
- **`SimulationCraft.exe`** — Qt-based graphical interface (GUI)
- **`simc.exe`** / `simc` — Command-line interface (CLI); preferred for scripting and advanced use

**Linux build:**
```bash
git clone https://github.com/simulationcraft/simc
cd simc
make OPENSSL=1 -C engine -j$(nproc)
```

---

## 2. Importing a Character

**In-game addon (`/simc`):**
1. Install the SimulationCraft addon
2. Type `/simc` in chat → copy the output
3. Paste into the **Simulate** tab in the GUI, or save as a `.simc` file for CLI use

**Armory import (TCI):**
```
armory=us,illidan,mycharacter
```
Change `us` to your region (`eu`, `kr`, `tw`).

**Inactive/off-spec:**
```
armory=us,illidan,mycharacter,spec=inactive
```

---

## 3. Textual Configuration Interface (TCI)

The TCI is SimC's configuration language. It's used in `.simc` files, the GUI's Overrides tab, or directly as CLI arguments.

**Syntax rules:**
- `option=value` — set an option
- `option+=value` — append to an option (especially useful for `actions` and `raid_events`)
- `#` — comment
- Lines are parsed sequentially; order matters for some options
- Files must be UTF-8 or latin1 encoded

**Running a `.simc` file:**
```bash
simc myprofile.simc
simc myprofile.simc html=result.html json2=result.json
```

**Inline options on the CLI:**
```bash
simc armory=us,illidan,john iterations=10000 html=john.html calculate_scale_factors=1
```

---

## 4. Core Simulation Options

### Fight Configuration

| Option | Default | Description |
|--------|---------|-------------|
| `fight_style` | `Patchwerk` | Combat scenario (see below) |
| `iterations` | `1000` | Number of simulated fights; higher = more accurate |
| `max_time` | `300` | Fight duration in seconds |
| `vary_combat_length` | `0.2` | Random variance in fight length (±20%) |
| `target_error` | `0` | Auto-scale iterations until error ≤ this % |
| `threads` | auto | CPU threads to use |
| `process_priority` | `below_normal` | OS scheduling priority |

### Fight Styles

| Value | Description |
|-------|-------------|
| `Patchwerk` | Pure single-target tank-and-spank; default |
| `HecticAddCleave` | Frequent adds with movement; AoE-heavy |
| `DungeonSlice` | Mimics a Mythic+ dungeon pull pattern |
| `DungeonRoute` | Specific dungeon path (from MDT export) |
| `HeavyMovement` | Frequent long movement phases |
| `LightMovement` | Occasional brief movements |
| `Beastlord` | Mixed single-target, cleave, and AoE |
| `HelterSkelter` | All raid events enabled simultaneously |

```
fight_style=Patchwerk
max_time=300
vary_combat_length=0.2
```

### Target Options

```
# Override target health (ignores max_time, ends on 0 HP)
override.target_health=100000000

# Number of enemies
desired_targets=3

# Target level (for bosses: player_level + 3)
target_level=+3
```

### Performance

```
threads=8
process_priority=below_normal
target_error=0.5    # Stop iterating when error drops below 0.5%
```

---

## 5. Output Options

```
html=report.html           # Full HTML report with charts
json2=report.json          # Structured JSON output
xml=report.xml             # XML output
text=report.txt            # Plain text summary
save=profile.simc          # Save full character profile
save_gear=gear.simc        # Save gear only
save_talents=talents.simc  # Save talents only
save_actions=apl.simc      # Save APL only
```

**Reference player** (compare DPS relative to a baseline):
```
armory=us,illidan,john
armory=us,illidan,bill
reference_player=john
```

---

## 6. Character Profile Structure

A typical `.simc` profile looks like:

```
# Character definition
warrior="MyWarrior"
level=80
race=human
spec=arms
role=attack

# Talents (string from /simc export)
talents=BYQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAgUSiIJRSSikEJBAAAAAAAAAAAAA

# Gear (auto-generated by /simc addon)
head=,id=12345,enchant=stormrider,gem1=bold_inferno_ruby
...

# Action Priority List (uses default if omitted)
actions=auto_attack
actions+=/mortal_strike,if=buff.overpower.up
actions+=/execute,if=target.health.pct<20

# Consumables
consumables.flask=flask_of_supreme_power
consumables.food=well_fed
consumables.augmentation=augmentation
consumables.potion=tempered_potion
use_pre_potion=1
```

---

## 7. Action Priority Lists (APL)

APLs are ordered priority lists — SimC checks from top to bottom and executes the first available action each GCD.

**Basic syntax:**
```
actions=spell_name
actions+=/next_spell,if=condition
actions+=/call_action_list,name=cooldowns
actions.cooldowns=major_cooldown
actions.cooldowns+=/potion,if=buff.bloodlust.react
```

**Conditional expressions (`if=`):**
```
# Buff checks
actions+=/execute,if=buff.sudden_death.react
actions+=/mortal_strike,if=buff.overpower.stack>=2

# Cooldown checks
actions+=/colossus_smash,if=cooldown.colossus_smash.ready
actions+=/slam,if=cooldown.mortal_strike.remains>1.5

# Health/resource
actions+=/execute,if=target.health.pct<20
actions+=/mortal_strike,if=rage>=30

# Time
actions+=/bloodlust,if=time>20
actions+=/potion,if=time_to_die<30

# Enemy count
actions+=/whirlwind,if=active_enemies>=3
```

**Logical operators** (in order of precedence, high→low):
- Unary: `+`, `-`, `@` (abs), `!` (not)
- Math: `*`, `%` (divide), `%%` (modulo), `^` (floor-divide)
- Math: `+`, `-`
- Comparisons: `<`, `>`, `<=`, `>=`, `=`, `!=`
- Logical: `&` (and), `|` (or), `xor`

**Named sub-lists (for readability):**
```
actions=call_action_list,name=precombat,if=!in_combat
actions+=/call_action_list,name=cooldowns
actions+=/call_action_list,name=single_target

actions.precombat=flask
actions.precombat+=/food
actions.precombat+=/snapshot_stats

actions.cooldowns=avatar
actions.cooldowns+=/colossus_smash
```

**Default APLs:** If no `actions=` lines are present, SimC uses the community-maintained default APL for your spec. These live in `ActionPriorityLists/` in the repo.

For more APL expression reference, see: `references/apl-expressions.md`

---

## 8. Stats Scaling (Scale Factors)

Scale factors measure how much 1 point of a stat increases DPS.

```
calculate_scale_factors=1
scale_only=str,crit,haste,mastery,vers   # Limit which stats to scale
iterations=10000                          # Use more iterations for accuracy
```

Results appear in the HTML report. Normalized factors divide everything by the primary stat, making it easy to compare (e.g., "haste = 0.85 means haste is 85% as valuable as strength").

**Stat delta** (how much of a stat is added per pass):
```
default_scale_delta=400    # Default is usually ~300
```

Scale factors can be exported to Pawn (a WoW addon for in-game item comparison) via the link in the HTML report.

---

## 9. Profile Sets (Batch Simulations)

Profile sets simulate a baseline character, then each variant independently. CLI only.

```
# Baseline
warrior=MyWarrior
...

# Each profileset runs separately against the baseline
profileset.with_trinket_a=trinket1=,id=12345
profileset.with_trinket_b=trinket1=,id=67890
profileset.haste_talents=talents=BYQ...alt_string...
```

**Parallel processing:**
```
profileset_work_threads=2   # Threads per profileset
# With threads=8 and work_threads=2 → 4 concurrent workers
```

Output in JSON (`json2=`) includes full statistics; HTML shows median/quartile comparison.

---

## 10. Raid Events

Raid events inject scripted events during the fight. Use `raid_events+=/` to add them.

```
# Movement every 60s for 10s
raid_events+=/movement,cooldown=60,duration=10

# Adds spawning every 30s, 3 adds lasting 15s
raid_events+=/adds,count=3,cooldown=30,duration=15

# Stun the player for 3s at 2 min
raid_events+=/stun,cooldown=120,duration=3

# Vulnerable window (damage amplifier)
raid_events+=/vulnerable,cooldown=45,duration=8,multiplier=1.3
```

**Common event types:** `movement`, `adds`, `stun`, `invulnerable`, `vulnerable`, `heal`, `damage`, `flying`

**Randomization:**
```
raid_events+=/movement,cooldown=60,cooldown_stddev=10,duration=8,duration_stddev=2
```

---

## 11. Buffs & Debuffs Overrides

```
# Disable bloodlust/heroism
override.bloodlust=0

# Force specific external buffs
override.mark_of_the_wild=1
override.battle_shout=1

# Disable all external buffs (solo sim)
override.bloodlust=0
override.arcane_intellect=0
```

---

## 12. PTR / Beta Testing

```
ptr=0   # Live (default)
ptr=1   # PTR version

# Compare live vs PTR twin
ptr=0
armory=us,illidan,john
ptr=1
copy=JohnPTR
```

---

## 13. Reading the Report

**Key metrics:**
- **DPS / Mean**: Average damage per second across all iterations
- **Error %**: Statistical confidence interval — lower is better; <0.5% is generally sufficient
- **DPS Range**: Min/max across all iterations (shows RNG variance)
- **Iterations**: Number of fights simulated
- **Abilities table**: Per-spell DPS contribution, hit counts, crit %, average damage
- **Buff uptimes**: % of fight time each buff was active
- **Sample Sequence**: Cast-by-cast breakdown of one iteration — useful for APL debugging

**Accuracy guidance:**
- 1,000 iterations: Quick check, ~1–3% error
- 10,000 iterations: Good for stat weights and comparisons
- 100,000+ iterations: High-precision theory crafting

---

## 14. Common Patterns & Examples

**Quick single-character sim:**
```bash
simc armory=us,illidan,mychar fight_style=Patchwerk iterations=10000 html=result.html
```

**Gear comparison with profilesets:**
```
# baseline.simc
warrior=MyWarrior
armory=us,illidan,mychar

profileset.item_a+=head=,id=111111,bonus_id=1234
profileset.item_b+=head=,id=222222,bonus_id=5678
```
```bash
simc baseline.simc json2=compare.json html=compare.html
```

**Custom fight with adds:**
```
fight_style=Patchwerk
max_time=300
raid_events+=/adds,count=2,cooldown=30,duration=20,first=30
```

**Tank simulation (survivability focus):**
```
# See references/tanking.md for full tank options
role=tank
position=front
target_adds=1
```

---

## 15. Useful Links

- **Wiki (TCI reference)**: https://github.com/simulationcraft/simc/wiki
- **Options reference**: https://github.com/simulationcraft/simc/wiki/Options
- **APL reference**: https://github.com/simulationcraft/simc/wiki/ActionLists
- **APL expressions**: https://github.com/simulationcraft/simc/wiki/Action-List-Conditional-Expressions
- **Profile sets**: https://github.com/simulationcraft/simc/wiki/ProfileSet
- **Raid events**: https://github.com/simulationcraft/simc/wiki/RaidEvents
- **Raidbots** (cloud SimC): https://www.raidbots.com
- **Discord**: https://discord.gg/tFR2uvK (#simulationcraft)

---

## Reference Files

- `references/apl-expressions.md` — Full APL conditional expression reference (buff/debuff/cooldown/resource/time tokens)
- `references/fight-styles.md` — Detailed fight style descriptions and use cases
- `references/tanking.md` — Tank-specific simulation options
