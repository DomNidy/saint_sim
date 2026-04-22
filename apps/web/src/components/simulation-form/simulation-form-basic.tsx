import { LoaderCircle, Sparkles } from "lucide-react";
import type { SubmitHandler, UseFormReturn } from "react-hook-form";
import type z from "zod";
import type { zSimulationConfigBasic } from "@/lib/saint-api/generated/zod.gen";
import { AddonExportTextarea } from "../addon-export-textarea";
import { Alert } from "../ui/alert";
import { Button } from "../ui/button";
import {
	Form,
	FormControl,
	FormDescription,
	FormField,
	FormItem,
	FormLabel,
	FormMessage,
} from "../ui/form";

type SimulationFormBasicProps = {
	form: UseFormReturn<z.infer<typeof zSimulationConfigBasic>>;
	submitHandler: SubmitHandler<z.infer<typeof zSimulationConfigBasic>>;

	// true if the form was submitting and is in a pending state.
	// you should set this to submit mutation isPending
	isSubmitPending: boolean;
};

export const SimulationFormBasic = ({
	form,
	submitHandler,
	isSubmitPending,
}: SimulationFormBasicProps) => {
	return (
		<Form {...form}>
			<form
				className="flex flex-col gap-5"
				onSubmit={form.handleSubmit(submitHandler)}
			>
				<FormField
					control={form.control}
					name="simc_addon_export"
					render={({ field }) => (
						<FormItem>
							<FormLabel>SimC addon export</FormLabel>
							<FormControl>
								<AddonExportTextarea
									placeholder={'priest="Example"\nlevel=80\nspec=shadow'}
									autoComplete="off"
									autoCapitalize="none"
									autoCorrect="off"
									className="h-32"
									spellCheck={false}
									{...field}
								/>
							</FormControl>
							<FormDescription>
								Paste the raw output from the SimulationCraft in-game addon.
								Saint will submit it to the backend verbatim.
							</FormDescription>
							<FormMessage />
						</FormItem>
					)}
				/>

				{form.formState?.errors?.root?.server?.message && (
					<Alert variant={"destructive"}>
						{form.formState?.errors?.root?.server?.message}
					</Alert>
				)}

				<div className="flex flex-wrap items-center gap-3">
					<Button disabled={isSubmitPending} type="submit">
						{isSubmitPending ? (
							<>
								<LoaderCircle
									data-icon="inline-start"
									className="animate-spin"
								/>
								Sending request
							</>
						) : (
							<>
								<Sparkles data-icon="inline-start" />
								Run simulation
							</>
						)}
					</Button>
					<Button
						type="button"
						variant="secondary"
						onClick={() => {
							form.reset({
								kind: "basic",
								simc_addon_export: "",
							});
						}}
					>
						Clear
					</Button>
				</div>
			</form>
		</Form>
	);
};
