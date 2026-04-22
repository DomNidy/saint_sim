import { createFileRoute } from "@tanstack/react-router";
import { useForm } from "react-hook-form";
import type z from "zod";
import { SimulationFormTopGear } from "@/components/simulation-form/simulation-form-topgear";
import type { zSimulationOptionsTopGear } from "@/lib/saint-api/generated/zod.gen";

export const Route = createFileRoute("/simulate/top-gear")({
	component: RouteComponent,
});

function RouteComponent() {
	const form = useForm<z.infer<typeof zSimulationOptionsTopGear>>({});

	return (
		<div>
			<SimulationFormTopGear
				form={form}
				isSubmitPending={false}
				submitHandler={(values) => console.log(values)}
			/>
		</div>
	);
}
