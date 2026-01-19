import { createRoot, createSignal, JSX } from "solid-js";

export enum TabComponent {
	REQUEST,
}

export type Tab = {
	id: string;
	name: string;
	component: TabComponent;
	componentProps: Record<string, unknown>;
	isActive: boolean;
	temporary: boolean;
};

const createTabsStore = () => {
	const [tabs, setTabs] = createSignal<Tab[]>([]);

	const addTab = (tab: Tab) => {
		setTabs(prev => {
			const hasSameTab = prev.find(t => t.id === tab.id);
			
			if (tab.isActive){
				prev = prev.map(t => ({ ...t, isActive: false }));
			}

			if (hasSameTab) {
				return prev.map(t => (t.id === tab.id ? { ...t, ...tab, isActive: true } : t));
			}

			return [...prev, tab];
		});
	};

	const removeTab = (id: string) => {
		const isActive = tabs().find(tab => tab.id === id)?.isActive;
		setTabs(prev => prev.filter(tab => tab.id !== id));
		if (isActive) {
			const lastTab = tabs()[tabs().length - 1];
			lastTab && activateTab(lastTab.id);
		}
	};

	const updateTab = (id: string, tab: Partial<Tab>) => {
		setTabs(prev => prev.map(t => (t.id === id ? { ...t, ...tab } : t)));
	};

	const activateTab = (id: string) => {
		setTabs(prev => prev.map(t => ({ ...t, isActive: t.id === id })));
	};

	return {
		tabs,
		addTab,
		removeTab,
		updateTab,
		activateTab,
	};
};

export const $tabs = createRoot(createTabsStore);
