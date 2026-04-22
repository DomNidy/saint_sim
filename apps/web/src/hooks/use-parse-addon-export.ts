import { useQuery } from "@tanstack/react-query";
import { useMemo } from "react";
import { groupEquipment } from "@/lib/equipment/group";
import { parseAddonExport } from "@/lib/simulation/parse-addon-export.functions";
import { useDebouncedValue } from "./use-debounced-value";

export function canonicalizeSimcAddonExport(raw: string): string {
	return raw
		.replace(/\r\n?/g, "\n")
		.split("\n")
		.map((line) => line.replace(/[ \t]+$/g, ""))
		.join("\n")
		.replace(/^\n+|\n+$/g, "");
}

/**
 * Parse a SimulationCraft addon export through the Saint API.
 *
 * In addition to the parsed addon export object, a `groups` object is also
 * returned (derived from the parsed items). Each item is grouped by its
 * intended slot (in WoW).
 *
 * Normalizes the input before querying so equivalent exports share the same
 * React Query cache key.
 *
 * Params:
 * - `rawInput`: raw addon export string to parse.
 * - `enabled`: whether parsing should run at all.
 *
 * Behavior:
 * - canonicalizes line endings and trailing whitespace
 * - fetches immediately for the initial non-empty enabled input
 * - debounces subsequent canonical input changes before calling the Saint API
 * - skips requests when disabled or when the canonicalized input is empty
 *
 * Intended usage:
 * - call this where UI needs parsed addon export data
 * - read the parsed result from `query.data?.addon_export`
 */
export function useParseAddonExport(rawInput: string, enabled: boolean) {
	const canonical = useMemo(
		() => canonicalizeSimcAddonExport(rawInput),
		[rawInput],
	);
	const debouncedCanonical = useDebouncedValue(canonical, 400);

	return useQuery({
		queryKey: ["parse-addon-export", debouncedCanonical],
		queryFn: () =>
			parseAddonExport({
				data: { rawAddonExport: debouncedCanonical },
			}),
		select: (data) => {
			return {
				...data,
				groups: groupEquipment(data.addon_export?.equipment ?? []),
			};
		},
		enabled: enabled && debouncedCanonical.length > 0,
		staleTime: 5 * 60 * 1000,
		gcTime: 30 * 60 * 1000,
		refetchOnWindowFocus: false,
		retry: false,
	});
}
