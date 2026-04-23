import { describe, expect, it } from "vitest";
import { EquipmentSlot, EquipmentSource } from "@/lib/saint-api/generated";
import {
	normalizeLineEndings,
	parseEquipmentItem,
	parseSimcAddonExport,
} from "./parse-addon-export";

describe("parseSimcAddonExport", () => {
	it("parses character metadata and improves on Go by preserving character name", () => {
		const got = parseSimcAddonExport(
			'priest="Example"\nlevel=80\nrace=void_elf\nregion=us\nserver=area-52\nrole=spell\nprofessions=alchemy=100/herbalism=100\nspec=shadow',
		);

		expect(got).toMatchObject({
			name: "Example",
			character_class: "priest",
			level: 80,
			race: "void_elf",
			region: "us",
			server: "area-52",
			role: "spell",
			professions: "alchemy=100/herbalism=100",
			spec: "shadow",
			equipped_items: [],
		});
	});

	it("splits active talents from non-active saved loadouts", () => {
		const got = parseSimcAddonExport(`# Saved Loadout: Dungeon
# talents=DUNGEON_TALENTS
talents=ACTIVE_TALENTS`);

		expect(got.active_talents).toEqual({
			name: "Active",
			talents: "ACTIVE_TALENTS",
		});
		expect(got.talent_loadouts).toEqual([
			{
				name: "Dungeon",
				talents: "DUNGEON_TALENTS",
			},
		]);
	});

	it("parses equipped and bag item groups from addon export sections", () => {
		const got = parseSimcAddonExport(`# Host Commander's Casque (684)
head=,id=250458,bonus_id=6652/12667,gem_id=240865,ilevel=684
# Gear from Bags
# Gnarlroot Spinecleaver (710)
main_hand=,id=249671,enchant_id=3368,bonus_id=6652,ilevel=710`);

		expect(got.equipped_items).toHaveLength(1);
		expect(got.equipped_items[0]).toMatchObject({
			slot: EquipmentSlot.HEAD,
			name: "Host Commander's Casque",
			display_name: "Host Commander's Casque",
			item_id: 250458,
			item_level: 684,
			bonus_ids: [6652, 12667],
			gem_ids: [240865],
			source: EquipmentSource.EQUIPPED,
			raw_line: "head=,id=250458,bonus_id=6652/12667,gem_id=240865,ilevel=684",
		});

		expect(got.bag_items).toHaveLength(1);
		expect(got.bag_items?.[0]).toMatchObject({
			slot: EquipmentSlot.MAIN_HAND,
			name: "Gnarlroot Spinecleaver",
			item_id: 249671,
			item_level: 710,
			enchant_id: 3368,
			source: EquipmentSource.BAG,
		});
	});

	it("parses Additional Character Info comments", () => {
		const got = parseSimcAddonExport(`# loot_spec=shadow
# Additional Character Info
# catalyst_currencies=3116:2/2912:1
# slot_high_watermarks=head:684:710/main_hand:710:720
# upgrade_achievements=1/2`);

		expect(got.loot_spec).toBe("shadow");
		expect(got.catalyst_currencies).toEqual([
			{ id: 2912, quantity: 1 },
			{ id: 3116, quantity: 2 },
		]);
		expect(got.slot_high_watermarks).toEqual([
			{
				slot: EquipmentSlot.HEAD,
				current_item_level: 684,
				max_item_level: 710,
			},
			{
				slot: EquipmentSlot.MAIN_HAND,
				current_item_level: 710,
				max_item_level: 720,
			},
		]);
		expect(got.upgrade_achievements).toEqual([1, 2]);
	});

	it("normalizes line endings before parsing", () => {
		expect(
			normalizeLineEndings('priest="Example"\r\nlevel=80\rspec=shadow'),
		).toBe('priest="Example"\nlevel=80\nspec=shadow');

		const got = parseSimcAddonExport(
			'priest="Example"\r\nlevel=80\rspec=shadow',
		);

		expect(got.level).toBe(80);
		expect(got.spec).toBe("shadow");
	});
});

describe("parseEquipmentItem", () => {
	it("parses optional equipment attributes and keeps raw line intact", () => {
		const rawLine =
			"head=,id=250458,enchant_id=3368,crafting_quality=5,bonus_id=6652/12667,gem_id=240865,gem_id2=240866,crafted_stats=36/40,drop_level=684";

		const got = parseEquipmentItem(
			"Host Commander's Casque (700)",
			rawLine,
			EquipmentSource.EQUIPPED,
		);

		expect(got).toMatchObject({
			slot: EquipmentSlot.HEAD,
			name: "Host Commander's Casque",
			display_name: "Host Commander's Casque",
			item_id: 250458,
			item_level: 700,
			enchant_id: 3368,
			crafting_quality: 5,
			bonus_ids: [6652, 12667],
			gem_ids: [240865, 240866],
			crafted_stats: [36, 40],
			source: EquipmentSource.EQUIPPED,
			raw_line: rawLine,
		});
	});

	it("returns undefined for unrecognized slots or missing item ids", () => {
		expect(
			parseEquipmentItem("", "unknown=,id=1", EquipmentSource.EQUIPPED),
		).toBeUndefined();
		expect(
			parseEquipmentItem("", "head=,bonus_id=6652", EquipmentSource.EQUIPPED),
		).toBeUndefined();
	});
});
