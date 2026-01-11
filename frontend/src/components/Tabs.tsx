import { $tabs, Tab } from "../stores/tabs";
import { createMemo, createSignal, For, Show } from "solid-js";
import "./Tabs.css";
import { OcFilecode2 } from "solid-icons/oc";

export const Tabs = () => {
	const { tabs } = $tabs;

	const tab = createMemo(() => {
		return tabs().find(tab => tab.isActive);
	});

	return (
		<div class="tabs h-full">
			<Show when={tab()}>
				<div class="tabs-list">
					<For each={tabs()}>
						{tab => {
							return <button class="tab">{tab.name}</button>;
						}}
					</For>
				</div>

				<div class="tab-content">{tab()?.content}</div>
			</Show>

			<Show when={!tab()}>
				<div class="flex h-full w-full items-center justify-center text-5xl text-base-content/20">
					<OcFilecode2 />
				</div>
			</Show>
		</div>
	);
};
