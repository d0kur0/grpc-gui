import { $tabs, Tab, TabType } from "../stores/tabs";
import { createEffect, createMemo, Index, Show, onMount } from "solid-js";
import "./Tabs.css";
import { OcFilecode2 } from "solid-icons/oc";
import { SendRequest } from "./SendRequest";
import { IoClose } from "solid-icons/io";
import { ContextMenu } from "@kobalte/core/context-menu";
import { truncateFromEnd } from "../utils/truncate";

export const Tabs = () => {
	const { tabs, loadTabs, isLoaded } = $tabs;
	let tabsListRef: HTMLDivElement | undefined;

	onMount(() => {
		loadTabs();
	});

	const activeTab = createMemo(() => {
		return tabs.find(tab => tab.isActive);
	});

	const handleWheel = (e: WheelEvent) => {
		if (Math.abs(e.deltaY) > Math.abs(e.deltaX)) {
			(e.currentTarget as HTMLDivElement).scrollLeft += e.deltaY;
		}
	};

	createEffect(() => {
		const tab = activeTab();
		if (!tab || !tabsListRef) return;

		const tabElement = tabsListRef.querySelector(`[data-tab-id="${tab.id}"]`) as HTMLElement;
		if (!tabElement) return;

		const containerRect = tabsListRef.getBoundingClientRect();
		const tabRect = tabElement.getBoundingClientRect();
		const currentScroll = tabsListRef.scrollLeft;

		const isLeftCut = tabRect.left < containerRect.left;
		const isRightCut = tabRect.right > containerRect.right;

		if (!isLeftCut && !isRightCut) return;

		const diff = isLeftCut ? containerRect.left - tabRect.left : tabRect.right - containerRect.right;

		const scrollTo = isLeftCut ? currentScroll - diff : currentScroll + diff;

		tabsListRef.scrollTo({ left: scrollTo, behavior: "smooth" });
	});

	return (
		<div class="app-tabs">
			<Show when={isLoaded() && activeTab()}>
				<div
					ref={tabsListRef}
					class="app-tabs-list scrollbar"
					on:wheel={{
						passive: true,
						handleEvent: handleWheel,
					}}>
					<Index each={tabs}>
						{(tab, index) => {
							const handleCloseTab = (e: MouseEvent) => {
								e.stopPropagation();
								$tabs.removeTab(tab().id);
							};

							const handleTabClick = () => {
								$tabs.activateTab(tab().id);
							};

							return (
								<ContextMenu>
									<ContextMenu.Trigger
										data-tab-id={tab().id}
										role="button"
										onClick={handleTabClick}
										class={`app-tab ${tab().isActive ? "app-tab-active" : ""}`}
										title={tab().name}>
										<span class="app-tab-name">{truncateFromEnd(tab().name, 40)}</span>
										<span title="Закрыть" role="button" class="app-tab-close" onClick={handleCloseTab}>
											<IoClose />
										</span>
									</ContextMenu.Trigger>
									<ContextMenu.Portal>
										<ContextMenu.Content class="context-menu">
											<ContextMenu.Item class="context-menu-item" onSelect={() => $tabs.removeTab(tab().id)}>
												Закрыть
											</ContextMenu.Item>
											<ContextMenu.Item
												class="context-menu-item"
												onSelect={() => $tabs.closeOtherTabs(tab().id)}>
												Закрыть остальные
											</ContextMenu.Item>
											<ContextMenu.Item class="context-menu-item" onSelect={() => $tabs.closeAllTabs()}>
												Закрыть все
											</ContextMenu.Item>
										</ContextMenu.Content>
									</ContextMenu.Portal>
								</ContextMenu>
							);
						}}
					</Index>
				</div>

				<div class="app-tab-component scrollbar">
					<Show when={activeTab()!.type === TabType.REQUEST}>
						<SendRequest tabId={activeTab()!.id} />
					</Show>
				</div>
			</Show>

			<Show when={isLoaded() && !activeTab()}>
				<div class="flex h-full w-full items-center justify-center text-5xl text-base-content/20">
					<OcFilecode2 />
				</div>
			</Show>

			<Show when={!isLoaded()}>
				<div class="flex h-full w-full items-center justify-center text-base-content/40">
					<span class="loading loading-spinner loading-lg"></span>
				</div>
			</Show>
		</div>
	);
};
