// Package sims renders sanitized SimulationCraft profile fragments.
package sims

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/DomNidy/saint_sim/internal/api"
)

var (
	errInvalidCharacterClass = errors.New("invalid character class")
	errInvalidEquipmentSlot  = errors.New("invalid equipment slot")
	errInvalidIdentifier     = errors.New("invalid simc identifier")
	errInvalidLevel          = errors.New("invalid character level")
	errInvalidProfileName    = errors.New("invalid simc profile name")
	errInvalidTalents        = errors.New("invalid talents")
)

// equipmentRawline converts an EquipmentItem into a SimulationCraft TCI equipment line.
//
// SimC equipment lines generally look like:
//
//	main_hand=,id=249671,enchant_id=3368,bonus_id=12786/6652
//
// The format is:
//
//	<slot>=,<property>=<value>,<property>=<value>,...
//
// Notes:
//   - The slot name must match SimC's expected slot key.
//   - Properties are comma-separated.
//   - Slice-like values such as bonus_id and crafted_stats are slash-separated.
//   - Optional pointer fields are emitted only when non-nil.
//   - Nil slice pointers and empty slices are safely skipped.
func equipmentRawline(item api.EquipmentItem) (string, error) {
	return equipmentRawlineForSlot(item, item.Slot)
}

// equipmentRawlineForSlot renders an equipment line using an explicit target
// slot. This is used for interchangeable slots such as rings and trinkets,
// where the same item can be assigned to either slot without trusting a raw
// input line's slot prefix.
func equipmentRawlineForSlot(item api.EquipmentItem, slot api.EquipmentSlot) (string, error) {
	slotValue, err := slotString(slot)
	if err != nil {
		return "", err
	}

	fields := []string{slotValue + "="}
	fields = append(fields, intField("id", item.ItemId))
	fields = appendOptionalIntField(fields, "enchant_id", item.EnchantId)
	fields = appendOptionalIntListField(fields, "gem_id", item.GemIds)
	fields = appendOptionalIntListField(fields, "bonus_id", item.BonusIds)
	fields = appendOptionalIntListField(fields, "crafted_stats", item.CraftedStats)
	fields = appendOptionalIntField(fields, "crafting_quality", item.CraftingQuality)

	return strings.Join(fields, ","), nil
}

// slotString returns the SimC slot key for supported combat equipment slots.
// Cosmetic slots and unknown values are rejected because they are not useful
// for generated sim profiles and should not become arbitrary profile keys.
func slotString(slot api.EquipmentSlot) (string, error) {
	switch slot {
	case api.Back, api.Chest, api.Feet, api.Finger1, api.Finger2, api.Hands,
		api.Head, api.Legs, api.MainHand, api.Neck, api.OffHand, api.Shoulder,
		api.Trinket1, api.Trinket2, api.Waist, api.Wrist:
		return string(slot), nil
	case api.Shirt, api.Tabard:
		return "", fmt.Errorf("%w: %s", errInvalidEquipmentSlot, string(slot))
	default:
		return "", fmt.Errorf("%w: %s", errInvalidEquipmentSlot, string(slot))
	}
}

// characterClassRawline returns the SimC player line key for a generated API
// character class. Rejecting unknown values prevents a caller from turning the
// class position into another SimC option such as proxy=.
func characterClassRawline(class api.CharacterClass) (string, error) {
	switch class {
	case api.Deathknight, api.Demonhunter, api.Druid, api.Evoker, api.Hunter,
		api.Mage, api.Monk, api.Paladin, api.Priest, api.Rogue, api.Shaman,
		api.Warlock, api.Warrior:
		return string(class), nil
	default:
		return "", fmt.Errorf("%w: %s", errInvalidCharacterClass, string(class))
	}
}

// characterNameRawline renders the opening SimC player line, for example
// monk="Celinka". The name is quoted by SimC syntax, so quotes, backslashes,
// and line breaks are rejected instead of escaped.
func characterNameRawline(class api.CharacterClass, name *string) (string, error) {
	classValue, err := characterClassRawline(class)
	if err != nil {
		return "", err
	}

	nameValue := "UnknownCharacter"
	if name != nil && *name != "" {
		nameValue = *name
	}

	if !isSafeQuotedValue(nameValue) {
		return "", fmt.Errorf("%w: %q", errInvalidProfileName, nameValue)
	}

	return fmt.Sprintf(`%s="%s"`, classValue, nameValue), nil
}

// characterBaseRawlines renders the shared character header used by basic and
// top gear simulations. Every caller should use this helper instead of
// interpolating user-controlled fields directly into profile text.
func characterBaseRawlines(
	class api.CharacterClass,
	name *string,
	level int,
	race string,
	spec string,
) ([]string, error) {
	nameLine, err := characterNameRawline(class, name)
	if err != nil {
		return nil, err
	}

	levelLine, err := levelRawline(level)
	if err != nil {
		return nil, err
	}

	raceLine, err := identifierRawline("race", race)
	if err != nil {
		return nil, err
	}

	specLine, err := identifierRawline("spec", spec)
	if err != nil {
		return nil, err
	}

	return []string{
		nameLine,
		levelLine,
		raceLine,
		specLine,
		"iterations=5",
	}, nil
}

// equippedItemsRawlines renders the complete equipped gear block from the
// structured CharacterEquippedItems object. The field position determines the
// emitted SimC slot, so callers do not rely on any untrusted raw line prefix or
// on a possibly mismatched EquipmentItem.Slot value.
func equippedItemsRawlines(items api.CharacterEquippedItems) ([]string, error) {
	type equippedSlot struct {
		item api.EquipmentItem
		slot api.EquipmentSlot
	}

	slots := []equippedSlot{
		{items.Head, api.Head},
		{items.Neck, api.Neck},
		{items.Shoulder, api.Shoulder},
		{items.Back, api.Back},
		{items.Chest, api.Chest},
		{items.Wrist, api.Wrist},
		{items.Hands, api.Hands},
		{items.Waist, api.Waist},
		{items.Legs, api.Legs},
		{items.Feet, api.Feet},
		{items.Finger1, api.Finger1},
		{items.Finger2, api.Finger2},
		{items.Trinket1, api.Trinket1},
		{items.Trinket2, api.Trinket2},
		{items.MainHand, api.MainHand},
	}
	if items.OffHand != nil {
		slots = append(slots, equippedSlot{*items.OffHand, api.OffHand})
	}

	lines := make([]string, 0, len(slots))
	for _, slot := range slots {
		line, err := equipmentRawlineForSlot(slot.item, slot.slot)
		if err != nil {
			return nil, err
		}
		lines = append(lines, line)
	}

	return lines, nil
}

// levelRawline renders the character level assignment and rejects impossible
// levels before they reach SimC.
func levelRawline(level int) (string, error) {
	if level <= 0 {
		return "", fmt.Errorf("%w: %d", errInvalidLevel, level)
	}

	return fmt.Sprintf("level=%d", level), nil
}

// identifierRawline renders a simple key=value assignment where value must be
// a SimC identifier-like token. It is intended for fields such as race and spec,
// not for free-form strings.
func identifierRawline(key string, value string) (string, error) {
	if !isSafeIdentifier(value) {
		return "", fmt.Errorf("%w: %s=%q", errInvalidIdentifier, key, value)
	}

	return key + "=" + value, nil
}

// talentsRawline renders the active talent string. The talent payload is
// otherwise opaque to the worker, but line breaks are rejected so it cannot
// append additional SimC options.
func talentsRawline(talents string) (string, error) {
	if talents == "" || strings.ContainsAny(talents, "\r\n") {
		return "", fmt.Errorf("%w: %q", errInvalidTalents, talents)
	}

	return "talents=" + talents, nil
}

// intField renders one integer assignment segment without a leading comma.
func intField(key string, value int) string {
	return key + "=" + strconv.Itoa(value)
}

// appendOptionalIntField appends key=value only when the pointer is present.
func appendOptionalIntField(fields []string, key string, value *int) []string {
	if value == nil {
		return fields
	}

	return append(fields, intField(key, *value))
}

// appendOptionalIntListField appends key=a/b/c only when a list has at least
// one value. Nil and empty slices both mean "do not emit this SimC property".
func appendOptionalIntListField(fields []string, key string, values *[]int) []string {
	if values == nil || len(*values) == 0 {
		return fields
	}

	parts := make([]string, 0, len(*values))
	for _, value := range *values {
		parts = append(parts, strconv.Itoa(value))
	}

	return append(fields, key+"="+strings.Join(parts, "/"))
}

// isSafeQuotedValue reports whether value can be placed inside SimC double
// quotes without escaping or creating another profile line.
func isSafeQuotedValue(value string) bool {
	return !strings.ContainsAny(value, "\"\\\r\n")
}

// isSafeIdentifier reports whether value is a conservative SimC token made of
// lowercase letters, digits, and underscores.
func isSafeIdentifier(value string) bool {
	if value == "" {
		return false
	}

	for _, char := range value {
		if char >= 'a' && char <= 'z' {
			continue
		}
		if char >= '0' && char <= '9' {
			continue
		}
		if char == '_' {
			continue
		}

		return false
	}

	return true
}
