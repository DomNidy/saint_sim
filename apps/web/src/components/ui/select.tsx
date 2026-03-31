import * as React from "react";
import * as SelectPrimitive from "@radix-ui/react-select";
import { Check, ChevronDown, ChevronUp } from "lucide-react";

import { cn } from "#/lib/utils";

function Select(props: React.ComponentProps<typeof SelectPrimitive.Root>) {
	return <SelectPrimitive.Root data-slot="select" {...props} />;
}

function SelectGroup(
	props: React.ComponentProps<typeof SelectPrimitive.Group>,
) {
	return <SelectPrimitive.Group data-slot="select-group" {...props} />;
}

function SelectValue(
	props: React.ComponentProps<typeof SelectPrimitive.Value>,
) {
	return <SelectPrimitive.Value data-slot="select-value" {...props} />;
}

const SelectTrigger = React.forwardRef<
	React.ElementRef<typeof SelectPrimitive.Trigger>,
	React.ComponentPropsWithoutRef<typeof SelectPrimitive.Trigger>
>(({ className, children, ...props }, ref) => (
	<SelectPrimitive.Trigger
		ref={ref}
		data-slot="select-trigger"
		className={cn(
			"flex h-11 w-full items-center justify-between gap-2 rounded-2xl border border-[var(--line)] bg-white/70 px-4 py-2 text-sm text-[var(--sea-ink)] shadow-[0_1px_0_var(--inset-glint)_inset] transition focus:outline-none focus:ring-2 focus:ring-[var(--lagoon)] focus:ring-offset-2 focus:ring-offset-transparent disabled:cursor-not-allowed disabled:opacity-50 data-[placeholder]:text-[var(--sea-ink-soft)]/70 dark:bg-[rgba(13,28,32,0.85)]",
			className,
		)}
		{...props}
	>
		{children}
		<SelectPrimitive.Icon asChild>
			<ChevronDown className="size-4 opacity-70" />
		</SelectPrimitive.Icon>
	</SelectPrimitive.Trigger>
));

SelectTrigger.displayName = SelectPrimitive.Trigger.displayName;

const SelectContent = React.forwardRef<
	React.ElementRef<typeof SelectPrimitive.Content>,
	React.ComponentPropsWithoutRef<typeof SelectPrimitive.Content>
>(({ className, children, position = "popper", ...props }, ref) => (
	<SelectPrimitive.Portal>
		<SelectPrimitive.Content
			ref={ref}
			data-slot="select-content"
			position={position}
			className={cn(
				"relative z-50 max-h-96 min-w-[8rem] overflow-hidden rounded-2xl border border-[var(--line)] bg-[var(--surface-strong)] text-[var(--sea-ink)] shadow-[0_24px_48px_rgba(23,58,64,0.18)] backdrop-blur-xl",
				position === "popper" &&
					"data-[side=bottom]:translate-y-1 data-[side=left]:-translate-x-1 data-[side=right]:translate-x-1 data-[side=top]:-translate-y-1",
				className,
			)}
			{...props}
		>
			<SelectScrollUpButton />
			<SelectPrimitive.Viewport
				className={cn(
					"p-1.5",
					position === "popper" &&
						"h-[var(--radix-select-trigger-height)] w-full min-w-[var(--radix-select-trigger-width)]",
				)}
			>
				{children}
			</SelectPrimitive.Viewport>
			<SelectScrollDownButton />
		</SelectPrimitive.Content>
	</SelectPrimitive.Portal>
));

SelectContent.displayName = SelectPrimitive.Content.displayName;

const SelectLabel = React.forwardRef<
	React.ElementRef<typeof SelectPrimitive.Label>,
	React.ComponentPropsWithoutRef<typeof SelectPrimitive.Label>
>(({ className, ...props }, ref) => (
	<SelectPrimitive.Label
		ref={ref}
		data-slot="select-label"
		className={cn(
			"px-3 py-2 text-xs font-semibold tracking-[0.18em] text-[var(--sea-ink-soft)] uppercase",
			className,
		)}
		{...props}
	/>
));

SelectLabel.displayName = SelectPrimitive.Label.displayName;

const SelectItem = React.forwardRef<
	React.ElementRef<typeof SelectPrimitive.Item>,
	React.ComponentPropsWithoutRef<typeof SelectPrimitive.Item>
>(({ className, children, ...props }, ref) => (
	<SelectPrimitive.Item
		ref={ref}
		data-slot="select-item"
		className={cn(
			"relative flex w-full cursor-default select-none items-center rounded-xl py-2.5 pl-9 pr-3 text-sm outline-none transition focus:bg-[rgba(79,184,178,0.14)] focus:text-[var(--sea-ink)] data-[disabled]:pointer-events-none data-[disabled]:opacity-50",
			className,
		)}
		{...props}
	>
		<span className="absolute left-3 flex size-4 items-center justify-center">
			<SelectPrimitive.ItemIndicator>
				<Check className="size-4" />
			</SelectPrimitive.ItemIndicator>
		</span>
		<SelectPrimitive.ItemText>{children}</SelectPrimitive.ItemText>
	</SelectPrimitive.Item>
));

SelectItem.displayName = SelectPrimitive.Item.displayName;

const SelectScrollUpButton = React.forwardRef<
	React.ElementRef<typeof SelectPrimitive.ScrollUpButton>,
	React.ComponentPropsWithoutRef<typeof SelectPrimitive.ScrollUpButton>
>(({ className, ...props }, ref) => (
	<SelectPrimitive.ScrollUpButton
		ref={ref}
		data-slot="select-scroll-up-button"
		className={cn(
			"flex cursor-default items-center justify-center py-1",
			className,
		)}
		{...props}
	>
		<ChevronUp className="size-4" />
	</SelectPrimitive.ScrollUpButton>
));

SelectScrollUpButton.displayName = SelectPrimitive.ScrollUpButton.displayName;

const SelectScrollDownButton = React.forwardRef<
	React.ElementRef<typeof SelectPrimitive.ScrollDownButton>,
	React.ComponentPropsWithoutRef<typeof SelectPrimitive.ScrollDownButton>
>(({ className, ...props }, ref) => (
	<SelectPrimitive.ScrollDownButton
		ref={ref}
		data-slot="select-scroll-down-button"
		className={cn(
			"flex cursor-default items-center justify-center py-1",
			className,
		)}
		{...props}
	>
		<ChevronDown className="size-4" />
	</SelectPrimitive.ScrollDownButton>
));

SelectScrollDownButton.displayName =
	SelectPrimitive.ScrollDownButton.displayName;

export {
	Select,
	SelectContent,
	SelectGroup,
	SelectItem,
	SelectLabel,
	SelectScrollDownButton,
	SelectScrollUpButton,
	SelectTrigger,
	SelectValue,
};
