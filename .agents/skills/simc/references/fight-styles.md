# SimC Fight Styles Reference

Fight styles control the combat scenario SimC simulates. Set with `fight_style=<name>`.

---

## Built-in Fight Styles

### `Patchwerk` (default)
Pure single-target, stand-still tank-and-spank. No movement, no adds, no interruptions.
- Best for: single-target DPS rankings, gear comparisons, stat weights
- Not representative of real raid encounters but is the standard baseline

```
fight_style=Patchwerk
max_time=300
```

---

### `HecticAddCleave`
Frequent waves of adds with substantial movement phases. Heavily favors AoE and cleave specs.
- Best for: AoE/cleave potential, Mythic+ multi-target throughput
- Spawns multiple short-lived add waves with movement between them

---

### `DungeonSlice`
Models a realistic Mythic+ dungeon — a mix of AoE trash pulls and single-target bosses.
- Best for: Mythic+ performance estimates
- Uses a curated sequence of pull sizes and mob health values
- Recommended for players optimizing for keys

```
fight_style=DungeonSlice
```

---

### `DungeonRoute`
Imports an actual MDT (Method Dungeon Tools) route as raid events, simulating the exact pull sequence.
- Requires a SimC export from MDT
- Most accurate Mythic+ simulation option

Usage: Export route from MDT → paste the raid_events block into your profile.

---

### `HeavyMovement`
Long, frequent movement phases. Very punishing for casters and melee with movement penalties.
- 25s movement every 45s by default
- Good for assessing movement DPS loss

---

### `LightMovement`
Brief, infrequent movement phases.
- 5s movement every 85s by default
- Modest performance impact

---

### `Beastlord`
Mixed encounter: starts single-target, transitions to adds, then back. Inspired by WoD encounter design.
- Tests sustained single-target + burst AoE + cleave combination

---

### `HelterSkelter`
All raid events enabled simultaneously: movement, adds, stuns, interrupts.
- Maximum chaos scenario
- Rarely used for serious comparisons; useful for stress-testing APLs

---

### `CastingPatchwerk`
Like Patchwerk but the target is a caster enemy — periodically casts interruptible spells.
- Useful for testing interrupt handling in APLs

---

## Custom Fight Configuration

You can approximate any fight style by combining `fight_style=Patchwerk` with custom `raid_events`:

```
fight_style=Patchwerk
max_time=300
vary_combat_length=0.2

# Two adds every 45s, lasting 20s starting at 30s
raid_events+=/adds,count=2,cooldown=45,duration=20,first=30

# 8s movement phase every 60s
raid_events+=/movement,cooldown=60,duration=8

# Stun for 3s at 90s, then every 120s
raid_events+=/stun,cooldown=120,duration=3,first=90

# 10s invulnerability window every 2min
raid_events+=/invulnerable,cooldown=120,duration=10,first=60
```

---

## Raid Event Reference

| Event | Key Params | Description |
|-------|-----------|-------------|
| `adds` | `count`, `duration`, `cooldown`, `first` | Spawns additional enemies |
| `movement` | `duration`, `cooldown` | Forced player movement |
| `stun` | `duration`, `cooldown` | Stuns the player |
| `invulnerable` | `duration`, `cooldown` | Target becomes immune to damage |
| `vulnerable` | `duration`, `cooldown`, `multiplier` | Target takes bonus damage |
| `damage` | `amount`, `cooldown` | Deals damage to player |
| `heal` | `amount`, `cooldown` | Heals the player |
| `interrupt` | N/A | Interrupts player's cast |
| `flying` | `duration`, `cooldown` | Target becomes untargetable |
| `distraction` | `duration`, `cooldown` | Penalizes player APL decisions |

**Common parameters for all events:**
```
cooldown=<seconds>          # Time between occurrences
cooldown_stddev=<seconds>   # Random variance in timing
duration=<seconds>          # How long the event lasts
duration_stddev=<seconds>   # Random variance in duration
first=<seconds>             # Time of first occurrence (default: cooldown)
last=<seconds>              # Time of last occurrence
```

---

## Choosing a Fight Style

| Scenario | Recommended Style |
|----------|------------------|
| Boss DPS ranking | `Patchwerk` |
| Council / cleave fight | `Patchwerk` + `adds` with high duration |
| Mythic+ | `DungeonSlice` or `DungeonRoute` |
| Movement-heavy boss | `HeavyMovement` or custom `movement` events |
| Stat weight calculation | `Patchwerk` with high iterations (10k+) |
| Realistic progression sim | Custom combination with `raid_events` |
