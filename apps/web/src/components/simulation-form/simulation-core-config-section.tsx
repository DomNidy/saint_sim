import { useFormContext } from "react-hook-form";
import type z from "zod";
import { FightStyle } from "@/lib/saint-api/generated";
import type {
	zSimulationConfigBasic,
	zSimulationConfigTopGear,
	zSimulationCoreConfig,
	zSimulationOptions,
} from "@/lib/saint-api/generated/zod.gen";
import {
	FormControl,
	FormDescription,
	FormField,
	FormItem,
	FormLabel,
	FormMessage,
} from "../ui/form";
import {
	Select,
	SelectContent,
	SelectItem,
	SelectTrigger,
	SelectValue,
} from "../ui/select";

function formatFightStyleLabel(fightStyle: string) {
	return fightStyle
		.split("_")
		.map((word) => word.charAt(0).toUpperCase() + word.slice(1))
		.join(" ");
}

// core config section is used as a subsection in multiple simulation
// parent form types. zSimulationOptions is the union type that the
// simulation endpoint recevies. All types in that union should have a
// core_config property at the root.
type ParentForm = z.infer<typeof zSimulationOptions>;

/**
 * A re-usable sub-form for the core simulation config.
 *
 * Intended usage:
 * - Use as a child inside a <FormProvider> hierarchy (or shadcn's <Form>)
 * - The parent provider's form type should be a zSimulationConfigXXX type
 * as they carry core config at the same path
 */
export function SimulationCoreConfigSection() {
	const form = useFormContext<ParentForm>();

	if (form == null) {
		console.error(
			"SimulationCoreConfigSection must be used in a Form Context, and that form context must contain/be assignable to zSimulationCoreConfig",
		);
		return;
	}
	return (
		<div className="flex flex-col gap-5">
			<FormField
				control={form.control}
				name="core_config.fight_style"
				render={({ field }) => (
					<FormItem>
						<FormLabel>Fight Style</FormLabel>
						<Select onValueChange={field.onChange} value={field.value}>
							<FormControl>
								<SelectTrigger className="w-full">
									<SelectValue placeholder="Select a fight style" />
								</SelectTrigger>
							</FormControl>
							<SelectContent>
								{Object.values(FightStyle).map((fightStyle) => (
									<SelectItem key={fightStyle} value={fightStyle}>
										{formatFightStyleLabel(fightStyle)}
									</SelectItem>
								))}
							</SelectContent>
						</Select>
						<FormDescription>
							Choose the SimulationCraft fight profile to run.
						</FormDescription>
						<FormMessage />
					</FormItem>
				)}
			/>
		</div>
	);
}
