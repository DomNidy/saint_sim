# SimC APL Conditional Expression Reference

This file documents the tokens and expressions available in SimulationCraft Action Priority List (APL) `if=` conditions.

---

## Operator Precedence (high → low)

| Level | Operators | Notes |
|-------|-----------|-------|
| 1 | `+x`, `-x`, `@x`, `!x` | Unary: positive, negative, absolute value, logical NOT |
| 2 | `*`, `%`, `%%`, `^` | Multiply, divide, modulo, floor-divide |
| 3 | `+`, `-` | Add, subtract |
| 4 | `<`, `>`, `<=`, `>=`, `=`, `!=` | Comparison (return 0 or 1) |
| 5 | `&` | Logical AND |
| 6 | `xor` | Logical XOR |
| 7 | `\|` | Logical OR |

All expressions evaluate as floating-point; `0` is false, anything nonzero is true.

---

## Time & Combat State

| Token | Description |
|-------|-------------|
| `time` | Elapsed combat time in seconds |
| `time_to_die` | Estimated seconds until target dies |
| `expected_combat_length` | Projected total fight duration for this iteration |
| `fight_remains` | `expected_combat_length - time` |
| `in_combat` | 1 if in combat, 0 otherwise |
| `active_enemies` | Number of active enemy targets |
| `active_allies` | Number of active allied players |

---

## Player Resources

| Token | Description |
|-------|-------------|
| `rage` | Current rage |
| `rage.max` | Max rage |
| `rage.deficit` | `rage.max - rage` |
| `energy` | Current energy |
| `energy.regen` | Energy regeneration rate |
| `mana` | Current mana |
| `mana.pct` | Mana as a percentage |
| `focus` | Current focus |
| `runic_power` | Current runic power |
| `combo_points` | Current combo points |
| `soul_shards` | Current soul shards |
| `holy_power` | Current holy power |
| `astral_power` | Current astral power (druid) |
| `chi` | Current chi (monk) |
| `insanity` | Current insanity (priest) |
| `fury` | Current fury (demon hunter) |
| `pain` | Current pain (vengeance DH) |
| `maelstrom` | Current maelstrom (shaman) |

---

## Buff Expressions

**Syntax:** `buff.<buff_name>.<property>`

The buff name uses underscores for spaces, ignoring non-alphanumeric characters (tokenized).

| Property | Description |
|----------|-------------|
| `up` | 1 if buff is active, 0 if not |
| `down` | 1 if buff is NOT active |
| `react` | 1 if buff is active and has been processed (avoids ICD issues) |
| `stack` | Current stack count |
| `stack_pct` | Stack count as % of max stacks |
| `remains` | Seconds until buff expires (0 if not active) |
| `duration` | Total duration of the buff |
| `value` | Custom buff value (used by some effects) |
| `cooldown_remains` | Cooldown before buff can be applied again |

**Examples:**
```
buff.bloodlust.up
buff.sudden_death.react
buff.overpower.stack>=2
buff.avatar.remains>gcd
```

---

## Debuff Expressions

**Syntax:** `debuff.<debuff_name>.<property>` (same properties as buffs)

`debuff.X` checks the current target first, then falls back to player buffs. `buff.X` only checks player buffs.

```
debuff.colossus_smash.up
debuff.faerie_fire.stack<3
```

---

## Cooldown Expressions

**Syntax:** `cooldown.<spell_name>.<property>`

| Property | Description |
|----------|-------------|
| `ready` | 1 if off cooldown |
| `remains` | Seconds until ready (0 if ready) |
| `duration` | Total cooldown duration |
| `charges` | Charges available (for spells with charges) |
| `max_charges` | Maximum charges |
| `charges_fractional` | Fractional charge count (e.g., 1.5) |
| `recharge_time` | Time until next charge recharges |

**Examples:**
```
cooldown.mortal_strike.ready
cooldown.mortal_strike.remains>1.5
cooldown.avatar.charges>=1
```

---

## Dot (DoT/HoT) Expressions

**Syntax:** `dot.<spell_name>.<property>`

| Property | Description |
|----------|-------------|
| `ticking` | 1 if DoT is active |
| `remains` | Seconds until DoT expires |
| `duration` | Total duration |
| `tick_time_remains` | Time until next tick |
| `stack` | Stack count (for stacking DoTs) |
| `pmultiplier` | Persistent multiplier snapshot |

```
dot.rend.ticking
dot.rend.remains<gcd
```

---

## Target Expressions

| Token | Description |
|-------|-------------|
| `target.health.pct` | Target health percentage |
| `target.health` | Target current health |
| `target.time_to_die` | Time until current target dies |
| `target.debuff.<name>.up` | Debuff on target |
| `target.distance` | Distance from player to target |
| `target.in_range` | 1 if target is in melee range |

---

## Spell / Ability Expressions

| Token | Description |
|-------|-------------|
| `spell_targets.<spell_name>` | Number of targets hit by the spell |
| `prev_gcd.1.<spell_name>` | 1 if this spell was cast in the previous GCD |
| `prev_gcd.2.<spell_name>` | 1 if cast 2 GCDs ago |
| `prev_off_gcd.<spell_name>` | 1 if cast off-GCD most recently |
| `cast_time` | Cast time of current spell |
| `gcd` | Current GCD duration |

---

## Pet Expressions

**Syntax:** `pet.<pet_name>.<property>`

```
pet.water_elemental.cooldown.freeze.remains<gcd
pet.ghoul.active
```

---

## Talent Expressions

**Syntax:** `talent.<talent_name>.enabled`

```
talent.avatar.enabled
```

---

## Equipped Item Expressions

| Token | Description |
|-------|-------------|
| `equipped.<item_name>` | 1 if player has item equipped |
| `trinket.1.has_proc.any` | 1 if trinket in slot 1 has any proc |
| `trinket.2.cooldown.remains` | Cooldown of trinket in slot 2 |

---

## Special Keywords

| Keyword | Description |
|---------|-------------|
| `gcd` | Current GCD length in seconds |
| `spell_haste` | Player's spell haste factor (1.0 = no haste) |
| `attack_haste` | Player's attack (melee) haste factor |
| `mastery_value` | Player's mastery rating as a multiplier |
| `crit_pct` | Player's crit chance percentage |
| `haste_pct` | Player's haste percentage |
| `versatility_damage_pct` | Versatility bonus to damage |

---

## Example APL Snippets

```
# Use execute on low health or sudden death proc
actions+=/execute,if=target.health.pct<20|buff.sudden_death.react

# Use major cooldown when colossus smash is up
actions+=/avatar,if=debuff.colossus_smash.up&cooldown.colossus_smash.remains<2

# Pool rage before spending
actions+=/mortal_strike,if=rage>=60|cooldown.avatar.remains>20

# Refresh DoT only when about to fall off
actions+=/rend,if=dot.rend.remains<gcd

# Use potion on pull or with heroism
actions+=/potion,if=time<=2|buff.bloodlust.react

# Multi-target cleave threshold
actions+=/whirlwind,if=active_enemies>=3

# Sub-list call
actions+=/call_action_list,name=cooldowns,if=target.time_to_die>20

# Conditional use of trinket
actions+=/use_item,name=my_trinket,if=buff.avatar.up
```
