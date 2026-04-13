import * as React from "react";

import { cn } from "@/lib/utils";

const Textarea = React.forwardRef<
	HTMLTextAreaElement,
	React.ComponentProps<"textarea">
>(({ className, ...props }, ref) => {
	return (
		<textarea
			ref={ref}
			className={cn(
				"flex min-h-32 w-full rounded-2xl border border-[var(--line)] bg-white/70 px-4 py-3 text-sm text-[var(--sea-ink)] shadow-[0_1px_0_var(--inset-glint)_inset] transition placeholder:text-[var(--sea-ink-soft)]/70 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-[var(--lagoon)] focus-visible:ring-offset-2 focus-visible:ring-offset-transparent disabled:cursor-not-allowed disabled:opacity-50 dark:bg-[rgba(13,28,32,0.85)]",
				className,
			)}
			{...props}
		/>
	);
});

Textarea.displayName = "Textarea";

export { Textarea };
