import { createSignal } from "solid-js";
import * as Popover from "@kobalte/core/popover";
import { ImClock } from "solid-icons/im";
import { IoCloseSharp } from "solid-icons/io";

type TimestampPopoverProps = {
	fieldPath: string;
	onSelect?: (value: string) => void;
};

export const TimestampPopover = (props: TimestampPopoverProps) => {
	const [datetime, setDatetime] = createSignal(new Date().toISOString().slice(0, 16));

	const handleApply = () => {
		const selectedDate = new Date(datetime());
		const isoString = selectedDate.toISOString();
		props.onSelect?.(isoString);
	};

	const setNow = () => {
		const now = new Date();
		setDatetime(now.toISOString().slice(0, 16));
	};

	return (
		<Popover.Root>
			<Popover.Trigger class="inline-flex items-center ml-1 cursor-pointer text-secondary hover:text-secondary/60 transition-colors">
				<ImClock class="w-3.5 h-3.5" />
			</Popover.Trigger>
			<Popover.Portal>
				<Popover.Content class="z-50 rounded-lg border border-base-300 bg-base-100 p-4 shadow-lg min-w-[280px]">
					<Popover.Arrow />
					<div class="flex flex-col gap-3">
						<div class="text-sm font-semibold text-base-content">google.protobuf.Timestamp</div>
						<div class="text-xs text-base-content/60">RFC3339 format: "2026-02-05T14:05:47Z"</div>
						<div class="divider my-0" />
						<div class="flex flex-col gap-2">
							<label class="text-xs text-base-content/80">Date & Time</label>
							<input
								type="datetime-local"
								value={datetime()}
								onInput={e => setDatetime(e.currentTarget.value)}
								class="input input-sm input-bordered"
							/>
						</div>
						<div class="flex gap-2">
							<button onClick={setNow} class="btn btn-sm btn-ghost flex-1">
								Now
							</button>
							<button onClick={handleApply} class="btn btn-sm btn-primary flex-1">
								Apply
							</button>
						</div>
					</div>
					<Popover.CloseButton class="absolute cursor-pointer top-2 right-2 w-4 h-4 flex transition-all duration-300 ease-in-out items-center justify-center text-base-content/50 hover:text-base-content text-xs">
						<IoCloseSharp />
					</Popover.CloseButton>
				</Popover.Content>
			</Popover.Portal>
		</Popover.Root>
	);
};
