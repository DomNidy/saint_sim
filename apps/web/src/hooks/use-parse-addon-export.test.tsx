// @vitest-environment jsdom

import { render } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { useParseAddonExport } from "./use-parse-addon-export";

describe("useParseAddonExport", () => {
	it("returns no parsed data when disabled", () => {
		let result: ReturnType<typeof useParseAddonExport> | undefined;

		const TestComponent = () => {
			result = useParseAddonExport('priest="Example"', false);
			return <p>disabled</p>;
		};

		render(<TestComponent />);

		expect(result).toEqual({
			equipmentItems: [],
			wowCharacter: undefined,
		});
	});

	it("parses addon export data into a wow character and grouped equipment", () => {
		let result: ReturnType<typeof useParseAddonExport> | undefined;

		const TestComponent = () => {
			result = useParseAddonExport(
				`priest="Example"
level=80
race=void_elf
spec=shadow
# Host Commander's Casque (684)
head=,id=250458,ilevel=684`,
				true,
			);
			return <p>parsed</p>;
		};

		render(<TestComponent />);

		expect(result?.wowCharacter).toMatchObject({
			name: "Example",
			character_class: "priest",
			level: 80,
			race: "void_elf",
			spec: "shadow",
		});
		expect(result?.equipmentItems).toHaveLength(1);
		expect(result?.equipmentItems[0]?.groupLabel).toBe("head");
	});

	it("assigns stable unique selection IDs to parsed equipment wrappers", () => {
		let result: ReturnType<typeof useParseAddonExport> | undefined;

		const TestComponent = () => {
			result = useParseAddonExport(
				`# Equipped Helm (684)
head=,id=250458,bonus_id=6652,ilevel=684
# Gear from Bags
# Bag Helm A (684)
head=,id=250458,bonus_id=6652,ilevel=684
# Bag Helm B (684)
head=,id=250458,bonus_id=6652,ilevel=684`,
				true,
			);
			return <p>parsed</p>;
		};

		render(<TestComponent />);

		const parsedEquipment = result?.equipmentItems;

		expect(parsedEquipment?.map((item) => item.selectionId)).toEqual([
			"addon-export:0",
			"addon-export:1",
			"addon-export:2",
		]);
		expect(
			new Set(parsedEquipment?.map((item) => item.selectionId)),
		).toHaveProperty("size", 3);
		expect(parsedEquipment?.map((item) => item.item.raw_line)).toEqual([
			"head=,id=250458,bonus_id=6652,ilevel=684",
			"head=,id=250458,bonus_id=6652,ilevel=684",
			"head=,id=250458,bonus_id=6652,ilevel=684",
		]);
		expect(result?.wowCharacter?.equipped_items.head).not.toHaveProperty(
			"selectionId",
		);
	});
});
