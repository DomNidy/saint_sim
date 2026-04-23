import { LoaderCircle, Sparkles } from "lucide-react";
import type { SubmitHandler, UseFormReturn } from "react-hook-form";
import type z from "zod";
import type { zSimulationConfigBasic } from "@/lib/saint-api/generated/zod.gen";
import { Alert } from "../ui/alert";
import { Button } from "../ui/button";
import { Form } from "../ui/form";

type SimulationFormBasicProps = {
	form: UseFormReturn<z.infer<typeof zSimulationConfigBasic>>;
	submitHandler: SubmitHandler<z.infer<typeof zSimulationConfigBasic>>;
	resetHandler: () => void;

	// true if the form was submitting and is in a pending state.
	// you should set this to submit mutation isPending
	isSubmitPending: boolean;
};

export const SimulationFormBasic = ({
	form,
	submitHandler,
	resetHandler,
	isSubmitPending,
}: SimulationFormBasicProps) => {
	return (
		<Form {...form}>
			<form
				className="flex flex-col gap-5"
				onSubmit={form.handleSubmit(submitHandler)}
			>
				{form.formState?.errors?.root?.server?.message && (
					<Alert variant={"destructive"}>
						{form.formState?.errors?.root?.server?.message}
					</Alert>
				)}

				{form.formState?.errors && (
					<Alert variant={"destructive"}>
						{JSON.stringify(form.formState.errors)}
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
						onClick={() => resetHandler()}
					>
						Clear
					</Button>
				</div>
			</form>
		</Form>
	);
};
