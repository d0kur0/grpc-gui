import { For } from "solid-js";
import * as Popover from "@kobalte/core/popover";
import { FaSolidCircleQuestion } from "solid-icons/fa";
import { IoCloseSharp } from 'solid-icons/io'

type EnumValue = {
	name: string;
	number: number;
};

type EnumPopoverProps = {
	enumValues: EnumValue[];
	fieldPath: string;
};

export const EnumPopover = (props: EnumPopoverProps) => {
	return (
		<Popover.Root>
			<Popover.Trigger class="inline-flex items-center ml-1 cursor-pointer text-accent hover:text-accent-focus transition-colors">
				<FaSolidCircleQuestion class="w-3 h-3" />
			</Popover.Trigger>
			<Popover.Portal>
				<Popover.Content class="z-50 rounded-lg border border-base-300 bg-base-100 p-4 shadow-lg">
					<Popover.Arrow />
					<div class="flex flex-col gap-1">
						<For each={props.enumValues}>
							{(enumValue) => (
								<div class="text-sm text-base-content/80 font-mono">
									<span class="text-primary">{enumValue.name}</span>
									<span class="text-base-content/50"> = {enumValue.number}</span>
								</div>
							)}
						</For>
					</div>
					<Popover.CloseButton class="absolute cursor-pointer top-0 right-0 w-4 h-4 flex transition-all duration-300 ease-in-out items-center justify-center text-base-content/50 hover:text-base-content text-xs">
						<IoCloseSharp  />
					</Popover.CloseButton>
				</Popover.Content>
			</Popover.Portal>
		</Popover.Root>
	);
};
