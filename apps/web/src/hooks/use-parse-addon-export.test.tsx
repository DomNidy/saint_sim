// @vitest-environment jsdom

import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { act, renderHook, waitFor } from "@testing-library/react";
import type { ReactNode } from "react";
import { describe, expect, it, vi } from "vitest";
import {
	canonicalizeSimcAddonExport,
	useParseAddonExport,
} from "./use-parse-addon-export";

const parseAddonExportMock = vi.fn();

vi.mock("@/lib/simulation/parse-addon-export.functions", () => ({
	parseAddonExport: (...args: unknown[]) => parseAddonExportMock(...args),
}));

function createWrapper() {
	const queryClient = new QueryClient({
		defaultOptions: {
			queries: {
				retry: false,
			},
		},
	});

	return ({ children }: { children: ReactNode }) => (
		<QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
	);
}

describe("useParseAddonExport", () => {
	it("canonicalizes equivalent addon exports to the same string", () => {
		const withWindowsLineEndings = canonicalizeSimcAddonExport(
			'\npriest="Example"\r\nlevel=80\t \r\nspec=shadow\r\n\n',
		);
		const alreadyCanonical = canonicalizeSimcAddonExport(
			'priest="Example"\nlevel=80\nspec=shadow',
		);

		expect(withWindowsLineEndings).toBe(alreadyCanonical);
	});

	it("does not issue a request when disabled", () => {
		parseAddonExportMock.mockReset();

		renderHook(() => useParseAddonExport('priest="Example"', false), {
			wrapper: createWrapper(),
		});

		expect(parseAddonExportMock).not.toHaveBeenCalled();
	});

	it("debounces requests and deduplicates equivalent canonical inputs", async () => {
		vi.useFakeTimers();
		parseAddonExportMock.mockReset();
		parseAddonExportMock.mockResolvedValue({
			addon_export: { equipment: [] },
		});

		const { rerender } = renderHook(
			({ value }) => useParseAddonExport(value, true),
			{
				initialProps: {
					value: 'priest="Example"\r\nlevel=80\t \r\n',
				},
				wrapper: createWrapper(),
			},
		);

		expect(parseAddonExportMock).not.toHaveBeenCalled();

		await act(async () => {
			vi.advanceTimersByTime(400);
		});

		await waitFor(() => {
			expect(parseAddonExportMock).toHaveBeenCalledTimes(1);
		});

		expect(parseAddonExportMock).toHaveBeenLastCalledWith({
			data: {
				rawAddonExport: 'priest="Example"\nlevel=80',
			},
		});

		rerender({
			value: 'priest="Example"\nlevel=80',
		});

		await act(async () => {
			vi.advanceTimersByTime(400);
		});

		expect(parseAddonExportMock).toHaveBeenCalledTimes(1);

		rerender({
			value: 'priest="Example"\nlevel=80\nspec=shadow',
		});

		await act(async () => {
			vi.advanceTimersByTime(400);
		});

		await waitFor(() => {
			expect(parseAddonExportMock).toHaveBeenCalledTimes(2);
		});

		vi.useRealTimers();
	});
});
