import { TiChevronRight } from "solid-icons/ti";
import { JSX, Show } from "solid-js";

export type DropDownContainerProps = {
	open: boolean;
	title: string;
	prefix?: JSX.Element;
	children: JSX.Element | JSX.Element[];
	onOpenChange: (open: boolean) => void;
	headerWrapper?: (header: JSX.Element) => JSX.Element;
};

export const DropDownContainer = (props: DropDownContainerProps) => {
	const header = (
		<button
			class="cursor-pointer flex gap-0.5 items-center text-base-content/70 py-1 w-full transition-all duration-300 hover:text-base-content"
			onClick={() => props.onOpenChange(!props.open)}>
			<TiChevronRight class="text-xs" classList={{ "rotate-90": props.open }} />
			<Show when={props.prefix}>
				{props.prefix}
			</Show>
			<span class="text-sm truncate">{props.title}</span>
		</button>
	);

	return (
		<div class="">
			{props.headerWrapper ? props.headerWrapper(header) : header}

			<Show when={props.open}>
				<div class="flex-1 ml-1 border-l-2 border-base-content/10 pl-4 hover:border-base-content/20 transition-all duration-300">
					{props.children}
				</div>
			</Show>
		</div>
	);
};
