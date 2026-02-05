import { createSignal } from "solid-js";
import * as Popover from "@kobalte/core/popover";
import { FaSolidHourglass } from "solid-icons/fa";
import { IoCloseSharp } from "solid-icons/io";

type DurationPopoverProps = {
	fieldPath: string;
	onSelect?: (value: string) => void;
};

export const DurationPopover = (props: DurationPopoverProps) => {
	const [value, setValue] = createSignal("1");
	const [unit, setUnit] = createSignal("s");

	const formatDuration = () => {
		const numValue = parseFloat(value());
		if (isNaN(numValue)) return "0s";
		return `${numValue}${unit()}`;
	};

	const handleApply = () => {
		props.onSelect?.(formatDuration());
	};

	const presetDurations = [
		{ label: "100ms", value: "0.1", unit: "s" },
		{ label: "1s", value: "1", unit: "s" },
		{ label: "5s", value: "5", unit: "s" },
		{ label: "30s", value: "30", unit: "s" },
		{ label: "1m", value: "60", unit: "s" },
		{ label: "5m", value: "300", unit: "s" },
		{ label: "1h", value: "3600", unit: "s" },
	];

	const applyPreset = (preset: { value: string; unit: string }) => {
		setValue(preset.value);
		setUnit(preset.unit);
	};

	return (
		<Popover.Root>
			<Popover.Trigger class="inline-flex items-center ml-1 cursor-pointer text-accent hover:text-accent-focus transition-colors">
				<FaSolidHourglass class="w-3 h-3" />
			</Popover.Trigger>
			<Popover.Portal>
				<Popover.Content class="z-50 rounded-lg border border-base-300 bg-base-100 p-4 shadow-lg min-w-[320px]">
					<Popover.Arrow />
					<div class="flex flex-col gap-3">
						<div class="text-sm font-semibold text-base-content">
							google.protobuf.Duration
						</div>
						<div class="text-xs text-base-content/60">
							Format: number + unit (s, m, h)
						</div>
						<div class="divider my-0" />
						<div class="flex gap-2">
							<input
								type="number"
								value={value()}
								onInput={(e) => setValue(e.currentTarget.value)}
								class="input input-sm input-bordered flex-1"
								placeholder="1"
								step="0.1"
							/>
							<select
								value={unit()}
								onChange={(e) => setUnit(e.currentTarget.value)}
								class="select select-sm select-bordered"
							>
								<option value="s">seconds</option>
								<option value="m">minutes</option>
								<option value="h">hours</option>
							</select>
						</div>
						<div class="text-xs text-base-content/80">Presets:</div>
						<div class="flex flex-wrap gap-1">
							{presetDurations.map((preset) => (
								<button
									onClick={() => applyPreset(preset)}
									class="btn btn-xs btn-ghost"
								>
									{preset.label}
								</button>
							))}
						</div>
						<div class="divider my-0" />
						<button onClick={handleApply} class="btn btn-sm btn-primary">
							Apply
						</button>
					</div>
					<Popover.CloseButton class="absolute cursor-pointer top-2 right-2 w-4 h-4 flex transition-all duration-300 ease-in-out items-center justify-center text-base-content/50 hover:text-base-content text-xs">
						<IoCloseSharp />
					</Popover.CloseButton>
				</Popover.Content>
			</Popover.Portal>
		</Popover.Root>
	);
};
