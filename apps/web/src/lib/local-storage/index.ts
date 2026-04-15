import { createClientOnlyFn } from "@tanstack/react-start";

// Storage key to previous simc export that was submitted
const PREV_SIMC_EXPORT_KEY = "previousSimcExport";

/**
 * Client only wrapper fn around local storage set.
 * Throws error if called on server
 *
 * TIP: useHydrated() hook to run only on client
 */
const localStorageSet = createClientOnlyFn((key: string, value: string) => {
	localStorage.setItem(key, value);
});

/**
 * Client only wrapper fn around local storage get.
 * Throws error if called on server.
 *
 * TIP: useHydrated() hook to run only on client
 */
const localStorageGet = createClientOnlyFn((key: string): string | null => {
	return localStorage.getItem(key);
});

export {
	PREV_SIMC_EXPORT_KEY as PREV_SIMC_PROFILE_KEY,
	localStorageGet,
	localStorageSet,
};
