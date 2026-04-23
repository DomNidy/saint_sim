import React from "react";
import type { UseFormReturn } from "react-hook-form";
import type z from "zod";
import type { zSimulationCoreConfig } from "@/lib/saint-api/generated/zod.gen";
import {
	Form,
	FormControl,
	FormDescription,
	FormField,
	FormItem,
	FormLabel,
} from "../ui/form";

type SimulationCoreConfigFormProps = {
	form: UseFormReturn<z.infer<typeof zSimulationCoreConfig>>;
};
export function SimulationCoreConfigForm({
	form,
}: SimulationCoreConfigFormProps) {
	return (
		<Form {...form}>
			<form>
				<FormField
					control={form.control}
					name="fight_style"
					render={({ field }) => (
						<FormItem>
							<FormLabel>Fight Style</FormLabel>
							<FormControl />
							<FormDescription>Fight style simc config</FormDescription>
						</FormItem>
					)}
				/>
			</form>
		</Form>
	);
}
