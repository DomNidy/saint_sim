import {
	CharacterClass,
	type CharacterEquippedItems,
	type CharacterSlotHighWatermark,
	type CharacterTalentLoadout,
	type EquipmentItem,
	EquipmentSlot,
	type EquipmentSource,
	type WowCharacter,
} from "@/lib/saint-api/generated";

type AddonExportParseState = {
	inBagSection: boolean;
	inAdditionalInfo: boolean;
	pendingEquipmentName: string;
	pendingLoadoutName: string;
};

type ParsedEquipmentAssignment = {
	slot: EquipmentSlot;
	attributes: Map<string, string>;
};

type CharacterEquippedSlot = keyof CharacterEquippedItems;

type CatalystCurrency = NonNullable<
	WowCharacter["catalyst_currencies"]
>[number];

const characterClasses = new Set<string>(Object.values(CharacterClass));
const equipmentSlots = new Set<string>(Object.values(EquipmentSlot));
const characterEquippedSlots = new Set<string>([
	EquipmentSlot.HEAD,
	EquipmentSlot.NECK,
	EquipmentSlot.SHOULDER,
	EquipmentSlot.BACK,
	EquipmentSlot.CHEST,
	EquipmentSlot.WRIST,
	EquipmentSlot.HANDS,
	EquipmentSlot.WAIST,
	EquipmentSlot.LEGS,
	EquipmentSlot.FEET,
	EquipmentSlot.FINGER1,
	EquipmentSlot.FINGER2,
	EquipmentSlot.TRINKET1,
	EquipmentSlot.TRINKET2,
	EquipmentSlot.MAIN_HAND,
	EquipmentSlot.OFF_HAND,
]);

function canonicalizeSimcAddonExport(raw: string): string {
	return raw
		.replace(/\r\n?/g, "\n")
		.split("\n")
		.map((line) => line.replace(/[ \t]+$/g, ""))
		.join("\n")
		.replace(/^\n+|\n+$/g, "");
}

/**
 * Parses a raw SimulationCraft addon export into the frontend's generated
 * `WowCharacter` API type.
 *
 * This parser is intentionally best-effort: malformed or absent values are
 * skipped, while required `WowCharacter` fields receive safe empty defaults so
 * callers can decide how strict their own validation should be.
 */
export function parseSimcAddonExport(rawAddonExport: string): WowCharacter {
	rawAddonExport = canonicalizeSimcAddonExport(rawAddonExport);
	const character: WowCharacter = {
		character_class: "" as CharacterClass,
		equipped_items: {} as CharacterEquippedItems,
		level: 0,
		race: "",
		spec: "",
		active_talents: {
			name: undefined,
			talents: "",
		},
	};

	const state: AddonExportParseState = {
		inBagSection: false,
		inAdditionalInfo: false,
		pendingEquipmentName: "",
		pendingLoadoutName: "",
	};

	for (const rawLine of normalizeLineEndings(rawAddonExport).split("\n")) {
		const line = rawLine.trim();
		if (line.length === 0) {
			continue;
		}

		if (line.startsWith("#")) {
			parseCommentLine(character, state, line);
			continue;
		}

		parseAssignmentLine(character, state, line);
	}

	return character;
}

/**
 * Converts CRLF and classic CR line endings to LF before the line-oriented
 * parser inspects the export.
 */
export function normalizeLineEndings(value: string): string {
	return value.replaceAll("\r\n", "\n").replaceAll("\r", "\n");
}

function parseCommentLine(
	character: WowCharacter,
	state: AddonExportParseState,
	line: string,
) {
	const comment = line.replace(/^#+/, "").trim();
	if (comment.length === 0) {
		return;
	}

	if (
		parseSectionComment(state, comment) ||
		parseStructuredComment(character, state, comment) ||
		parseLoadoutComment(character, state, comment)
	) {
		return;
	}

	if (state.inBagSection && looksLikeEquipmentLine(comment)) {
		const item = parseEquipmentItem(state.pendingEquipmentName, comment, "bag");
		if (item !== undefined) {
			character.bag_items = [...(character.bag_items ?? []), item];
		}

		state.pendingEquipmentName = "";
		return;
	}

	if (looksLikeEquipmentNameComment(comment)) {
		state.pendingEquipmentName = comment;
	}
}

function parseAssignmentLine(
	character: WowCharacter,
	state: AddonExportParseState,
	line: string,
) {
	const [key, value] = splitOnce(line, "=");
	if (key === undefined || value === undefined) {
		return;
	}

	const characterClass = parseCharacterClass(key);
	if (characterClass !== undefined) {
		character.character_class = characterClass;
		character.name = trimQuotes(value) || undefined;
		return;
	}

	if (parseMetadataAssignment(character, key, value)) {
		return;
	}

	if (!looksLikeEquipmentLine(line)) {
		return;
	}

	const source: EquipmentSource = state.inBagSection ? "bag" : "equipped";
	const item = parseEquipmentItem(state.pendingEquipmentName, line, source);
	if (item !== undefined) {
		if (source === "bag") {
			character.bag_items = [...(character.bag_items ?? []), item];
		} else if (isCharacterEquippedSlot(item.slot)) {
			character.equipped_items[item.slot] = item;
		}
	}

	state.pendingEquipmentName = "";
}

function isCharacterEquippedSlot(
	slot: EquipmentSlot,
): slot is CharacterEquippedSlot {
	return characterEquippedSlots.has(slot);
}

function parseSectionComment(
	state: AddonExportParseState,
	comment: string,
): boolean {
	switch (comment) {
		case "Gear from Bags":
			state.inBagSection = true;
			state.inAdditionalInfo = false;
			state.pendingEquipmentName = "";
			return true;
		case "Additional Character Info":
			state.inBagSection = false;
			state.inAdditionalInfo = true;
			state.pendingEquipmentName = "";
			return true;
		default:
			return false;
	}
}

function parseStructuredComment(
	character: WowCharacter,
	state: AddonExportParseState,
	comment: string,
): boolean {
	if (
		comment.startsWith("SimC Addon ") ||
		comment.startsWith("WoW ") ||
		comment.startsWith("Requires SimulationCraft ")
	) {
		return true;
	}

	const [key, value] = splitOnce(comment, "=");
	if (key === undefined || value === undefined) {
		return false;
	}

	switch (key) {
		case "loot_spec":
			if (!state.inAdditionalInfo) {
				character.loot_spec = value || undefined;
				return true;
			}
			return false;
		case "catalyst_currencies": {
			const parsed = parseCatalystCurrencies(value);
			if (parsed.length > 0) {
				character.catalyst_currencies = parsed;
			}
			return true;
		}
		case "slot_high_watermarks": {
			const parsed = parseSlotHighWatermarks(value);
			if (parsed.length > 0) {
				character.slot_high_watermarks = parsed;
			}
			return true;
		}
		case "upgrade_achievements":
			character.upgrade_achievements = parseIntegerList(value, "/");
			return true;
		default:
			return false;
	}
}

function parseLoadoutComment(
	character: WowCharacter,
	state: AddonExportParseState,
	comment: string,
): boolean {
	if (comment.startsWith("Saved Loadout:")) {
		state.pendingLoadoutName = comment.replace("Saved Loadout:", "").trim();
		return true;
	}

	if (comment.startsWith("talents=") && state.pendingLoadoutName.length > 0) {
		const loadout: CharacterTalentLoadout = {
			name: state.pendingLoadoutName,
			talents: comment.replace("talents=", "").trim(),
		};
		character.talent_loadouts = [...(character.talent_loadouts ?? []), loadout];
		state.pendingLoadoutName = "";
		return true;
	}

	return false;
}

function parseMetadataAssignment(
	character: WowCharacter,
	key: string,
	value: string,
): boolean {
	switch (key) {
		case "level": {
			const level = parseInteger(value);
			if (level !== undefined) {
				character.level = level;
			}
			return true;
		}
		case "race":
			character.race = value;
			return true;
		case "region":
			character.region = value || undefined;
			return true;
		case "server":
			character.server = value || undefined;
			return true;
		case "role":
			character.role = value || undefined;
			return true;
		case "professions":
			character.professions = value || undefined;
			return true;
		case "spec":
			character.spec = value;
			return true;
		case "talents":
			character.active_talents = {
				name: "Active",
				talents: value,
			};
			return true;
		default:
			return false;
	}
}

function looksLikeEquipmentLine(line: string): boolean {
	const [key, value] = splitOnce(line, "=");
	return (
		key !== undefined && key.length > 0 && value?.startsWith(",id=") === true
	);
}

function looksLikeEquipmentNameComment(comment: string): boolean {
	const openParenIndex = comment.lastIndexOf("(");
	const closeParenIndex = comment.lastIndexOf(")");
	if (
		openParenIndex <= 0 ||
		closeParenIndex !== comment.length - 1 ||
		openParenIndex >= closeParenIndex
	) {
		return false;
	}

	const itemLevelText = comment.slice(openParenIndex + 1, closeParenIndex);
	return /^[0-9]+$/.test(itemLevelText);
}

/**
 * Parses one SimC equipment assignment line, preserving the original raw line
 * while extracting the item fields the API schema exposes.
 */
export function parseEquipmentItem(
	commentName: string,
	rawLine: string,
	source: EquipmentSource,
): EquipmentItem | undefined {
	const assignment = parseEquipmentAssignment(rawLine);
	if (assignment === undefined) {
		return undefined;
	}

	const itemID = parseInteger(assignment.attributes.get("id"));
	if (itemID === undefined) {
		return undefined;
	}

	const commentMetadata = parseItemCommentMetadata(commentName);
	const displayName = commentMetadata.name || `Item ${itemID}`;

	return {
		slot: assignment.slot,
		name: displayName,
		display_name: displayName,
		item_id: itemID,
		item_level:
			commentMetadata.itemLevel ??
			parseOptionalIntegerAttribute(assignment.attributes, "ilevel") ??
			parseOptionalIntegerAttribute(assignment.attributes, "ilvl") ??
			parseOptionalIntegerAttribute(assignment.attributes, "drop_level"),
		enchant_id: parseOptionalIntegerAttribute(
			assignment.attributes,
			"enchant_id",
		),
		crafting_quality: parseOptionalIntegerAttribute(
			assignment.attributes,
			"crafting_quality",
		),
		bonus_ids: parseIntegerListAttribute(
			assignment.attributes,
			"bonus_id",
			"/",
		),
		gem_ids: parseGemIDs(assignment.attributes),
		crafted_stats: parseIntegerListAttribute(
			assignment.attributes,
			"crafted_stats",
			"/",
		),
		source,
		raw_line: rawLine,
	};
}

function parseEquipmentAssignment(
	line: string,
): ParsedEquipmentAssignment | undefined {
	const [rawSlot, rest] = splitOnce(line, "=");
	if (rawSlot === undefined || rawSlot.length === 0 || rest === undefined) {
		return undefined;
	}

	const slot = parseEquipmentSlot(rawSlot);
	if (slot === undefined) {
		return undefined;
	}

	const attributes = new Map<string, string>();
	for (const segment of rest.split(",")) {
		const trimmedSegment = segment.trim();
		if (trimmedSegment.length === 0) {
			continue;
		}

		const [key, value] = splitOnce(trimmedSegment, "=");
		if (key !== undefined && value !== undefined) {
			attributes.set(key, value);
		}
	}

	return { slot, attributes };
}

function parseEquipmentSlot(value: string): EquipmentSlot | undefined {
	if (equipmentSlots.has(value)) {
		return value as EquipmentSlot;
	}

	return undefined;
}

function parseCharacterClass(value: string): CharacterClass | undefined {
	if (characterClasses.has(value)) {
		return value as CharacterClass;
	}

	return undefined;
}

function parseItemCommentMetadata(comment: string): {
	name: string;
	itemLevel?: number;
} {
	const trimmedComment = comment.trim();
	if (trimmedComment.length === 0) {
		return { name: "" };
	}

	const openParenIndex = trimmedComment.lastIndexOf("(");
	const closeParenIndex = trimmedComment.lastIndexOf(")");
	if (
		openParenIndex <= 0 ||
		closeParenIndex !== trimmedComment.length - 1 ||
		openParenIndex >= closeParenIndex
	) {
		return { name: trimmedComment };
	}

	const itemLevel = parseInteger(
		trimmedComment.slice(openParenIndex + 1, closeParenIndex).trim(),
	);
	if (itemLevel === undefined) {
		return { name: trimmedComment };
	}

	const name = trimmedComment.slice(0, openParenIndex).trim();
	return { name: name || trimmedComment, itemLevel };
}

function parseOptionalIntegerAttribute(
	attributes: Map<string, string>,
	key: string,
): number | undefined {
	return parseInteger(attributes.get(key));
}

function parseIntegerListAttribute(
	attributes: Map<string, string>,
	key: string,
	separator: string,
): number[] | undefined {
	const value = attributes.get(key);
	if (value === undefined || value.trim().length === 0) {
		return undefined;
	}

	return parseIntegerList(value, separator);
}

function parseIntegerList(value: string, separator: string): number[] {
	return value
		.split(separator)
		.map((part) => parseInteger(part.trim()))
		.filter((part): part is number => part !== undefined);
}

function parseGemIDs(attributes: Map<string, string>): number[] | undefined {
	const keys = Array.from(attributes.keys())
		.filter((key) => key.startsWith("gem_id"))
		.sort();
	if (keys.length === 0) {
		return undefined;
	}

	const gems = keys
		.map((key) => parseInteger(attributes.get(key)))
		.filter((gem): gem is number => gem !== undefined);

	return gems;
}

function parseCatalystCurrencies(value: string): CatalystCurrency[] {
	return value
		.split("/")
		.map((entry) => {
			const [rawID, rawQuantity] = splitOnce(entry.trim(), ":");
			const id = parseInteger(rawID);
			const quantity = parseInteger(rawQuantity);
			if (id === undefined || quantity === undefined) {
				return undefined;
			}

			return { id, quantity };
		})
		.filter((currency): currency is CatalystCurrency => currency !== undefined)
		.sort((left, right) => left.id - right.id);
}

function parseSlotHighWatermarks(value: string): CharacterSlotHighWatermark[] {
	return value
		.split("/")
		.map((entry) => {
			const [rawSlot, rawCurrentItemLevel, rawMaxItemLevel] = entry
				.trim()
				.split(":");
			const slot =
				rawSlot === undefined ? undefined : parseEquipmentSlot(rawSlot);
			const currentItemLevel = parseInteger(rawCurrentItemLevel);
			const maxItemLevel = parseInteger(rawMaxItemLevel);
			if (
				slot === undefined ||
				currentItemLevel === undefined ||
				maxItemLevel === undefined
			) {
				return undefined;
			}

			return {
				slot,
				current_item_level: currentItemLevel,
				max_item_level: maxItemLevel,
			};
		})
		.filter(
			(watermark): watermark is CharacterSlotHighWatermark =>
				watermark !== undefined,
		);
}

function parseInteger(value: string | undefined): number | undefined {
	if (value === undefined || value.length === 0) {
		return undefined;
	}

	const parsed = Number.parseInt(value, 10);
	return Number.isNaN(parsed) ? undefined : parsed;
}

function splitOnce(
	value: string,
	separator: string,
): [string, string] | [undefined, undefined] {
	const separatorIndex = value.indexOf(separator);
	if (separatorIndex < 0) {
		return [undefined, undefined];
	}

	return [
		value.slice(0, separatorIndex),
		value.slice(separatorIndex + separator.length),
	];
}

function trimQuotes(value: string): string {
	return value.replace(/^"/, "").replace(/"$/, "");
}
