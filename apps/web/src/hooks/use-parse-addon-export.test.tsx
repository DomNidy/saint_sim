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
			equipmentGroups: [],
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
		expect(result?.equipmentGroups).toHaveLength(1);
		expect(result?.equipmentGroups?.[0]?.groupLabel).toBe("head");
		expect(result?.equipmentGroups?.[0]?.items).toHaveLength(1);
	});
});
