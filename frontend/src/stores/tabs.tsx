import { createRoot, createSignal, JSX } from "solid-js";

export type Tab = {
	id: string;
	name: string;
	content: JSX.Element;
	isActive: boolean;
	temporary: boolean;
};

const createTabsStore = () => {
	const [tabs, setTabs] = createSignal<Tab[]>([]);

	const addTab = (tab: Tab) => {
		setTabs(prev => {
			const hasSameTab = prev.find(t => t.id === tab.id);
			if (hasSameTab) {
				return prev.map(t => (t.id === tab.id ? { ...t, ...tab, isActive: true } : t));
			}
			return [...prev, tab];
		});
	};

	const removeTab = (id: string) => {
		setTabs(prev => prev.filter(tab => tab.id !== id));
	};

	const updateTab = (id: string, tab: Partial<Tab>) => {
		setTabs(prev => prev.map(t => (t.id === id ? { ...t, ...tab } : t)));
	};

	return {
		tabs,
		addTab,
		removeTab,
		updateTab,
	};
};

export const $tabs = createRoot(createTabsStore);
