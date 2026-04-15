"use client";

import { Link } from "@tanstack/react-router";
import { LogOut, User } from "lucide-react";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Button } from "@/components/ui/button";
import {
	DropdownMenu,
	DropdownMenuContent,
	DropdownMenuGroup,
	DropdownMenuItem,
	DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { authClient } from "@/lib/auth/auth-client";

type HeaderUserAvatarProps = {
	image?: string | null;
	name?: string | null;
};

function HeaderUserAvatar({ image, name }: HeaderUserAvatarProps) {
	const initial = name?.charAt(0).toUpperCase() || "U";

	return (
		<Avatar className="size-8 border border-transparent transition-all duration-200 group-hover:border-border group-hover:bg-accent/40 group-hover:shadow-sm">
			{image ? <AvatarImage src={image} alt={name ?? ""} /> : null}
			<AvatarFallback>{initial}</AvatarFallback>
		</Avatar>
	);
}

export default function BetterAuthHeader() {
	const { data: session, isPending } = authClient.useSession();

	if (isPending) {
		return <div className="size-8 rounded-full bg-muted animate-pulse" />;
	}

	if (session?.user) {
		return (
			<DropdownMenu>
				<DropdownMenuTrigger asChild>
					<Button
						variant="ghost"
						size="icon-sm"
						className="group rounded-full"
						aria-label="Open account menu"
					>
						<HeaderUserAvatar
							image={session.user.image}
							name={session.user.name}
						/>
					</Button>
				</DropdownMenuTrigger>
				<DropdownMenuContent align="start">
					<DropdownMenuGroup>
						<DropdownMenuItem asChild>
							<Link to="/account">
								<User data-icon="inline-start" />
								Account
							</Link>
						</DropdownMenuItem>
						<DropdownMenuItem
							onSelect={() => {
								void authClient.signOut();
							}}
						>
							<LogOut data-icon="inline-start" />
							Logout
						</DropdownMenuItem>
					</DropdownMenuGroup>
				</DropdownMenuContent>
			</DropdownMenu>
		);
	}

	return (
		<Link
			to="/auth/sign-in"
			className="inline-flex h-9 items-center rounded-md border border-border bg-background px-4 text-sm font-medium text-foreground transition-colors hover:bg-accent hover:text-accent-foreground"
		>
			Sign in
		</Link>
	);
}
