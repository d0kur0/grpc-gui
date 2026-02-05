import { $tabs, Tab, TabComponent } from "../stores/tabs";
import { createEffect, createMemo, Index, Show } from "solid-js";
import "./Tabs.css";
import { OcFilecode2 } from "solid-icons/oc";
import { SendRequest, SendRequestProps } from "./SendRequest";
import { IoClose } from "solid-icons/io";


export const Tabs = () => {
	const { tabs } = $tabs;
	let tabsListRef: HTMLDivElement | undefined;

	const tab = createMemo(() => {
		return tabs().find(tab => tab.isActive);
	});

	const handleWheel = (e: WheelEvent) => {
		if (Math.abs(e.deltaY) > Math.abs(e.deltaX)) {
			(e.currentTarget as HTMLDivElement).scrollLeft += e.deltaY;
		}
	};

	createEffect(() => {
		if (!tab() || !tabsListRef) return;

		const tabElement = tabsListRef.querySelector(`[data-tab-id="${tab()!.id}"]`) as HTMLElement; 
		if (!tabElement) return;

		const containerRect = tabsListRef.getBoundingClientRect();
		const tabRect = tabElement.getBoundingClientRect();
		const currentScroll = tabsListRef.scrollLeft;

		const isLeftCut = tabRect.left < containerRect.left;
		const isRightCut = tabRect.right > containerRect.right;

		if (!isLeftCut && !isRightCut) return;

		const diff = isLeftCut 
			? containerRect.left - tabRect.left
			: tabRect.right - containerRect.right;
		
		const scrollTo = isLeftCut ? currentScroll - diff : currentScroll + diff;

		tabsListRef.scrollTo({ left: scrollTo, behavior: 'smooth' });
	})



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
				<div ref={tabsListRef} class="app-tabs-list scrollbar" on:wheel={{
					passive: true,
					handleEvent: handleWheel
				}
				}>
					<Index each={tabs()}>
						{(tab, index) => {
							const handleCloseTab = (e: MouseEvent) => {
								e.stopPropagation();
								$tabs.removeTab(tab().id);
							};

							const handleTabClick = () => {
								$tabs.activateTab(tab().id);
							};
							
							return <div data-tab-id={tab().id} role="button" onClick={handleTabClick} classList={{ "app-tab-active": tab().isActive }} class="app-tab">
								<span class="app-tab-name">{tab().name}</span>
								<span title="Закрыть" role="button" class="app-tab-close" onClick={handleCloseTab}>
									<IoClose />
								</span>
							</div>;
						}}
					</Index>
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
