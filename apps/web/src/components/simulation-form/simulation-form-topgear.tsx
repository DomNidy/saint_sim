import { LoaderCircle, Sparkles } from "lucide-react";
import type { SubmitHandler, UseFormReturn } from "react-hook-form";
import type z from "zod";
import type { zSimulationConfigTopGear } from "@/lib/saint-api/generated/zod.gen";
import { Alert } from "../ui/alert";
import { Button } from "../ui/button";
import { Form } from "../ui/form";

type SimuationFormTopGearProps = {
	form: UseFormReturn<z.infer<typeof zSimulationConfigTopGear>>;
	submitHandler: SubmitHandler<z.infer<typeof zSimulationConfigTopGear>>;

	// true if the form was submitting and is in a pending state.
	// you should set this to submit mutation isPending
	isSubmitPending: boolean;
};

export const SimulationFormTopGear = ({
	form,
	submitHandler,
	isSubmitPending,
}: SimuationFormTopGearProps) => {
	return (
		<Form {...form}>
			<form
				className="flex flex-col gap-5"
				onSubmit={form.handleSubmit(submitHandler)}
			>
				<p>Top Gear is a W.I.P</p>

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
								kind: "topGear",
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
