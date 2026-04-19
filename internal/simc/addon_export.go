// Package simc provides various utilities for performing transformations
// on TCI strings.
//
// TCI (Textual Configuration Interface) is the language used to configure
// simc.
//
// TCI Docs:
//   - https://github.com/simulationcraft/simc/wiki/TextualConfigurationInterface
package simc

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strconv"
	"strings"

	api "github.com/DomNidy/saint_sim/internal/api"
	"github.com/DomNidy/saint_sim/internal/utils"
)

type addonExportParseState struct {
	inBagSection         bool
	inAdditionalInfo     bool
	pendingEquipmentName string
	pendingLoadoutName   string
}

const exportHeaderSeparatorCount = 2

// Parse converts a raw SimulationCraft addon export string into a structured
// AddonExport API model.
func Parse(tciString string) api.AddonExport {
	talentLoadouts := []api.AddonExportTalentLoadout{}
	equipment := []api.AddonExportEquipmentItem{}
	catalystCurrencies := map[string]int{}
	slotHighWatermarks := map[string]api.AddonExportSlotHighWatermark{}
	upgradeAchievements := []int{}

	export := api.AddonExport{
		CharacterName:       nil,
		Class:               nil,
		Level:               nil,
		Race:                nil,
		Region:              nil,
		Server:              nil,
		Role:                nil,
		Professions:         nil,
		Spec:                nil,
		TalentLoadouts:      &talentLoadouts,
		Checksum:            nil,
		HeaderComment:       nil,
		SimcAddonComment:    nil,
		WowBuildComment:     nil,
		RequiredSimcComment: nil,
		LootSpec:            nil,
		Equipment:           &equipment,
		CatalystCurrencies:  &catalystCurrencies,
		SlotHighWatermarks:  &slotHighWatermarks,
		UpgradeAchievements: &upgradeAchievements,
	}

	lines := strings.Split(NormalizeLineEndings(tciString), "\n")
	state := addonExportParseState{
		inBagSection:         false,
		inAdditionalInfo:     false,
		pendingEquipmentName: "",
		pendingLoadoutName:   "",
	}

	for _, rawLine := range lines {
		line := strings.Trim(rawLine, " \t")
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "#") {
			parseCommentLine(&export, &state, line)

			continue
		}

		parseAssignmentLine(&export, &state, line)
	}

	return export
}

// HasRecognizedData reports whether Parse extracted any recognizable addon export data.
func HasRecognizedData(export api.AddonExport) bool {
	return hasRecognizedScalarData(export) ||
		hasRecognizedSliceData(export) ||
		hasRecognizedMapData(export)
}

func hasRecognizedScalarData(export api.AddonExport) bool {
	return export.CharacterName != nil ||
		export.Class != nil ||
		export.Level != nil ||
		export.Race != nil ||
		export.Region != nil ||
		export.Server != nil ||
		export.Role != nil ||
		hasRecognizedSupplementaryScalarData(export)
}

func hasRecognizedSupplementaryScalarData(export api.AddonExport) bool {
	return export.Professions != nil ||
		export.Spec != nil ||
		export.Checksum != nil ||
		export.HeaderComment != nil ||
		export.SimcAddonComment != nil ||
		export.WowBuildComment != nil ||
		export.RequiredSimcComment != nil ||
		export.LootSpec != nil
}

func hasRecognizedSliceData(export api.AddonExport) bool {
	return export.TalentLoadouts != nil && len(*export.TalentLoadouts) > 0 ||
		export.Equipment != nil && len(*export.Equipment) > 0 ||
		export.UpgradeAchievements != nil && len(*export.UpgradeAchievements) > 0
}

func hasRecognizedMapData(export api.AddonExport) bool {
	return export.CatalystCurrencies != nil && len(*export.CatalystCurrencies) > 0 ||
		export.SlotHighWatermarks != nil && len(*export.SlotHighWatermarks) > 0
}

func parseCommentLine(
	export *api.AddonExport,
	state *addonExportParseState,
	line string,
) {
	comment := strings.TrimSpace(strings.TrimLeft(line, "#"))
	if comment == "" {
		return
	}

	if parseSectionComment(state, comment) {
		return
	}

	if parseStructuredComment(export, state, comment) {
		return
	}

	if parseLoadoutComment(export, state, comment) {
		return
	}

	if parseChecksumComment(export, comment) {
		return
	}

	if state.inBagSection && looksLikeEquipmentLine(comment) {
		if item, ok := ParseEquipmentItem(state.pendingEquipmentName, comment, api.Bag); ok {
			*export.Equipment = append(*export.Equipment, item)
		}

		state.pendingEquipmentName = ""

		return
	}

	if looksLikeEquipmentNameComment(comment) {
		state.pendingEquipmentName = comment

		return
	}

	parseHeaderComment(export, comment)
}

func parseAssignmentLine(
	export *api.AddonExport,
	state *addonExportParseState,
	line string,
) {
	key, value, ok := strings.Cut(line, "=")
	if !ok {
		return
	}

	if isClassIdentifier(key) {
		classValue := string(tciClassIdentifier(key))
		export.Class = &classValue

		characterName := strings.Trim(value, "\"")
		if characterName != "" {
			export.CharacterName = &characterName
		}

		return
	}

	if parseMetadataAssignment(export, key, value) {
		return
	}

	if looksLikeEquipmentLine(line) {
		if item, itemOK := ParseEquipmentItem(
			state.pendingEquipmentName,
			line,
			api.Equipped, // api.Equipped source bc this is an assignment line
		); itemOK {
			*export.Equipment = append(*export.Equipment, item)
		}

		state.pendingEquipmentName = ""
	}
}

func parseSectionComment(state *addonExportParseState, comment string) bool {
	switch comment {
	case "Gear from Bags":
		state.inBagSection = true
		state.inAdditionalInfo = false
		state.pendingEquipmentName = ""

		return true
	case "Additional Character Info":
		state.inBagSection = false
		state.inAdditionalInfo = true
		state.pendingEquipmentName = ""

		return true
	default:
		return false
	}
}

func parseStructuredComment(
	export *api.AddonExport,
	state *addonExportParseState,
	comment string,
) bool {
	if strings.HasPrefix(comment, "SimC Addon ") {
		export.SimcAddonComment = strPtr(comment)

		return true
	}

	if strings.HasPrefix(comment, "WoW ") {
		export.WowBuildComment = strPtr(comment)

		return true
	}

	if strings.HasPrefix(comment, "Requires SimulationCraft ") {
		export.RequiredSimcComment = strPtr(comment)

		return true
	}

	key, value, ok := strings.Cut(comment, "=")
	if !ok {
		return false
	}

	switch key {
	case "loot_spec":
		if !state.inAdditionalInfo {
			export.LootSpec = strPtr(value)

			return true
		}
	case "catalyst_currencies":
		parsed := parseIntMap(value)
		export.CatalystCurrencies = &parsed

		return true
	case "slot_high_watermarks":
		parsed := parseSlotHighWatermarks(value)
		export.SlotHighWatermarks = &parsed

		return true
	case "upgrade_achievements":
		parsed := parseIntListValue(value, "/")
		export.UpgradeAchievements = &parsed

		return true
	}

	return false
}

func parseLoadoutComment(
	export *api.AddonExport,
	state *addonExportParseState,
	comment string,
) bool {
	if strings.HasPrefix(comment, "Saved Loadout:") {
		state.pendingLoadoutName = strings.TrimSpace(
			strings.TrimPrefix(comment, "Saved Loadout:"),
		)

		return true
	}

	// Named saved loadouts are emitted as comments; unnamed talent assignments are
	// handled in parseMetadataAssignment as the active loadout.
	if strings.HasPrefix(comment, "talents=") && state.pendingLoadoutName != "" {
		*export.TalentLoadouts = append(
			*export.TalentLoadouts,
			api.AddonExportTalentLoadout{
				Name:    utils.StrPtr(state.pendingLoadoutName),
				Talents: strings.TrimSpace(strings.TrimPrefix(comment, "talents=")),
			},
		)
		state.pendingLoadoutName = ""

		return true
	}

	return false
}

func parseChecksumComment(export *api.AddonExport, comment string) bool {
	if !strings.HasPrefix(comment, "Checksum:") {
		return false
	}

	export.Checksum = strPtr(
		strings.TrimSpace(strings.TrimPrefix(comment, "Checksum:")),
	)

	return true
}

func parseHeaderComment(export *api.AddonExport, comment string) {
	if export.HeaderComment == nil && looksLikeExportHeaderComment(comment) {
		export.HeaderComment = strPtr(comment)
	}
}

func appendActiveTalentLoadout(export *api.AddonExport, value string) {
	if export.TalentLoadouts == nil {
		return
	}

	*export.TalentLoadouts = append(*export.TalentLoadouts, api.AddonExportTalentLoadout{
		Name:    utils.StrPtr("Active"),
		Talents: value,
	})
}

func parseMetadataAssignment(
	export *api.AddonExport,
	key string,
	value string,
) bool {
	switch key {
	case "level":
		export.Level = strPtr(value)
	case "race":
		export.Race = strPtr(value)
	case "region":
		export.Region = strPtr(value)
	case "server":
		export.Server = strPtr(value)
	case "role":
		export.Role = strPtr(value)
	case "professions":
		export.Professions = strPtr(value)
	case "spec":
		export.Spec = strPtr(value)
	case "talents":
		appendActiveTalentLoadout(export, value)
	default:
		return false
	}

	return true
}

func isClassIdentifier(value string) bool {
	switch tciClassIdentifier(value) {
	case Warrior,
		Hunter,
		Monk,
		Paladin,
		Rogue,
		Shaman,
		Mage,
		Warlock,
		Druid,
		DeathKnight,
		Priest,
		DemonHunter,
		Evoker:
		return true
	default:
		return false
	}
}

func looksLikeEquipmentLine(line string) bool {
	key, value, ok := strings.Cut(line, "=")
	if !ok || key == "" {
		return false
	}

	return strings.HasPrefix(value, ",id=")
}

func looksLikeEquipmentNameComment(comment string) bool {
	openParen := strings.LastIndexByte(comment, '(')
	closeParen := strings.LastIndexByte(comment, ')')

	if openParen <= 0 || closeParen != len(comment)-1 || openParen >= closeParen {
		return false
	}

	levelText := comment[openParen+1 : closeParen]
	if levelText == "" {
		return false
	}

	for _, r := range levelText {
		if r < '0' || r > '9' {
			return false
		}
	}

	return true
}

func looksLikeExportHeaderComment(comment string) bool {
	return strings.Count(comment, " - ") >= exportHeaderSeparatorCount
}

// ParseEquipmentItem takes a raw TCI line (and optional preceding comment to
// attempt to extract item name from), and parses it into a structured item
// representation.
func ParseEquipmentItem(
	commentName string,
	rawLine string,
	source api.AddonExportEquipmentSource,
) (api.AddonExportEquipmentItem, bool) {
	slot, attributes, foundAssignment := parseEquipmentAssignment(rawLine)
	if !foundAssignment {
		return emptyEquipmentItem(), false
	}

	itemID, foundItemID := parseIntAttribute(attributes, "id")
	if !foundItemID {
		return emptyEquipmentItem(), false
	}

	displayName, commentItemLevel := parseItemCommentMetadata(commentName)
	itemLevel := firstNonNil(
		commentItemLevel,
		parseOptionalIntAttribute(attributes, "ilevel"),
		parseOptionalIntAttribute(attributes, "ilvl"),
		parseOptionalIntAttribute(attributes, "drop_level"),
	)

	if displayName == "" {
		displayName = fmt.Sprintf("Item %d", itemID)
	}

	return api.AddonExportEquipmentItem{
		Fingerprint:     fingerprintForItem(rawLine, source),
		Slot:            slot,
		Name:            displayName,
		DisplayName:     displayName,
		ItemId:          itemID,
		ItemLevel:       itemLevel,
		EnchantId:       parseOptionalIntAttribute(attributes, "enchant_id"),
		CraftingQuality: parseOptionalIntAttribute(attributes, "crafting_quality"),
		BonusIds:        intSlicePtr(parseIntListAttribute(attributes, "bonus_id", "/")),
		GemIds:          intSlicePtr(parseGemIDs(attributes)),
		CraftedStats:    intSlicePtr(parseIntListAttribute(attributes, "crafted_stats", "/")),
		Source:          source,
		RawLine:         rawLine,
	}, true
}

func emptyEquipmentItem() api.AddonExportEquipmentItem {
	return api.AddonExportEquipmentItem{
		Fingerprint:     "",
		Slot:            "",
		Name:            "",
		DisplayName:     "",
		ItemId:          0,
		ItemLevel:       nil,
		EnchantId:       nil,
		CraftingQuality: nil,
		BonusIds:        nil,
		GemIds:          nil,
		CraftedStats:    nil,
		Source:          "",
		RawLine:         "",
	}
}

func parseEquipmentAssignment(line string) (string, map[string]string, bool) {
	slot, rest, ok := strings.Cut(line, "=")
	if !ok || slot == "" {
		return "", nil, false
	}

	attributes := map[string]string{}
	for _, segment := range strings.Split(rest, ",") {
		segment = strings.TrimSpace(segment)
		if segment == "" {
			continue
		}

		key, value, keyOK := strings.Cut(segment, "=")
		if !keyOK {
			continue
		}

		attributes[key] = value
	}

	return slot, attributes, true
}

func parseItemCommentMetadata(comment string) (string, *int) {
	comment = strings.TrimSpace(comment)
	if comment == "" {
		return "", nil
	}

	openIdx := strings.LastIndex(comment, "(")
	closeIdx := strings.LastIndex(comment, ")")
	if openIdx <= 0 || closeIdx != len(comment)-1 || openIdx >= closeIdx {
		return comment, nil
	}

	levelText := strings.TrimSpace(comment[openIdx+1 : closeIdx])
	level, err := strconv.Atoi(levelText)
	if err != nil {
		return comment, nil
	}

	name := strings.TrimSpace(comment[:openIdx])
	if name == "" {
		name = comment
	}

	return name, &level
}

func parseIntAttribute(attributes map[string]string, key string) (int, bool) {
	value, ok := attributes[key]
	if !ok || value == "" {
		return 0, false
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, false
	}

	return parsed, true
}

func parseOptionalIntAttribute(attributes map[string]string, key string) *int {
	value, ok := parseIntAttribute(attributes, key)
	if !ok {
		return nil
	}

	return &value
}

func parseIntListAttribute(attributes map[string]string, key, separator string) []int {
	value, ok := attributes[key]
	if !ok || strings.TrimSpace(value) == "" {
		return nil
	}

	return parseIntListValue(value, separator)
}

func parseIntListValue(value, separator string) []int {
	parts := strings.Split(value, separator)
	result := make([]int, 0, len(parts))
	for _, part := range parts {
		parsed, err := strconv.Atoi(strings.TrimSpace(part))
		if err != nil {
			continue
		}
		result = append(result, parsed)
	}

	return result
}

func parseGemIDs(attributes map[string]string) []int {
	keys := make([]string, 0)
	for key := range attributes {
		if strings.HasPrefix(key, "gem_id") {
			keys = append(keys, key)
		}
	}
	if len(keys) == 0 {
		return nil
	}
	sort.Strings(keys)

	gems := make([]int, 0, len(keys))
	for _, key := range keys {
		parsed, err := strconv.Atoi(attributes[key])
		if err == nil {
			gems = append(gems, parsed)
		}
	}

	return gems
}

func intSlicePtr(values []int) *[]int {
	if values == nil {
		return nil
	}

	return &values
}

func firstNonNil(values ...*int) *int {
	for _, value := range values {
		if value != nil {
			return value
		}
	}

	return nil
}

func fingerprintForItem(rawLine string, source api.AddonExportEquipmentSource) string {
	normalized := strings.TrimSpace(strings.ToLower(rawLine)) + "|" + string(source)
	hash := sha256.Sum256([]byte(normalized))

	return hex.EncodeToString(hash[:])
}

func parseIntMap(value string) map[string]int {
	result := map[string]int{}
	for _, entry := range strings.Split(value, "/") {
		key, rawValue, ok := strings.Cut(strings.TrimSpace(entry), ":")
		if !ok || key == "" {
			continue
		}

		parsed, err := strconv.Atoi(rawValue)
		if err != nil {
			continue
		}

		result[key] = parsed
	}

	return result
}

func parseSlotHighWatermarks(
	value string,
) map[string]api.AddonExportSlotHighWatermark {
	result := map[string]api.AddonExportSlotHighWatermark{}
	for _, entry := range strings.Split(value, "/") {
		parts := strings.Split(strings.TrimSpace(entry), ":")
		if len(parts) != 3 || parts[0] == "" {
			continue
		}

		currentItemLevel, currentErr := strconv.Atoi(parts[1])
		maxItemLevel, maxErr := strconv.Atoi(parts[2])
		if currentErr != nil || maxErr != nil {
			continue
		}

		result[parts[0]] = api.AddonExportSlotHighWatermark{
			CurrentItemLevel: currentItemLevel,
			MaxItemLevel:     maxItemLevel,
		}
	}

	return result
}

func strPtr(value string) *string {
	if value == "" {
		return nil
	}

	return &value
}
