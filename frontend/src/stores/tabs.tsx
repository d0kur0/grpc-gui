import { createRoot, createSignal, JSX } from "solid-js";

export type Tab = {
	id: string;
	name: string;
	content: JSX.Element;
	temporary: boolean;
};

const createTabsStore = () => {
	const [tabs, setTabs] = createSignal<Tab[]>([]);

	const addTab = (tab: Omit<Tab, "id">) => {
		setTabs(prev => [...prev, { ...tab, id: crypto.randomUUID() }]);
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
