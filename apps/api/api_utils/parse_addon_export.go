package api_utils

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	api "github.com/DomNidy/saint_sim/internal/api"
)

type addonExportParseState struct {
	inBagSection         bool
	inAdditionalInfo     bool
	pendingEquipmentName string
	pendingLoadoutName   string
}

// Parse converts a raw SimulationCraft addon export string into a structured
// WoW character API model.
func ParseAddonExport(tciString string) api.WowCharacter {
	export := api.WowCharacter{
		EquippedItems: []api.EquipmentItem{},
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

func parseCommentLine(
	export *api.WowCharacter,
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

	if state.inBagSection && looksLikeEquipmentLine(comment) {
		if item, ok := ParseEquipmentItem(state.pendingEquipmentName, comment, api.Bag); ok {
			appendBagItem(export, item)
		}

		state.pendingEquipmentName = ""

		return
	}

	if looksLikeEquipmentNameComment(comment) {
		state.pendingEquipmentName = comment

		return
	}
}

func parseAssignmentLine(
	export *api.WowCharacter,
	state *addonExportParseState,
	line string,
) {
	key, value, ok := strings.Cut(line, "=")
	if !ok {
		return
	}

	if classValue, ok := parseClassIdentifier(key); ok {
		export.CharacterClass = classValue

		return
	}

	if parseMetadataAssignment(export, key, value) {
		return
	}

	if looksLikeEquipmentLine(line) {
		source := api.Equipped
		if state.inBagSection {
			source = api.Bag
		}

		if item, itemOK := ParseEquipmentItem(
			state.pendingEquipmentName,
			line,
			source,
		); itemOK {
			if source == api.Bag {
				appendBagItem(export, item)
			} else {
				export.EquippedItems = append(export.EquippedItems, item)
			}
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
	export *api.WowCharacter,
	state *addonExportParseState,
	comment string,
) bool {
	if strings.HasPrefix(comment, "SimC Addon ") {
		return true
	}

	if strings.HasPrefix(comment, "WoW ") {
		return true
	}

	if strings.HasPrefix(comment, "Requires SimulationCraft ") {
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
		parsed := parseCatalystCurrencies(value)
		if len(parsed) > 0 {
			export.CatalystCurrencies = &parsed
		}

		return true
	case "slot_high_watermarks":
		parsed := parseSlotHighWatermarks(value)
		if len(parsed) > 0 {
			export.SlotHighWatermarks = &parsed
		}

		return true
	case "upgrade_achievements":
		parsed := parseIntListValue(value, "/")
		export.UpgradeAchievements = &parsed

		return true
	}

	return false
}

func parseLoadoutComment(
	export *api.WowCharacter,
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
		appendTalentLoadout(export, api.CharacterTalentLoadout{
			Name:    strPtr(state.pendingLoadoutName),
			Talents: strings.TrimSpace(strings.TrimPrefix(comment, "talents=")),
		})
		state.pendingLoadoutName = ""

		return true
	}

	return false
}

func setActiveTalents(export *api.WowCharacter, value string) {
	export.ActiveTalents = &api.CharacterTalentLoadout{
		Name:    strPtr("Active"),
		Talents: value,
	}
}

func appendTalentLoadout(export *api.WowCharacter, loadout api.CharacterTalentLoadout) {
	if export.TalentLoadouts == nil {
		loadouts := []api.CharacterTalentLoadout{}
		export.TalentLoadouts = &loadouts
	}

	*export.TalentLoadouts = append(*export.TalentLoadouts, loadout)
}

func appendBagItem(export *api.WowCharacter, item api.EquipmentItem) {
	if export.BagItems == nil {
		bagItems := []api.EquipmentItem{}
		export.BagItems = &bagItems
	}

	*export.BagItems = append(*export.BagItems, item)
}

func parseMetadataAssignment(
	export *api.WowCharacter,
	key string,
	value string,
) bool {
	switch key {
	case "level":
		if level, err := strconv.Atoi(value); err == nil {
			export.Level = level
		}
	case "race":
		export.Race = value
	case "region":
		export.Region = strPtr(value)
	case "server":
		export.Server = strPtr(value)
	case "role":
		export.Role = strPtr(value)
	case "professions":
		export.Professions = strPtr(value)
	case "spec":
		export.Spec = value
	case "talents":
		setActiveTalents(export, value)
	default:
		return false
	}

	return true
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

// ParseEquipmentItem takes a raw TCI line (and optional preceding comment to
// attempt to extract item name from), and parses it into a structured item
// representation.
func ParseEquipmentItem(
	commentName string,
	rawLine string,
	source api.EquipmentSource,
) (api.EquipmentItem, bool) {
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

	return api.EquipmentItem{
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

func emptyEquipmentItem() api.EquipmentItem {
	return api.EquipmentItem{
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

func parseEquipmentAssignment(line string) (api.EquipmentSlot, map[string]string, bool) {
	rawSlot, rest, foundAssignment := strings.Cut(line, "=")
	if !foundAssignment || rawSlot == "" {
		return "", nil, false
	}

	slot, recognizedSlot := parseEquipmentSlot(rawSlot)
	if !recognizedSlot {
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

func parseEquipmentSlot(value string) (api.EquipmentSlot, bool) {
	switch api.EquipmentSlot(value) {
	case api.Back,
		api.Chest,
		api.Feet,
		api.Finger1,
		api.Finger2,
		api.Hands,
		api.Head,
		api.Legs,
		api.MainHand,
		api.Neck,
		api.OffHand,
		api.Shirt,
		api.Shoulder,
		api.Tabard,
		api.Trinket1,
		api.Trinket2,
		api.Waist,
		api.Wrist:
		return api.EquipmentSlot(value), true
	default:
		return "", false
	}
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

func parseCatalystCurrencies(value string) []struct {
	Id       int `json:"id"`
	Quantity int `json:"quantity"`
} {
	result := []struct {
		Id       int `json:"id"`
		Quantity int `json:"quantity"`
	}{}
	for _, entry := range strings.Split(value, "/") {
		key, rawValue, ok := strings.Cut(strings.TrimSpace(entry), ":")
		if !ok || key == "" {
			continue
		}

		id, idErr := strconv.Atoi(key)
		quantity, quantityErr := strconv.Atoi(rawValue)
		if idErr != nil || quantityErr != nil {
			continue
		}

		result = append(result, struct {
			Id       int `json:"id"`
			Quantity int `json:"quantity"`
		}{
			Id:       id,
			Quantity: quantity,
		})
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Id < result[j].Id
	})

	return result
}

func parseSlotHighWatermarks(
	value string,
) []api.CharacterSlotHighWatermark {
	result := []api.CharacterSlotHighWatermark{}
	for _, entry := range strings.Split(value, "/") {
		parts := strings.Split(strings.TrimSpace(entry), ":")
		if len(parts) != 3 || parts[0] == "" {
			continue
		}

		slot, slotOK := parseEquipmentSlot(parts[0])
		if !slotOK {
			continue
		}

		currentItemLevel, currentErr := strconv.Atoi(parts[1])
		maxItemLevel, maxErr := strconv.Atoi(parts[2])
		if currentErr != nil || maxErr != nil {
			continue
		}

		result = append(result, api.CharacterSlotHighWatermark{
			CurrentItemLevel: currentItemLevel,
			MaxItemLevel:     maxItemLevel,
			Slot:             slot,
		})
	}

	return result
}

func strPtr(value string) *string {
	if value == "" {
		return nil
	}

	return &value
}

// NormalizeLineEndings converts CRLF (\r\n) and CR (\r) line endings to LF (\n).
func NormalizeLineEndings(tciString string) string {
	// Normalize CRLF first so the subsequent CR replacement does not
	// introduce double newlines.
	return strings.ReplaceAll(
		strings.ReplaceAll(tciString, "\r\n", "\n"),
		"\r",
		"\n",
	)
}

// StripAllComments returns a new string where each commented line in the
// TCI profile is removed.
//
// TCI comments are lines that begin with a "#".
// TCI does not recognize trailing comments, so we don't try to remove them.
func StripAllComments(tciString string) string {
	tciString = NormalizeLineEndings(tciString)
	lines := strings.Split(tciString, "\n")

	filteredLines := make([]string, 0, len(lines))

	for _, line := range lines {
		if !strings.HasPrefix(line, "#") {
			filteredLines = append(filteredLines, line)
		}
	}

	return strings.Join(filteredLines, "\n")
}

// TrimLineWhitespace trims leading and trailing spaces and tabs from each
// line in the TCI string.
func TrimLineWhitespace(tciString string) string {
	tciString = NormalizeLineEndings(tciString)
	lines := strings.Split(tciString, "\n")

	trimmedLines := make([]string, 0, len(lines))

	for _, line := range lines {
		trimmedLines = append(trimmedLines, strings.Trim(line, " \t"))
	}

	return strings.Join(trimmedLines, "\n")
}

func parseClassIdentifier(value string) (api.CharacterClass, bool) {
	switch api.CharacterClass(value) {
	case api.Warrior,
		api.Hunter,
		api.Monk,
		api.Paladin,
		api.Rogue,
		api.Shaman,
		api.Mage,
		api.Warlock,
		api.Druid,
		api.Deathknight,
		api.Priest,
		api.Demonhunter,
		api.Evoker:
		return api.CharacterClass(value), true
	default:
		return "", false
	}
}
