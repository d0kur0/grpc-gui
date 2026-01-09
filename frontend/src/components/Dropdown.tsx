import { TiChevronRight } from "solid-icons/ti";
import { JSX, Show } from "solid-js";

export type DropDownContainerProps = {
	open: boolean;
	title: string;
	children: JSX.Element | JSX.Element[];
	onOpenChange: (open: boolean) => void;
};

export const DropDownContainer = (props: DropDownContainerProps) => {
	return (
		<div class="">
			<button
				class="cursor-pointer flex gap-0.5 items-center text-base-content/70 py-1 transition-all duration-300 hover:text-base-content"
				onClick={() => props.onOpenChange(!props.open)}>
				<TiChevronRight class="text-xs" classList={{ "rotate-90": props.open }} />
				<span class="text-sm">{props.title}</span>
			</button>

			<Show when={props.open}>
				<div class="flex-1 ml-1 border-l-2 border-base-content/10 pl-4 hover:border-base-content/20 transition-all duration-300">
					{props.children}
				</div>
			</Show>
		</div>
	);
};
