import type { ErrorResponse } from "@/lib/saint-api/generated";

/**
 * Extract an API error message when Saint returns one, otherwise fall back to
 * the supplied default message.
 */
export function readSaintApiErrorMessage(
	error: string | ErrorResponse | undefined,
	fallback: string,
) {
	if (error === undefined) {
		return fallback;
	}

	if (typeof error === "string") {
		return error.trim();
	}

	if (error?.message) {
		return error.message;
	}

	return fallback;
}
