import { $tabs, Tab, TabComponent } from "../stores/tabs";
import { createMemo, createSignal, For, onMount, Show } from "solid-js";
import "./Tabs.css";
import { OcFilecode2 } from "solid-icons/oc";
import { SendRequest, SendRequestProps } from "./SendRequest";
import { IoClose } from "solid-icons/io";


export const Tabs = () => {
	const { tabs } = $tabs;

	const tab = createMemo(() => {
		return tabs().find(tab => tab.isActive);
	});

	const handleWheel = (e: WheelEvent) => {
		if (Math.abs(e.deltaY) > Math.abs(e.deltaX)) {
			(e.currentTarget as HTMLDivElement).scrollLeft += e.deltaY;
		}
	};

	const renderTab = (tab: Tab) => {
		switch (tab.component) {
			case TabComponent.REQUEST:
				console.log(tab.componentProps);
				return <SendRequest {...(tab.componentProps as SendRequestProps)} />;
			default:
				return null;
		}
	};

	return (
		<div class="app-tabs">
			<Show when={tab()}>
				<div class="app-tabs-list scrollbar" on:wheel={{
					passive: true,
					handleEvent: handleWheel
				}
				}>
					<For each={tabs()}>
						{tab => {

							
							return <div role="button" onClick={() => $tabs.activateTab(tab.id)} classList={{ "app-tab-active": tab.isActive }} class="app-tab">
								<span class="app-tab-name">{tab.name}</span>
								<span title="Закрыть" role="button" class="app-tab-close" onClick={(e) => {
									e.stopPropagation();
									$tabs.removeTab(tab.id);
								}}>
									<IoClose />
								</span>
							</div>;
						}}
					</For>
				</div>

				<div class="app-tab-component scrollbar">{renderTab(tab()!)}</div>
			</Show>

			<Show when={!tab()}>
				<div class="flex h-full w-full items-center justify-center text-5xl text-base-content/20">
					<OcFilecode2 />
				</div>
			</Show>
		</div>
	);
};
