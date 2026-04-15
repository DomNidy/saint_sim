// @vitest-environment jsdom

import { cleanup, fireEvent, render, screen } from "@testing-library/react";
import type { ReactNode } from "react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

const useSessionMock = vi.fn();
const signOutMock = vi.fn(() => Promise.resolve());

vi.mock("@tanstack/react-router", () => ({
	Link: ({ children, to, ...props }: { children: ReactNode; to: string }) => (
		<a href={to} {...props}>
			{children}
		</a>
	),
}));

vi.mock("@/lib/auth/auth-client", () => ({
	authClient: {
		useSession: () => useSessionMock(),
		signOut: () => signOutMock(),
	},
}));

import BetterAuthHeader from "./header-user";

describe("BetterAuthHeader", () => {
	beforeEach(() => {
		useSessionMock.mockReset();
		signOutMock.mockClear();
	});

	afterEach(() => {
		cleanup();
	});

	it("renders a loading placeholder while the session is pending", () => {
		useSessionMock.mockReturnValue({ data: null, isPending: true });

		const { container } = render(<BetterAuthHeader />);

		expect(container.querySelector(".animate-pulse")).not.toBeNull();
	});

	it("renders the sign in link when there is no user session", () => {
		useSessionMock.mockReturnValue({ data: null, isPending: false });

		render(<BetterAuthHeader />);

		const signInLink = screen.getByRole("link", { name: "Sign in" });
		expect(signInLink.getAttribute("href")).toBe("/auth/sign-in");
	});

	it("opens the menu and wires account and logout actions for an authenticated user", () => {
		useSessionMock.mockReturnValue({
			data: {
				user: {
					name: "Ada Lovelace",
					image: null,
				},
			},
			isPending: false,
		});

		render(<BetterAuthHeader />);

		fireEvent.pointerDown(
			screen.getByRole("button", { name: "Open account menu" }),
			{
				button: 0,
				ctrlKey: false,
			},
		);

		const accountLink = screen.getByRole("menuitem", { name: "Account" });
		expect(accountLink.getAttribute("href")).toBe("/account");

		fireEvent.click(screen.getByRole("menuitem", { name: "Logout" }));
		expect(signOutMock).toHaveBeenCalledTimes(1);
	});
});
