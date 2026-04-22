import { cn } from "@/lib/utils";
import { Textarea } from "./ui/textarea";

/**
 * Wrap shadcn textarea with styling we use for addon export
 */
export function AddonExportTextarea({
	className,
	...props
}: Parameters<typeof Textarea>[0]) {
	return (
		<div className="flex flex-col">
			<Textarea
				placeholder={'priest="Example"\nlevel=80\nspec=shadow'}
				autoComplete="off"
				autoCapitalize="none"
				autoCorrect="off"
				className={cn("rounded-none h-32", className)}
				spellCheck={false}
				{...props}
			/>
			<div className="bg-secondary text-secondary-foreground p-2">
				Paste in your profile from the{" "}
				<a
					className="underline"
					href="https://www.curseforge.com/wow/addons/simulationcraft"
				>
					SimulationCraft
				</a>{" "}
				addon
			</div>
		</div>
	);
}
