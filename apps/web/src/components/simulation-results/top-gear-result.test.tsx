// @vitest-environment jsdom

import { cleanup, render, screen } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import type {
	EquipmentItem,
	EquipmentSlot,
	SimulationResultTopGear,
} from "@/lib/saint-api/generated";
import { TopGearSimulationResultDisplay } from "./top-gear-result";

const refreshLinks = vi.fn();

const slotOrder = [
	"head",
	"neck",
	"shoulder",
	"back",
	"chest",
	"wrist",
	"hands",
	"waist",
	"legs",
	"feet",
	"finger1",
	"finger2",
	"trinket1",
	"trinket2",
	"main_hand",
	"off_hand",
] as const satisfies readonly EquipmentSlot[];

function buildEquipment(
	slot: EquipmentSlot,
	index: number,
	overrides: Partial<EquipmentItem> = {},
): EquipmentItem {
	return {
		slot,
		name: `${slot}_${index}`,
		display_name: `${slot} item ${index}`,
		item_id: 1000 + index,
		item_level: 480 + index,
		source: index % 2 === 0 ? "equipped" : "bag",
		raw_line: `${slot}=item_${index}`,
		...overrides,
	};
}

function buildTopGearResult(): SimulationResultTopGear {
	const equipment = slotOrder.map((slot, index) =>
		buildEquipment(
			slot,
			index,
			index === 0
				? {
						display_name: "Crown of Focused Tests",
						item_id: 190001,
						item_level: 489,
						enchant_id: 42,
						bonus_ids: [6652, 123],
						gem_ids: [11, 22],
					}
				: {},
		),
	);

	return {
		kind: "topGear",
		metric: "dps",
		equipment,
		profilesets: [
			{
				name: "Combo1",
				mean: 124_567,
				mean_error: 98.765,
				items: {
					head: 0,
					neck: 1,
					shoulder: 2,
					back: 3,
					chest: 4,
					wrist: 5,
					hands: 6,
					waist: 7,
					legs: 8,
					feet: 9,
					finger1: 10,
					finger2: 11,
					trinket1: 12,
					trinket2: 13,
					main_hand: 14,
					off_hand: 99,
				},
			},
			{
				name: "Combo2",
				mean: 120_000,
				items: {
					head: 0,
					neck: 1,
					shoulder: 2,
					back: 3,
					chest: 4,
					wrist: 5,
					hands: 6,
					waist: 7,
					legs: 8,
					feet: 9,
					finger1: 10,
					finger2: 11,
					trinket1: 12,
					trinket2: 13,
					main_hand: 14,
				},
			},
		],
	};
}

describe("TopGearSimulationResultDisplay", () => {
	beforeEach(() => {
		refreshLinks.mockClear();
		window.$WowheadPower = {
			refreshLinks,
		};
	});

	afterEach(() => {
		cleanup();
		delete window.$WowheadPower;
	});

	it("renders each profileset with metrics and resolved item cards", () => {
		const { container } = render(
			<TopGearSimulationResultDisplay {...buildTopGearResult()} />,
		);

		expect(screen.getByText("Top gear results")).toBeTruthy();
		expect(screen.getByText("2 profilesets ranked by dps")).toBeTruthy();
		expect(screen.getByText("Combo1")).toBeTruthy();
		expect(screen.getByText("Combo2")).toBeTruthy();
		expect(screen.getByText("Mean dps: 124,567")).toBeTruthy();
		expect(screen.getByText("Mean error: 98.77")).toBeTruthy();
		expect(screen.getByText("Mean error: unavailable")).toBeTruthy();
		expect(screen.getAllByText("Off Hand")).toHaveLength(1);
		expect(screen.getByText("Missing equipment index 99")).toBeTruthy();
		expect(screen.getAllByText("Crown of Focused Tests")).toHaveLength(2);

		const itemLink = container.querySelector(
			'a[href="https://www.wowhead.com/item=190001"]',
		);
		expect(itemLink?.getAttribute("data-wowhead")).toBe(
			"item=190001&bonus=6652:123&ench=42&gems=11:22&ilvl=489",
		);
		expect(itemLink?.getAttribute("data-wh-icon-size")).toBe("medium");
		expect(refreshLinks).toHaveBeenCalledTimes(1);
	});
});
