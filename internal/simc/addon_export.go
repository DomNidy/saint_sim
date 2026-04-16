// Package simc provides various utilities for performing transformations
// on TCI strings.
//
// TCI (Textual Configuration Interface) is the language used to configure
// simc.
//
// TCI Docs:
//   - https://github.com/simulationcraft/simc/wiki/TextualConfigurationInterface
package simc

import "strings"

// AddonExport is a TCI string that is produced when a player
// uses the Simc WoW addon to export their character as a TCI profile.
//
// It adheres to TCI syntax as usual, but includes additional comments
// containing metadata, such as:
// - Character name,
// - Simc Addon version,
// - WoW client version,
// - required SimulationCraft version
// - "Gear from Bags" (items in character's bags)
// - "Addition Character Info" catalyst charges, upgrade currencies, etc.
// - Checksum hash of the profile.
type AddonExport struct {
	class       tciClassIdentifier // character class
	level       string             // character level
	race        string             // character race identifier
	region      string             // character region, e.g. us
	server      string             // character server/realm name
	role        string             // character combat role
	professions string             // raw professions export value
	spec        string             // character specialization identifier

	activeTalents          string
	alternateTalentLoadout []alternateTalentLoadout

	equipmentItems []equipmentItem

	checksum string
}

type equipmentItem struct {
	name      string // name of item parsed from preceding comment
	equipment string // raw line
	bagItem   bool   // is this a bag item? (i.e. not active)
}

type alternateTalentLoadout struct {
	name    string // name of loadout parsed from preceding comment
	talents string
}

type addonExportParseState struct {
	inBagSection         bool
	pendingEquipmentName string
	pendingLoadoutName   string
}

// Parse converts a raw SimulationCraft addon export string into a structured
// SimcAddonExport.
func Parse(tciString string) AddonExport {
	var export AddonExport

	lines := strings.Split(NormalizeLineEndings(tciString), "\n")
	state := addonExportParseState{
		inBagSection:         false,
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
	export *AddonExport,
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

	if parseLoadoutComment(export, state, comment) {
		return
	}

	if parseChecksumComment(export, comment) {
		return
	}

	if state.inBagSection && looksLikeEquipmentLine(comment) {
		export.equipmentItems = append(
			export.equipmentItems,
			equipmentItem{
				name:      state.pendingEquipmentName,
				equipment: comment,
				bagItem:   true,
			},
		)
		state.pendingEquipmentName = ""

		return
	}

	if looksLikeEquipmentNameComment(comment) {
		state.pendingEquipmentName = comment
	}
}

func parseAssignmentLine(
	export *AddonExport,
	state *addonExportParseState,
	line string,
) {
	key, value, ok := strings.Cut(line, "=")
	if !ok {
		return
	}

	if isClassIdentifier(key) {
		export.class = tciClassIdentifier(key)

		return
	}

	if parseMetadataAssignment(export, key, value) {
		return
	}

	if looksLikeEquipmentLine(line) {
		export.equipmentItems = append(
			export.equipmentItems,
			equipmentItem{
				name:      state.pendingEquipmentName,
				equipment: line,
				bagItem:   false,
			},
		)
		state.pendingEquipmentName = ""
	}
}

func parseSectionComment(state *addonExportParseState, comment string) bool {
	switch comment {
	case "Gear from Bags":
		state.inBagSection = true
		state.pendingEquipmentName = ""

		return true
	case "Additional Character Info":
		state.inBagSection = false
		state.pendingEquipmentName = ""

		return true
	default:
		return false
	}
}

func parseLoadoutComment(
	export *AddonExport,
	state *addonExportParseState,
	comment string,
) bool {
	if strings.HasPrefix(comment, "Saved Loadout:") {
		state.pendingLoadoutName = strings.TrimSpace(
			strings.TrimPrefix(comment, "Saved Loadout:"),
		)

		return true
	}

	if strings.HasPrefix(comment, "talents=") && state.pendingLoadoutName != "" {
		export.alternateTalentLoadout = append(
			export.alternateTalentLoadout,
			alternateTalentLoadout{
				name:    state.pendingLoadoutName,
				talents: strings.TrimSpace(strings.TrimPrefix(comment, "talents=")),
			},
		)
		state.pendingLoadoutName = ""

		return true
	}

	return false
}

func parseChecksumComment(export *AddonExport, comment string) bool {
	if !strings.HasPrefix(comment, "Checksum:") {
		return false
	}

	export.checksum = strings.TrimSpace(
		strings.TrimPrefix(comment, "Checksum:"),
	)

	return true
}

func parseMetadataAssignment(
	export *AddonExport,
	key string,
	value string,
) bool {
	switch key {
	case "level":
		export.level = value
	case "race":
		export.race = value
	case "region":
		export.region = value
	case "server":
		export.server = value
	case "role":
		export.role = value
	case "professions":
		export.professions = value
	case "spec":
		export.spec = value
	case "talents":
		export.activeTalents = value
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
